package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const DatabasePath string = "sdig.db"

// Create the table for the chats in the database
func CreateChatsTable() {
	const chatTable = `
	CREATE TABLE IF NOT EXISTS chats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chatId TEXT UNIQUE NOT NULL,
		chatName TEXT NOT NULL,
		password TEXT NOT NULL
	);
	`
	db, err := sql.Open("sqlite3", DatabasePath)
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

func CreateMessageTable() {
	const messageTable = `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT REFERENCES users(username),
		chatId TEXT REFERENCES chats(chatId),
		content TEXT NOT NULL
	);
	`
	db, err := sql.Open("sqlite3", DatabasePath)
	defer db.Close()

	if err != nil {
		log.Fatalln("ERROR: COULD NOT OPEN DATABASE:", err)
	}

	stmnt, err := db.Prepare(messageTable)
	defer stmnt.Close()
	if err != nil {
		log.Fatalln("ERROR: COULD NOT PREPARE STATMENT:", err)
	}

	_, err = stmnt.Exec()
	if err != nil {
		log.Fatalln("ERROR: COULD CREATE MESSAGES TABLE:", err)
	}
	log.Println("Messages table created")
}
