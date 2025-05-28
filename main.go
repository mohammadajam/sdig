package main

import (
	"log"
	"net"
	"sync"

	"sdig/database"
	"sdig/server"
)

func main() {
	database.CreateChatsTable()
	database.CreateUsersTable()
	database.CreateLoggedInChatsTable()

	var mu sync.RWMutex
	chatManager := server.NewChatManager(&mu)
	go chatManager.HandleRequests()

	listener, err := net.Listen("tcp", "0.0.0.0:4000")
	if err != nil {
		log.Fatalln("ERROR: COULD NOT LISTEN:", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("ERROR: COULD NOT ACCEPT CONNECTION:", err)
			continue
		}

		user := server.NewUser(conn, chatManager.ManagerChan)
		go user.HandleUserRequest()
		go user.HandleMessagesToUser()
	}
}
