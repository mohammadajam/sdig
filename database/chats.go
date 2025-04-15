package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)


func CreateChatsTable() {
	const chatTable = `
	CREATE TABLE IF NOT EXISTS chats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chatId TEXT UNIQUE NOT NULL,
		chatName TEXT NOT NULL,
		password TEXT NOT NULL
	);
	`
	db, err := sql.Open("sqlite3", "sdig.db")
	defer db.Close()

	if err != nil {
		log.Fatalln("ERROR: COULD NOT OPEN DATABASE:", err)
	}

	stmnt, err := db.Prepare(chatTable)
	defer stmnt.Close()
	if err != nil {
		log.Fatalln("ERROR: COULD NOT PREPARE STATMENT:", err)
	}

	_, err = stmnt.Exec()
	if err != nil {
		log.Fatalln("ERROR: COULD CREATE CHATS TABLE:", err)
	}
	log.Println("Chats table created")
}








