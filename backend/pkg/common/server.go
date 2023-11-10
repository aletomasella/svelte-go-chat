package common

import (
	"fmt"
	"net"
	"strings"
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
		hashedIP, err := AdressHash(msg.Conn.RemoteAddr().String())

		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			hashedIP = "ERROR"
		}

		clients[address] = &domain.Client{
			Conn:        msg.Conn,
			LastMessage: time.Now(),
			StrikeCount: 0,
			UserName:    "",
			HashedIP:    hashedIP,
		}
		// send commnads available
		msg.Conn.Write([]byte(fmt.Sprintf("Commands available: %v\n", commandsKeys)))

		fmt.Printf(ClientConnected, safeAdress(msg.Conn))

	case domain.ClientDisconnected:
		disconnectedClient := clients[address]
		msg.Conn.Close()
		fmt.Printf(ClientDisconnected, safeAdress(msg.Conn))
		delete(clients, address)
		// Send message to all clients that a user disconnected
		for _, client := range clients {
			if client.Conn.RemoteAddr().String() != address {
				if client.Conn != nil {
					client.Conn.Write([]byte(fmt.Sprintf("Client %s Disconnected\n", disconnectedClient.UserName)))
				}
			}
		}

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
					UserName:    "",
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
		disconnectedClient := clients[address]
		delete(clients, address)
		// Send message to all clients that a user disconnected
		for _, client := range clients {
			if client.Conn.RemoteAddr().String() != address {
				if client.Conn != nil {
					client.Conn.Write([]byte(fmt.Sprintf("Client %s Disconnected\n", disconnectedClient.UserName)))
				}
			}
		}

	case domain.SetUsername:
		clients[address].UserName = strings.TrimSpace(strings.Replace(msg.Body, "/username ", "", 1))
		fmt.Printf("Client %s set username to %s\n", safeAdress(msg.Conn), clients[address].UserName)
		msg.Conn.Write([]byte(fmt.Sprintf("Username set to %s\n", clients[address].UserName)))

	case domain.GetUsers:
		fmt.Printf("Client %s requested users\n", safeAdress(msg.Conn))
		usersHashedIPs := make([]string, 0)

		for _, client := range clients {
			if client.Conn.RemoteAddr().String() != address {
				if client.Conn != nil {
					usersHashedIPs = append(usersHashedIPs, client.HashedIP)
				}
			}
		}

		msg.Conn.Write([]byte(fmt.Sprintf("Users: %v\n", usersHashedIPs)))

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
