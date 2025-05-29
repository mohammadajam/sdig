package server

import (
	"log"
	"net"
	"strings"
)

// Message is a message by a chat or the chat manager to a client.
type Message struct {
	// The string is the type of the message.
	// The types are for now ("n" for "notify" and "e" for "error")
	string
	content string 	// the content of the message.
}

// Create a new Message.
func NewMessage(message string, content string) Message {
	return Message{
		string: message,
		content: content,
	}
}


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
	for {
		var buffer []byte = make([]byte, 1240)
		_, err := u.conn.Read(buffer)
		if err != nil {
			log.Println("ERROR: Failed to read from user:", err)
		}
		message := strings.Fields(strings.TrimSpace(string(buffer)))

		if u.connected == false {
			if message[0] == "lo" {
				if len(message)-1 != 3 {
					u.messages <- NewMessage("e", "Error: Unknown username or password")
				}
				username, password := message[1], message[2]
						u.managerChan <- NewLoginRequest(username, password, u)
			}
		} else {
			switch message[0] {
			// TODO: handle requests after the client has logged in.
			case "nm":
				log.Println("Got Message")
				chatId, content := message[1] ,strings.Join(message[2:], " ")
				u.chats[chatId] <- NewClientRequest("nm", content, u)
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
