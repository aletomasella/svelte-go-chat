package common

import (
	"fmt"
	"net"
	"time"

	"github.com/aletomasella/svelte-go-chat/pkg/domain"
	"golang.org/x/exp/maps"
)

const (
	SafeMode               = false
	ClientDisconnected     = "Client %s Disconnected\n"
	BufferSize             = 128
	ClientConnected        = "Client %s Connected\n"
	ErrorReadingConnection = "ERROR: Trying to read connection, but failed because of %s\n"
	ErrorInvalidMessage    = "ERROR: Invalid Message Provided\n"
	MessageAdded           = "Message added: %s. From user %s\n"
	MessageReceived        = "Message received: %s\n"
	DefaultCase            = "Unknown message type: %s\n"
)

func safeAdress(addr net.Conn) string {
	if SafeMode {
		return "[SAFEMODE]"
	}
	return addr.RemoteAddr().String()
}

func handleMessages(msg domain.Message, clients map[string]*domain.Client) {

	commandsKeys := maps.Keys(Commands)

	address := msg.Conn.RemoteAddr().String()
	switch msg.Type {
	case domain.ClientConnected:
		clients[address] = &domain.Client{
			Conn:        msg.Conn,
			LastMessage: time.Now(),
		}
		// send commnads available
		msg.Conn.Write([]byte(fmt.Sprintf("Commands available: %v\n", commandsKeys)))

		fmt.Printf(ClientConnected, safeAdress(msg.Conn))

	case domain.ClientDisconnected:
		delete(clients, address)
		msg.Conn.Close()
		fmt.Printf(ClientDisconnected, safeAdress(msg.Conn))

	case domain.MessageReceived:
		fmt.Printf(MessageReceived, msg.Body)
		// Send message to all clients except the one who sent it
		for _, client := range clients {
			if client.Conn.RemoteAddr().String() != address {
				if client.Conn != nil {
					client.Conn.Write([]byte(msg.Body))

				}
			}
		}

	case domain.DisconnectRequest:
		fmt.Printf(ClientDisconnected, safeAdress(msg.Conn))
		msg.Conn.Close()
	default:
		fmt.Printf(DefaultCase, msg.Body)
		client := clients[address]
		if client != nil {
			client.Conn.Write([]byte(ErrorInvalidMessage))
			fmt.Print(ErrorInvalidMessage)
			client.Conn.Close()
		}
	}
}

func ChannelServer(msgChannel chan domain.Message) {
	clients := make(map[string]*domain.Client)

	// Infinite loop
	for {
		msg := <-msgChannel
		handleMessages(msg, clients)
	}

}
