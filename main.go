package main

import (
	"log"
	"net"
	"sync"

	"sdig/database"
	"sdig/server"
)

func main() {
	database.CreateTables()

	var mu sync.RWMutex
	serverManager := server.NewServerManager(&mu)
	go serverManager.HandleRequests()
	serverManager.StartChatsHandleRequests()

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

		user := server.NewUser(conn, serverManager.ManagerChan)
		go user.HandleUserRequest()
		go user.HandleMessagesToUser()
	}
}
