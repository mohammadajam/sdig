package server

import (
	"log"
	"net"
	"strings"
	"sync"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	username string
	name string
	conn net.Conn
	chats map[string]chan ClientRequest
	managerChan chan ClientRequest
	requests chan Request
	connected bool
	mu sync.RWMutex
}

func NewUser(conn net.Conn, managerChan chan ClientRequest) User {
	return User {
		conn: conn,
		managerChan: managerChan,
		connected: false,
	}
}


func (u *User) HandleUser() {
	var buffer []byte = make([]byte, 1240)

	for {
		n, err := u.conn.Read(buffer)
			
		if err != nil {
			log.Println("ERROR: Failed to read from user:", err)
		}

		message := string(buffer[:n])

		message_slices := strings.Split(message, " ")
		
		if u.connected == false {
			if buffer[0] == 'l' {
				username := strings.TrimSpace(message_slices[1])
				password := strings.TrimSpace(message_slices[2])

				db, err := sql.Open("sqlite3", "sdig.db")
				if err != nil {
					log.Println("ERROR: Failed to open database")
					u.conn.Write([]byte("l DatabaseError"))
					continue
				}

				var dbPassword string

				stmnt, err := db.Prepare("SELECT password FROM users WHERE username = ?")
				if err != nil {
					log.Fatalln("ERROR: COULD NOT PREPARE STATMENT:", err)
				}
				
				u.mu.RLock()
				err = stmnt.QueryRow(username).Scan(&dbPassword)
				if err == sql.ErrNoRows {
					u.conn.Write([]byte("l NoSuchUser"))
					continue
				}
				u.mu.RUnlock()
				if err != nil {
					log.Println("ERROR: An error occured during reading from database:", err)
					u.conn.Write([]byte("l DatabaseError"))
					continue
				}
				dbPassword = strings.TrimSpace(dbPassword)

				if dbPassword == password {
					u.conn.Write([]byte("l connected"))
					u.connected = true
					u.managerChan <- NewClientRequest("get chats", "logged in", u)
				}
			}
			continue
		}
		
		if buffer[0] == 'm' {
			client_message := strings.Join(message_slices[2:], " ")
			u.chats[message_slices[1]] <- NewClientRequest("send", client_message, u)
			
		}

	}
}

func (u *User) HandleUserRequest() {

}

func (u *User) HandleRequestsToUser() {
	for {
		select {

		}
	}
}
