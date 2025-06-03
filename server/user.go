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
		var buffer []byte = make([]byte, 1024)
		n, err := u.conn.Read(buffer)
		if err != nil {
			log.Println("ERROR: Failed to read from user:", err)
		}
		message := strings.Fields(strings.TrimSpace(string(buffer[0:n])))
		argCount := len(message)-1
		if argCount == -1 {
			continue;
		}

		if u.connected == false {
			switch message[0] {
			case LoginRequestType:
				if argCount != 2 {
					u.messages <- NewMessage("e", "Error: Unknown username or password")
					continue
				}
				username, password := message[1], message[2]
				u.managerChan <- LoginRequest(username, password, u)

			case NewUserRequestType:
				if argCount < 3 {
					u.messages <- NewMessage("e", "Error: User data format Error")
					continue
				}
				u.managerChan <- NewUserRequest(message[1], strings.Join(message[2:argCount], " "), message[argCount], u)

			case QuitRequestType:
				if argCount != 0 {
					u.messages <- NewMessage("e", "Error: User data format Error")
					continue
				}
				for _, chat := range u.chats {
					chat <- QuitRequest(u)
				}
				u.messages <- NewMessage("q", "quitting")
				u.conn.Close()
				return
			}
		} else {
			switch message[0] {
			case LogoutRequestType:
				if argCount != 0 {
					u.messages <- NewMessage("e", "Error: User data format Error")
					continue
				}
				for _, chat := range u.chats {
					chat <- LogoutRequest(u)
				}
				u.connected = false
				u.chats =  make(map[string]chan ClientRequest)
				u.messages <- NewMessage("n", "logged out")

			case DeleteUserRequestType:
				if argCount != 1 {
					u.messages <- NewMessage("e", "Error: User data format Error")
					continue
				}
				u.managerChan <- DeleteUserRequest(message[1], u)

			case JoinChatRequestType:
				if argCount != 2 {
					u.messages <- NewMessage("e", "Error: User data format Error")
					continue
				}
				u.managerChan <- JoinChatRequest(message[1], message[2], u)

			case LeaveChatRequestType:
				if argCount != 1 {
					u.messages <- NewMessage("e", "Error: User data format Error")
					continue
				}
				u.managerChan <- LeaveChatRequest(message[1], u)

			case NewChatRequestType:
				if argCount < 3 {
					u.messages <- NewMessage("e", "Error: User data format Error")
					continue
				}
				u.managerChan <- NewChatRequest(message[1], strings.Join(message[2:argCount], " "), message[argCount], u)

			case DeleteChatRequestType:
				if argCount != 2 {
					u.messages <- NewMessage("e", "Error: User data format Error")
					continue
				}
				u.managerChan <- DeleteChatRequest(message[1], message[2], u)

			case QuitRequestType:
				if argCount != 0 {
					u.messages <- NewMessage("e", "Error: User data format Error")
					continue
				}
				for _, chat := range u.chats {
					chat <- QuitRequest(u)
				}
				u.messages <- NewMessage("q", "quitting")
				u.conn.Close()
				return

			case NewMessageRequestType:
				if argCount < 2 {
					u.messages <- NewMessage("e", "Error: Message is empty or chat id is missing")
					continue
				}
				chatId, content := message[1] ,strings.Join(message[2:], " ")
				u.chats[chatId] <- NewMessageRequest(content, u)

			case DeleteMessageRequestType:
				if argCount < 2 {
					u.messages <- NewMessage("e", "Error: Message ID is not present or chat id is missing")
					continue
				}
				u.chats[message[1]] <- DeleteMessageRequest(message[2], u)

			case GetMessagesRequestType:
				if argCount < 3 {
					u.messages <- NewMessage("e", "Error: Message IDs are not present empty or chat id is missing")
					continue
				}
				u.chats[message[1]] <- GetMessagesRequest(message[2], message[3], u)

			case GetUsersRequestType:
				if argCount < 1 {
					u.messages <- NewMessage("e", "Error:Chat ID is missing")
					continue
				}
				u.chats[message[1]] <- GetUsersRequest(u)
			}
		}
	}
}

// Handles messages from the chat manager or from other chats.
func (u *User) HandleMessagesToUser() {
	for {
		mes := <- u.messages
		mesType := mes.string
		content := mes.content
		if mesType == "q" {
			break
		}
		message := mesType + " " +  content
		u.conn.Write([]byte(message))
	}
}
