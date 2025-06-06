package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Creates the users table in the database.
func CreateUsersTable() {
	const userTable = `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		password TEXT NOT NULL,
		created_at TEXT NOT NULL DEFAULT(datetime('now'))
	);
	`

	db, err := sql.Open("sqlite3", DatabasePath)
	defer db.Close()


	if err != nil {
		log.Fatalln("ERROR: COULD NOT OPEN DATABASE:", err)
	}

	stmnt, err := db.Prepare(userTable)
	defer stmnt.Close()
	if err != nil {
		log.Fatalln("ERROR: COULD NOT PREPARE STATMENT:", err)
	}

	_, err = stmnt.Exec()
	if err != nil {
		log.Fatalln("ERROR: COULD CREATE USERS TABLE:", err)
	}
	log.Println("Users table created")
}

// Create the logged_in table in the database.
// it contains data about which chats are each user in.
func CreateJoinedTable() {
	const loggedInTable = `
	CREATE TABLE IF NOT EXISTS joined (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL REFERENCES users(username) ON DELETE CASCADE,
		chatId TEXT NOT NULL REFERENCES chats(chatId) ON DELETE CASCADE,
		joined_at TEXT NOT NULL DEFAULT(datetime('now'))
	);
	`

	db, err := sql.Open("sqlite3", DatabasePath)
	defer db.Close()

	if err != nil {
		log.Fatalln("ERROR: COULD NOT OPEN DATABASE:", err)
	}

	stmnt, err := db.Prepare(loggedInTable)
	defer stmnt.Close()
	if err != nil {
		log.Fatalln("ERROR: COULD NOT PREPARE STATMENT:", err)
	}

	_, err = stmnt.Exec()
	if err != nil {
		log.Fatalln("ERROR: COULD CREATE LOGGED IN TABLE:", err)
	}
	log.Println("Logged in table created")
}
