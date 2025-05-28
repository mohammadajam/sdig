package server

import (
	"log"
	"net"
	"strings"
)

// This type contains information about the client.
// The information is the following:
//	username: a unique name to each user.
// 	name: a nickname of sort, it doesn't have to be unique.
// 	conn: the socket.
// 	chats: a map of strings that represents a unique id to a channel of the chat of that id.
// 	managerChan: the channel of the chat manager.
// 	messages: a channel of messages to be sent to the client.
// 	connected: a bool that represents whether a client has logged in to a user.
type User struct {
	username string						// a unique name to each user
	name string							// a nickname of sort, it doesn't have to be unique.
	conn net.Conn						// the socket of the client.
	chats map[string]chan ClientRequest	// a map of strings that represents a unique id to a channel of the chat of that id.
	managerChan chan ClientRequest		// the chanel of the chat manager.
	messages chan Message				// a chanel of messages to be sent to the client.
	connected bool						// a bool that represents whether a client has logged in to a user.
}

// Initializes a new user that isn't logged in to any account.
func NewUser(conn net.Conn, managerChan chan ClientRequest) User {
	return User {
		conn: conn,
		managerChan: managerChan,
		chats: make(map[string]chan ClientRequest),
		messages: make(chan Message),
		connected: false,
	}
}

// Handles and procceses requests sent by the user throgh the socket and sends the proccesed request to a chat or to the chat manager.
func (u *User) HandleUserRequest() {
	var buffer []byte = make([]byte, 1240)

	for {
		n, err := u.conn.Read(buffer)
		if err != nil {
			log.Println("ERROR: Failed to read from user:", err)
		}
		message := string(buffer[:n])

		if u.connected == false {
			if buffer[0] == 'l' {
				username, password, found := strings.Cut(
					strings.TrimPrefix(message, "l "),
					" ");
				if found == false {
					u.messages <- NewMessage("e", "Error: unable to find username and password")
				} else {
					if strings.ContainsAny(username, " ") || 
						strings.ContainsAny(password, " ") {
						u.messages <- NewMessage("e", "Error: username or password contains a space")
					} else {
						u.managerChan <- NewLoginRequest(username, password, u)
					}
				}
			}
		} else {
			switch buffer[0] {
			// TODO: handle requests after the client has logged in.
			// example "m" for message, which means that the message should be sent to the chat specified in the message.
			}
		}
	}
}

// Handles messages from the chat manager or from other chats.
// NOTE: for now the function just send the messages to the client, though this may change in the future.
func (u *User) HandleMessagesToUser() {
	for {
		mes := <- u.messages
		mesType := mes.string
		content := mes.content
		message := mesType + " " +  content
		u.conn.Write([]byte(message))
	}
}
