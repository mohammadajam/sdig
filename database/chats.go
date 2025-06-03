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
		password TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT(datetime('now')),
		owner TEXT NOT NULL DEFAULT 'Dev' REFERENCES users(username) ON DELETE RESTRICT
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
		log.Fatalln("ERROR: COULD NOT CREATE CHATS TABLE:", err)
	}
	log.Println("Chats table created")
}

func CreateMessageTable() {
	const messageTable = `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL DEFAULT 'Unknown User' REFERENCES users(username) ON DELETE SET DEFAULT,
		chatId TEXT NOT NULL REFERENCES chats(chatId) ON DELETE CASCADE,
		content TEXT NOT NULL,
		date TEXT NOT NULL DEFAULT(datetime('now'))
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
		log.Fatalln("ERROR: COULD NOT CREATE MESSAGES TABLE:", err)
	}
	log.Println("Messages table created")
}

func CreateTables() {
	db, err := sql.Open("sqlite3", DatabasePath)
	defer db.Close()

	if err != nil {
		log.Fatalln("ERROR: COULD NOT OPEN DATABASE:", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatalln("ERROR: COULD NOT ENABLE foreign_keys:", err)
	}

	//stmnt, err := db.Prepare("PRAGMA foreign_keys = ON;")
	//defer stmnt.Close()
	//if err != nil {
	//	log.Fatalln("ERROR: COULD NOT PREPARE STATMENT:", err)
	//}

	//_, err = stmnt.Exec()
	//if err != nil {
	//	log.Fatalln("ERROR: COULD NOT ENABLE foreign_keys:", err)
	//}
	CreateUsersTable()
	CreateChatsTable()
	CreateMessageTable()
	CreateJoinedTable()
}
