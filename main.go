package main

import (
	_ "database/sql"
	"fmt"
	"net"

	"sdig/server"
	"sdig/database_sdig"
	_ "sdig/database_sdig"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:4000")

	database_sdig.CreateChatsTable()
	database_sdig.CreateUsersTable()
	database_sdig.CreateLoggedInChatsTable()


	chatManager := server.NewChatManager()

	go chatManager.HandleRequests()

	if err != nil {
		fmt.Println("ERROR: COULD NOT LISTEN:", err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("ERROR: COULD NOT ACCEPT CONNECTION:", err)
			continue
		}

		user := server.NewUser(conn, chatManager.ManagerChan)
		go user.HandleUser()


	}
}

