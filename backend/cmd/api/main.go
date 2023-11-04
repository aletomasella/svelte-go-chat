package main

import (
	"fmt"
	"net"
	"os"

	"github.com/aletomasella/svelte-go-chat/pkg/common"
	"github.com/aletomasella/svelte-go-chat/pkg/domain"
)

// Go All Format Types : https://pkg.go.dev/fmt

const (
	Port                  = "3000"
	Protocol              = "tcp"
	ErrorInitializingPort = "ERROR: Trying to listen in port %s, but failed because of %s\n"
	ErrorAcceptConnection = "ERROR: Trying to accept connection, but failed because of %s\n"
	Listening             = "Listening in port %s\n"
)

func main() {
	ln, err := net.Listen(Protocol, ":"+Port)

	if err != nil {
		fmt.Printf(ErrorInitializingPort, Port, err)
		os.Exit(1)
	}

	fmt.Printf(Listening, Port)

	msgChannel := make(chan domain.Message)

	//Inicialize the server that will listen to the messages
	// Go Routine means that the function will run in parallel
	go common.ChannelServer(msgChannel)

	for {
		conn, err := ln.Accept()

		if err != nil {
			fmt.Printf(ErrorAcceptConnection, err)
			continue
		}

		msgChannel <- domain.Message{
			Type: domain.ClientConnected,
			Conn: conn,
			Body: "",
		}

		//Inicialize the client that will listen to the messages
		go common.Client(conn, msgChannel)

	}
}
