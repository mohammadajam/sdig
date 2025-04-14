package server

import (
	"database/sql"
	"log"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)


type Chat struct {
	chatId string
	chatName string
	chatChan chan ClientRequest
	users map[string]*User
	mutex sync.RWMutex
}



func LoadChats() map[string]Chat {
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
		
		chats[strings.TrimSpace(chatId)] = NewChat(strings.TrimSpace(chatId), strings.TrimSpace(chatName))
	}
	
	err = rows.Err()
	if err != nil {
		log.Fatalln("ERROR: AN ERROR OCCURED DURING READING ROWS:", err)
	}

	return chats
}

func NewChat(chatId string, chatName string) (Chat) {
	return Chat {
		chatId: chatId,
		chatName: chatName,
		chatChan: make(chan ClientRequest),
	}
}

func (chat *Chat) HandleChat() {
	for {
		req := <- chat.chatChan
		if req.string == "send" {
			
		}

		switch (req.string) {
			case "send":
			case "get messages":
			case "disconnect":


		}
	}
}

type ChatManager struct {
	chats map[string]Chat
	ManagerChan chan ClientRequest
	mu sync.RWMutex
}

// NOTE: ClientRequest is requests by the client to a chat or to  the chat manager
type ClientRequest struct {
	string
	data string
	sender *User
}

// NOTE: Request is a request by a chat or the chat manager to a client
type Request struct {
	string
	data string
}

func NewChatManager() (ChatManager) {
	return ChatManager{
		chats: LoadChats(),
		ManagerChan: make(chan ClientRequest),
	}
}

func NewClientRequest(request string, data string, user *User) ClientRequest {
	return ClientRequest{
		string: request,
		data: data,
		sender: user,
	}
}

func (cm *ChatManager) HandleRequests() {
	db, err := sql.Open("sqlite3", "sdig.db")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO OPEN DATABASE:", err)
	}

	get_chats, err := db.Prepare("SELECT chatId FROM logged_in WHERE user=?")
	if err != nil {
		log.Fatalln("ERROR: FAILED TO PREPARE STATEMENT:", err)
	}

	// TODO: add other sql statements such as (add_to_chat, remove_from_chat, new_chat, delete_chat)
	// NOTE: should try adding other ones such as (rename_chat, change_password)

	for {
		req := <- cm.ManagerChan

		switch (req.string) {
			case "get chats":
				cm.mu.Lock()
				rows, err := get_chats.Query(req.sender.username)
				cm.mu.Unlock()
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
				}
				
		}
		log.Println(req)
	}
}
