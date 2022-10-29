package main

import (
	"flag"
	"log"

	"github.com/tiwo/go-mitm/v2"
)

func main() {
	network := "tcp"
	//network := flag.String("tcp-network", "tcp", "TCP network")
	listenAddress := flag.String("local", "localhost:9997", "listen address")
	forwardAddress := flag.String("remote", "", "remote Address")
	flag.Parse()

	px, err := mitm.New(network, *listenAddress, *forwardAddress)
	if err != nil {
		log.Fatal()
	}

	px.SetupDefaultCallbacks()
	err = px.Serve()
	if err != nil {
		log.Fatalf("Errror from Server(): %v", err)
	}
}
