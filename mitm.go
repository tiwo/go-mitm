package mitm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
)

const _BUFSIZE = 24 * 8192

var (
	ErrMultipleServe = errors.New("Proxy.Serve() called more than once!")
	ErrConnectReject = errors.New("Connection rejected by OnConnect callback")
)

type Direction int8

const (
	DIR_NONE Direction = iota
	DIR_UP
	DIR_DOWN
)

func (dir Direction) String() string {
	switch dir {
	case DIR_NONE:
		return "[NONE]"
	case DIR_UP:
		return "[ up ]"
	case DIR_DOWN:
		return "[down]"
	default:
		return "[????]"
	}
}

type Proxy struct {
	Network        string
	ListenAddress  *net.TCPAddr
	ForwardAddress *net.TCPAddr

	OnError   func(connectionid uint64, dir Direction, err error)
	OnConnect func(connectionid uint64, conn net.Conn) bool
	OnClose   func(connectionid uint64, dir Direction)
	OnReceive func(connectionid uint64, dir Direction, buf *[]byte, length int) bool

	Ctx context.Context

	lastConnectionId uint64
}

func New(Network string, ListenAddress string, ForwardAddress string) (*Proxy, error) {

	listen_address, err := net.ResolveTCPAddr(Network, ListenAddress)
	if err != nil {
		return nil, fmt.Errorf("while resolving ListenAddress %#v: %w", ListenAddress, err)
	}

	forward_address, err := net.ResolveTCPAddr(Network, ForwardAddress)
	if err != nil {
		return nil, fmt.Errorf("while resolving ForwardAddress %#v: %w", ForwardAddress, err)
	}

	px := Proxy{}
	px.Network = Network
	px.ListenAddress = listen_address
	px.ForwardAddress = forward_address

	return &px, nil
}

func (px *Proxy) Serve() error {
	var err error

	if px.lastConnectionId > 0 {
		return ErrMultipleServe
	}

	if px.Network == "" {
		px.Network = "tcp"
	}

	if px.Ctx == nil {
		px.Ctx = context.Background()
	}

	listener, err := net.ListenTCP(px.Network, px.ListenAddress)
	if err != nil {
		return fmt.Errorf("Could not listen: %w", err)
	}
	defer listener.Close()

	for {
		px.lastConnectionId += 1

		fmt.Printf("inc: %d\n", px.lastConnectionId)

		conn, err := listener.Accept()
		if err != nil {
			px.OnError(px.lastConnectionId, DIR_NONE, fmt.Errorf("while accepting connection %d: %w", px.lastConnectionId, err))
			continue
		}

		go px.serveConnection(conn, px.lastConnectionId)
	}

}

func (px *Proxy) serveConnection(conn net.Conn, connid uint64) error {

	fmt.Printf("%d: %v\n", connid, conn)

	if !px.OnConnect(connid, conn) {
		err := conn.Close()
		px.OnError(connid, DIR_NONE, ErrConnectReject)
		if err != nil {
			px.OnError(connid, DIR_NONE, fmt.Errorf("while closing connection %d (rejected by OnConnect callback): %w", connid, err))
		}
		return fmt.Errorf("rejecting connection %d: %w", connid, ErrConnectReject)
	}

	upconn, err := net.DialTCP(px.Network, nil, px.ForwardAddress)
	if err != nil {
		err = fmt.Errorf("while connecting to forward address: %w", err)
		px.OnError(connid, DIR_UP, err)
		return err
	}

	ctx, cancel := context.WithCancel(px.Ctx)
	defer cancel()

	go px.forward(ctx, cancel, connid, DIR_UP, conn, upconn)
	go px.forward(ctx, cancel, connid, DIR_DOWN, upconn, conn)

	<-ctx.Done()

	conn.Close()
	upconn.Close()

	return nil
}

func (px *Proxy) forward(ctx context.Context, cancel context.CancelFunc, connid uint64, dir Direction, from io.ReadWriteCloser, to io.ReadWriteCloser) {
	defer cancel()

	buf := make([]byte, _BUFSIZE)
	for {
		n, err := from.Read(buf)
		if err != nil {
			px.OnError(connid, dir, fmt.Errorf("while reading: %w", err))
			break
		}

		if px.OnReceive(connid, dir, &buf, n) {
			_, err = to.Write(buf[:n])
			if err != nil {
				px.OnError(connid, dir, fmt.Errorf("while writing: %w", err))
				break
			}
		}
	}

	px.OnError(connid, dir, fmt.Errorf("Closing!"))
}
