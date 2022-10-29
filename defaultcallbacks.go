package mitm

import (
	"fmt"
	"log"
	"net"
)

func (px *Proxy) SetupDefaultCallbacks() {
	if px.OnError == nil {
		px.OnError = DefaultCallback_onError
	}
	if px.OnConnect == nil {
		px.OnConnect = DefaultCallback_onConnect
	}
	if px.OnReceive == nil {
		px.OnReceive = DefaultCallback_onReceive
	}
	if px.OnClose == nil {
		px.OnClose = DefaultCallback_onClose
	}
}

func DefaultCallback_onError(connectionid uint64, dir Direction, err error) {
	log.Printf("Error:  connection %d %s: %v", connectionid, dir, err)
}

func DefaultCallback_onConnect(connectionid uint64, conn net.Conn) bool {
	remote_address := conn.RemoteAddr().String()
	log.Printf("Accept: connection %d from %s", connectionid, remote_address)
	return true
}

func DefaultCallback_onClose(connectionid uint64, dir Direction) {
	log.Printf("Closed: connection %d from below", connectionid)
}

func DefaultCallback_onReceive(connectionid uint64, dir Direction, buf *[]byte, length int) bool {
	return true
}

func DefaultCallback_PrintReceive(connectionid uint64, dir Direction, buf *[]byte, length int) bool {
	fmt.Printf("Payload: %d %s: %#v\n", connectionid, dir, string((*buf)[:length]))
	return true
}
