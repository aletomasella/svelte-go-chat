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
	RateLimiter            = 2
	RateLimiterMessage     = "Your are sending messages too fast. Strike Count : %v\n"
	MaxStrikeCount         = 5
	MaxStrikeCountReach    = "You reached the max strike count. Disconnecting...\n"
	BannedTime             = 30
)

func safeAdress(addr net.Conn) string {
	if SafeMode {
		return "[SAFEMODE]"
	}
	return addr.RemoteAddr().String()
}

func secondsUntilUnban(banTime time.Time) int {
	return int(time.Until(banTime.Add(BannedTime * time.Second)).Seconds())
}

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

func handleMessages(msg domain.Message, clients map[string]*domain.Client, bannedClients map[string]*domain.Client) {

	commandsKeys := maps.Keys(Commands)
	address := msg.Conn.RemoteAddr().String()
	addressWithoutPort := address[:len(address)-6]
	bannedClient, ok := bannedClients[addressWithoutPort]

	if ok {
		// compare time to check if it is still banned
		if inTimeSpan(bannedClient.BanTime, bannedClient.BanTime.Add(BannedTime*time.Second), time.Now()) {
			msg.Conn.Write([]byte(fmt.Sprintf("You are banned for %v seconds\n", secondsUntilUnban(bannedClient.BanTime))))
			msg.Conn.Close()
			return
		}

		delete(bannedClients, address)
	}

	switch msg.Type {
	case domain.ClientConnected:
		clients[address] = &domain.Client{
			Conn:        msg.Conn,
			LastMessage: time.Now(),
			StrikeCount: 0,
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

		// rate limit messages
		if time.Since(clients[address].LastMessage) < RateLimiter*time.Second {
			msg.Conn.Write([]byte(fmt.Sprintf(RateLimiterMessage, clients[address].StrikeCount)))
			clients[address].StrikeCount++

			if clients[address].StrikeCount > MaxStrikeCount {
				msg.Conn.Write([]byte(MaxStrikeCountReach))
				msg.Conn.Close()
				fmt.Printf(ClientDisconnected, safeAdress(msg.Conn))
				delete(clients, address)
				bannedClients[addressWithoutPort] = &domain.Client{
					Conn:        msg.Conn,
					BanTime:     time.Now(),
					LastMessage: time.Now(),
					StrikeCount: MaxStrikeCount,
				}
			}
			return
		}

		for _, client := range clients {
			if client.Conn.RemoteAddr().String() != address {
				if client.Conn != nil {
					client.Conn.Write([]byte(msg.Body))
				}
			}
			// update last message time
			clients[address].LastMessage = time.Now()
			clients[address].StrikeCount = 0
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
	bannedClients := make(map[string]*domain.Client)

	// Infinite loop
	for {
		msg := <-msgChannel
		handleMessages(msg, clients, bannedClients)
	}

}
