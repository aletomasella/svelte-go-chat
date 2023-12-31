package common

import (
	"net"
	"strings"

	"github.com/aletomasella/svelte-go-chat/pkg/domain"
)

var (
	Commands = make(map[string]int)
)

func Client(conn net.Conn, messages chan domain.Message) {
	buffer := make([]byte, BufferSize)
	Commands[":quit"] = int(domain.DisconnectRequest)
	Commands[":users"] = int(domain.GetUsers)
	// Commands[":username"] = int(domain.SetUsername)

	for {
		n, err := conn.Read(buffer)

		if err != nil {
			conn.Close()
			messages <- domain.Message{
				Type: domain.ClientDisconnected,
				Conn: conn,
				Body: "",
			}
			return
		}

		msg := string(buffer[:n])

		if n > 0 && len(msg) == n {
			if strings.Contains(msg, ":username") {
				messages <- domain.Message{
					Type: domain.SetUsername,
					Conn: conn,
					Body: msg,
				}
				return
			}

			// need to trim the message because it comes with spaces
			val, ok := Commands[strings.TrimSpace(msg)]

			if ok {
				messages <- domain.Message{
					Type: domain.MessageType(val),
					Conn: conn,
					Body: msg,
				}
				return
			}

			messages <- domain.Message{
				Type: domain.MessageReceived,
				Body: msg,
				Conn: conn,
			}
		}
	}

}
