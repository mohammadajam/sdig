package server

import (
	"database/sql"
	"log"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

// This type contains the information about a chat.
// the information is the following:
//	chatId: a unique name for each chat.
//	chatName: the public name of the chat that is displayed.
//	chatChan: the channel that receives the requests from users that are logged in to the chat.
// 	users: a map of where the key the username and the value is a pointer to the user. note that the users in this map are not all users added to the chat in the database but only the connected to the chat.
//	mu: a pointer to a shared mutex.
type Chat struct {
	chatId string				// a unique name for each chat.
	chatName string 			// the public name of the chat that is displayed.
	chatChan chan ClientRequest	// the channel that receives the requests from users that are logged in to the chat.
	users map[string]*User		// a map of where the key the username and the value is a pointer to the user. Note that the users in this map are not all users added to the chat but only the connected to the chat.
	mu *sync.RWMutex			// a pointer to a shared mutex.
}

// Loads chats from the database and putting them in map where the key is the chat id and the value is a chat object.
func LoadChats(mu *sync.RWMutex) map[string]Chat {
	chats := make(map[string]Chat)

	db, err := sql.Open("sqlite3", "sdig.db")
	if err != nil {
		log.Fatalln("ERROR: COULD NOT LOAD CHATS:", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT chatId, chatName FROM chats;")
	if err != nil {
		log.Fatalln("ERROR: COULD NOT QUERY CHATS:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var chatId, chatName string

		err := rows.Scan(&chatId, &chatName)
		if err != nil {
			log.Fatalln("ERROR: COULD NOT READ ROW:", err)
		}
		
		chats[strings.TrimSpace(chatId)] = NewChat(strings.TrimSpace(chatId), strings.TrimSpace(chatName), mu)
	}
	
	err = rows.Err()
	if err != nil {
		log.Fatalln("ERROR: AN ERROR OCCURED DURING READING ROWS:", err)
	}

	return chats
}

// Creates a chat object from the input.
func NewChat(chatId string, chatName string, mu *sync.RWMutex) Chat {
	return Chat {
		chatId: chatId,
		chatName: chatName,
		chatChan: make(chan ClientRequest),
		users: make(map[string]*User),
		mu: mu,
	}
}

// Handles requests from users connected to the chat.
func (chat *Chat) HandleChat() {
	for {
		req := <- chat.chatChan
		switch (req.string) {

		}
	}
}


// ClientRequest is requests by the client to a chat or to  the chat manager.
type ClientRequest struct {
	// the type of the request.
	//	"lo": "login"
	//	"nu": "new user"
	//	"du": "delete user"
	//	"jo": "join chat"
	//	"le": "leave chat"
	//	"nc": "new chat"
	//	"dc": "delete chat"
	string
	content string	// the content of the request.
	sender *User	// a pointer to the user who sent the request.
}


func NewClientRequest(request string, data string, user *User) ClientRequest {
	return ClientRequest{
		string: request,
		content: data,
		sender: user,
	}
}

func NewLoginRequest(username string, password string, user *User) ClientRequest {
	req_content := strings.TrimSpace(username) + " " + strings.TrimSpace(password)
	return NewClientRequest("lo", req_content, user)
}

// TODO: add functions for all request types specified in the ClientRequest description.


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


// The chat manager is resposible for many things
// They are:
//	1. Load chats.
//	2. Add new chats.
//	3. Delete chats.
//	4. Log users in.
//	5. Create new users.
//	6. Delete users.
//	7. Add users to a chat.
//	8. Remove users from a chat(leaving or banning).
// The chat manager stores the following:
//	chats: a map of chat ids to chats.
//	ManagerChan: the channel through the client sends requests.
//	mu: a pointer to a shared mutex.
type ChatManager struct {
	chats map[string]Chat			// a map of chat ids to chats. should be loaded through LoadChats function.
	ManagerChan chan ClientRequest	// the channel through the client sends requests.
	mu *sync.RWMutex				// a pointer to a shared mutex.
}

// Creates a chat manager. uses LoadChats functions.
func NewChatManager(mu *sync.RWMutex) ChatManager {
	return ChatManager{
		chats: LoadChats(mu),
		ManagerChan: make(chan ClientRequest),
		mu: mu,
	}
}

// Handles user requests.
func (cm *ChatManager) HandleRequests() {
	db, err := sql.Open("sqlite3", "sdig.db")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO OPEN DATABASE:", err)
	}
	defer db.Close()


	get_chats, err := db.Prepare("SELECT chatId FROM logged_in WHERE user=?")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}
	defer get_chats.Close()

	get_user, err := db.Prepare("SELECT name, password FROM users WHERE username=?")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}
	defer get_user.Close()

	// TODO: add other sql statements such as (add_to_chat, remove_from_chat, new_chat, delete_chat)
	// NOTE: should try adding other ones such as (rename_chat, change_password)

	for {
		req := <- cm.ManagerChan

		log.Println("Got Request")

		switch req.string {
		case "lo":
			cm.mu.RLock()
			username, sentPassword, _ := strings.Cut(req.content, " ")

			var password string
			var nickname string

			username = strings.TrimSpace(username)
			err := get_user.QueryRow(username).Scan(&nickname, &password)
			if err == sql.ErrNoRows {
				req.sender.messages <- NewMessage("e", "NoSuchUser")
			} else if err != nil {
				log.Println("Error: Could not search for user", err)
			}

			sentPassword = strings.TrimSpace(sentPassword)
			if password == sentPassword {
				rows, err := get_chats.Query(strings.TrimSpace(username))
				if err != nil {
					log.Println("Error: Unable to query logged_in table:", err)
				}

				for rows.Next() {
					var chatId string

					err := rows.Scan(&chatId)
					if err != nil {
						log.Println("Error: Unable to scan row:", err)
					}
					req.sender.chats[chatId] = cm.chats[chatId].chatChan
					cm.chats[chatId].users[username] = req.sender
				}
				req.sender.name = nickname
				req.sender.username = username
				req.sender.connected = true
				req.sender.messages <- NewMessage("n", "connected")
			}
			cm.mu.RUnlock()
		}
	}
}
