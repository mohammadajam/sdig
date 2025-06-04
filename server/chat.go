package server

import (
	"database/sql"
	"log"
	"strings"
	"sync"

	"github.com/mattn/go-sqlite3"
)

const DatabasePath string = "sdig.db"

// This type contains the information about a chat.
// the information is the following:
//	chatId: a unique name for each chat.
//	chatName: the public name of the chat that is displayed.
//	chatChan: the channel that receives the requests from users that are logged in to the chat.
// 	owner: a string of the username of the owner of the chat.
// 	users: a map of where the key the username and the value is a pointer to the user. note that the users in this map are not all users added to the chat in the database but only the connected to the chat.
//	mu: a pointer to a shared mutex.
type Chat struct {
	chatId string				// a unique name for each chat.
	chatName string 			// the public name of the chat that is displayed.
	chatChan chan ClientRequest	// the channel that receives the requests from users that are logged in to the chat.
	owner string				// the username of the owner of the chat.
	users map[string]*User		// a map of where the key the username and the value is a pointer to the user. Note that the users in this map are not all users added to the chat but only the connected to the chat.
	mu *sync.RWMutex			// a pointer to a shared mutex.
}

// Loads chats from the database and putting them in map where the key is the chat id and the value is a chat object.
func LoadChats(mu *sync.RWMutex) map[string]Chat {
	chats := make(map[string]Chat)

	db, err := sql.Open("sqlite3", DatabasePath)
	if err != nil {
		log.Fatalln("ERROR: COULD NOT LOAD CHATS:", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT chatId, chatName, owner FROM chats;")
	if err != nil {
		log.Fatalln("ERROR: COULD NOT QUERY CHATS:", err)
	}
	defer rows.Close()

	for rows.Next() {
		var chatId, chatName, owner string

		err := rows.Scan(&chatId, &chatName, &owner)
		if err != nil {
			log.Fatalln("ERROR: COULD NOT READ ROW:", err)
		}
		
		chats[strings.TrimSpace(chatId)] = NewChat(chatId, chatName, owner, mu)
	}
	
	err = rows.Err()
	if err != nil {
		log.Fatalln("ERROR: AN ERROR OCCURED DURING READING ROWS:", err)
	}

	return chats
}

// Creates a chat object from the input.
func NewChat(chatId string, chatName string, owner string, mu *sync.RWMutex) Chat {
	return Chat {
		chatId: chatId,
		chatName: chatName,
		chatChan: make(chan ClientRequest),
		owner: owner,
		users: make(map[string]*User),
		mu: mu,
	}
}

// Handles requests from users connected to the chat.
func (chat *Chat) HandleRequests() {
	db, err := sql.Open("sqlite3", DatabasePath)
	if err != nil {
		log.Panicln("ERROR: COULD OPEN DATABASE:", err)
	}

	insertMessage, err := db.Prepare("INSERT INTO messages (username, chatId, content) VALUES (?, ?, ?)")
	if err != nil {
		log.Panicln("ERROR: COULD PREPARE STATEMENT:", err)
	}
	defer insertMessage.Close()

	getDate, err := db.Prepare("SELECT date FROM messages WHERE id = ?")
	if err != nil {
		log.Panicln("ERROR: COULD PREPARE STATEMENT:", err)
	}
	defer getDate.Close()

	for {
		req := <- chat.chatChan
		switch (req.string) {
		case NewMessageRequestType:
			message := []byte(chat.chatId + " " + req.content)
			chat.mu.Lock()
			_, err = insertMessage.Exec(req.sender.username, chat.chatId, req.content)
			chat.mu.Unlock()
			if err != nil {
				log.Println("Error: Could not insert message", err)
				req.sender.messages <- NewMessage("e", "An error occured")
				continue
			}
			
			for username, user := range chat.users {
				if username != req.sender.username {
					user.conn.Write(message)
				}
			}

		case DeleteChatRequestType:
			for _, user := range chat.users {
				delete(user.chats, chat.chatId)
				user.messages <- NewMessage("n", chat.chatId + " got deleted")
			}
			return

		case QuitRequestType:
			delete(chat.users, req.content)
		}
	}
}


// The server manager is resposible for many things
// They are:
//	1. Load chats.
//	2. Add new chats.
//	3. Delete chats.
//	4. Log users in.
//	5. Create new users.
//	6. Delete users.
//	7. Add users to a chat.
//	8. Remove users from a chat(leaving or banning).
// The server manager stores the following:
//	chats: a map of chat ids to chats.
//	ManagerChan: the channel through the client sends requests.
//	mu: a pointer to a shared mutex.
type ServerManager struct {
	chats map[string]Chat			// a map of chat ids to chats. should be loaded through LoadChats function.
	ManagerChan chan ClientRequest	// the channel through the client sends requests.
	mu *sync.RWMutex				// a pointer to a shared mutex.
}

// Creates a server manager. uses LoadChats functions.
func NewServerManager(mu *sync.RWMutex) ServerManager {
	return ServerManager{
		chats: LoadChats(mu),
		ManagerChan: make(chan ClientRequest, 10),
		mu: mu,
	}
}

// Start the request handling of each chat.
func (cm *ServerManager)StartChatsHandleRequests() {
	for _, chat := range cm.chats {
		go chat.HandleRequests()
	}
}

// Handles user requests.
func (cm *ServerManager) HandleRequests() {
	db, err := sql.Open("sqlite3", "sdig.db")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO OPEN DATABASE:", err)
	}
	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatalln("ERROR: COULD NOT ENABLE foreign_keys:", err)
	}

	getChats, err := db.Prepare("SELECT chatId FROM joined WHERE username=?")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}
	defer getChats.Close()

	getUser, err := db.Prepare("SELECT name, password FROM users WHERE username=?")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}
	defer getUser.Close()

	addUser, err := db.Prepare("INSERT INTO users (username, name, password) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}
	defer addUser.Close()

	deleteUser, err := db.Prepare("DELETE FROM users where username = ? and password = ?")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}
	defer deleteUser.Close()

	getChat, err := db.Prepare("SELECT chatName, password FROM chats WHERE chatId = ?")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}
	defer getChat.Close()

	joinChat, err := db.Prepare("INSERT INTO joined (username, chatId) VALUES (?, ?)")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}
	defer joinChat.Close()

	leaveChat, err := db.Prepare("DELETE FROM joined WHERE username = ? and chatId = ?")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}
	defer leaveChat.Close()

	addChat, err := db.Prepare("INSERT INTO chats (chatId, chatName, password, owner) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}
	defer addChat.Close()

	getOwner, err := db.Prepare("SELECT owner FROM chats WHERE chatId = ?")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}
	defer getOwner.Close()

	deleteChat, err := db.Prepare("DELETE FROM chats WHERE chatId = ? and password = ?")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}
	defer deleteChat.Close()

	// NOTE: should try adding other ones such as (rename_chat, change_password)

	for {
		req := <- cm.ManagerChan

		switch req.string {
		case LoginRequestType:
			username, sentPassword, _ := strings.Cut(req.content, " ")

			var password string
			var nickname string

			username = strings.TrimSpace(username)
			cm.mu.RLock()
			err := getUser.QueryRow(username).Scan(&nickname, &password)
			cm.mu.RUnlock()
			if err == sql.ErrNoRows {
				req.sender.messages <- NewMessage("e", "NoSuchUser")
			} else if err != nil {
				log.Println("Error: Could not search for user", err)
				req.sender.messages <- NewMessage("e", "An error occured")
			}

			sentPassword = strings.TrimSpace(sentPassword)
			if password == sentPassword {
				cm.mu.RLock()
				rows, err := getChats.Query(strings.TrimSpace(username))
				cm.mu.RUnlock()
				if err != nil {
					log.Println("Error: Unable to query logged_in table:", err)
					req.sender.messages <- NewMessage("e", "An error occured")
				}

				for rows.Next() {
					var chatId string

					err := rows.Scan(&chatId)
					if err != nil {
						log.Println("Error: Unable to scan row:", err)
						req.sender.messages <- NewMessage("e", "An error occured")
					}
					req.sender.chats[chatId] = cm.chats[chatId].chatChan
					cm.chats[chatId].users[username] = req.sender
				}
				req.sender.name = nickname
				req.sender.username = username
				req.sender.connected = true
				req.sender.messages <- NewMessage("n", "connected")
			}

		case NewUserRequestType:
			parts:= strings.Split(req.content, " ")
			numParts := len(parts)
			username := parts[0]
			name := strings.Join(parts[1:numParts-1], " ")
			password := parts[numParts-1]

			cm.mu.Lock()
			_, err = addUser.Exec(username, name, password)
			cm.mu.Unlock()
			if err != nil {
				if sqliteErr, ok := err.(sqlite3.Error); ok {
					if sqliteErr.Code == sqlite3.ErrConstraint {
						req.sender.messages <- NewMessage("n", "Username already taken")
						continue
					}
				}
				log.Println("Error: Could not add user to users table", err)
				req.sender.messages <- NewMessage("e", "An error occured")
			}

			req.sender.username = username
			req.sender.name = name
			req.sender.connected = true
			req.sender.messages <- NewMessage("n", "User Created and logged in")

		case DeleteUserRequestType:
			password := req.content
			cm.mu.Lock()
			res, err := deleteUser.Exec(req.sender.username, password)
			cm.mu.Unlock()
			if sqliteErr, ok := err.(sqlite3.Error); ok {
				if sqliteErr.Code == sqlite3.ErrNo(sqlite3.ErrConstraint) {
					req.sender.messages <- NewMessage("n", "You are the owner of at least one chat, delete or transfer ownership of the chats first.")
					continue
				}
			} else if err != nil {
				log.Println("Error: Could not delete user:", err)
				req.sender.messages <- NewMessage("e", "An error occured")
				continue
			}

			affected, err := res.RowsAffected()
			if err != nil {
				log.Println("Error: Could not get affected rows number:", err)
			}

			if affected != 0 {
				req.sender.username = ""
				req.sender.chats = make(map[string]chan ClientRequest)
				req.sender.connected = false
				req.sender.messages <- NewMessage("n", "User deleted")
			}
		
		case JoinChatRequestType:
			chatId, sentChatPassword, _ := strings.Cut(req.content, " ")

			var chatName string
			var chatPassword string

			chatId = strings.TrimSpace(chatId)
			cm.mu.RLock()
			err := getChat.QueryRow(chatId).Scan(&chatName, &chatPassword)
			cm.mu.RUnlock()
			if err == sql.ErrNoRows {
				req.sender.messages <- NewMessage("e", "No Such Chat")
				continue
			} else if err != nil {
				log.Println("Error: Could not search for user", err)
				req.sender.messages <- NewMessage("e", "An error occured")
				continue
			}

			sentChatPassword = strings.TrimSpace(sentChatPassword)
			if sentChatPassword == chatPassword {
				cm.mu.Lock()
				res, err := joinChat.Exec(req.sender.username, chatId)
				cm.mu.Unlock()
				if err != nil {
					log.Println("Error: Could not join user to chat:", err)
					req.sender.messages <- NewMessage("e", "An error occured")
				}

				affected, err := res.RowsAffected()
				if err != nil {
					log.Println("Error: Could not get affected rows number:", err)
					req.sender.messages <- NewMessage("e", "An error occured")
				}

				if affected == 0 {
					req.sender.messages <- NewMessage("n", "Could not join, probably already joined")
				} else if affected == 1 {
					req.sender.messages <- NewMessage("n", chatId)
					req.sender.chats[chatId] = cm.chats[chatId].chatChan
					cm.chats[chatId].users[req.sender.username] = req.sender
				}
			}

		case LeaveChatRequestType:
			chatId := req.content
			
			var owner string
			cm.mu.RLock()
			err := getOwner.QueryRow(chatId).Scan(&owner)
			cm.mu.RUnlock()
			if err != nil {
				log.Println("Error: Could not leave chat:", err)
				req.sender.messages <- NewMessage("e", "An error occured")
				continue
			}

			if owner == req.sender.username {
				req.sender.messages <- NewMessage("n", "You are the owner of the chat, transfer the ownership of the chat or delete the chat.")
				continue
			}

			cm.mu.Lock()
			res, err := leaveChat.Exec(req.sender.username, chatId)
			cm.mu.Unlock()
			if err != nil {
				log.Println("Error: Could not leave chat:", err)
				req.sender.messages <- NewMessage("e", "An error occured")
				continue
			}

			affected, err := res.RowsAffected()
			if err != nil {
				log.Println("Error: Could not get affected rows number:", err)
				req.sender.messages <- NewMessage("e", "An error occured")
				continue
			}

			if affected != 0 {
				delete(req.sender.chats, chatId)
				delete(cm.chats[chatId].users, req.sender.username)
				req.sender.messages <- NewMessage("n", "Left " + chatId)
			}

		case NewChatRequestType:
			parts:= strings.Split(req.content, " ")
			numParts := len(parts)
			chatId := parts[0]
			chatName := strings.Join(parts[1:numParts-1], " ")
			password := parts[numParts-1]

			cm.mu.Lock()
			_, err = addChat.Exec(chatId, chatName, password, req.sender.username)
			cm.mu.Unlock()
			if sqliteErr, ok := err.(sqlite3.Error); ok {
				if sqliteErr.Code == sqlite3.ErrConstraint {
					req.sender.messages <- NewMessage("n", "ChatId already taken")
					break
				}
			} else if err != nil{
				log.Println("Error: Could not add chat to chata table", err)
			}
			
			newChat := NewChat(chatId, chatName, req.sender.username, cm.mu)
			cm.chats[chatId] = newChat
			req.sender.chats[chatId] = newChat.chatChan
			go newChat.HandleRequests()
			cm.ManagerChan <- JoinChatRequest(chatId, password, req.sender)
			req.sender.messages <- NewMessage("n", "Created new chat: " + chatId)

		case DeleteChatRequestType:
			chatId, chatPassword, _ := strings.Cut(req.content, " ")

			var owner string
			cm.mu.RLock()
			err := getOwner.QueryRow(chatId).Scan(&owner)
			cm.mu.RUnlock()
			if err != nil {
				log.Println("Error: Could not leave chat:", err)
				req.sender.messages <- NewMessage("e", "An error occured")
				continue
			}

			if owner != req.sender.username {
				req.sender.messages <- NewMessage("n", "You are not the owner of the chat")
				continue
			}

			cm.mu.Lock()
			res, err := deleteChat.Exec(chatId, chatPassword)
			cm.mu.Unlock()
			if err != nil {
				log.Println("Error: Could not delete user:", err)
			}

			affected, err := res.RowsAffected()
			if err != nil {
				log.Println("Error: Could not get affected rows number:", err)
			}

			if affected != 0 {
				cm.chats[chatId].chatChan <- DeleteChatRequest(chatId, chatPassword, req.sender)
				delete(cm.chats, chatId)
			}
		}
	}
}
