// init_db.go
package main

import (
    "database/sql"
    "log"

    _ "github.com/mattn/go-sqlite3"
)

func main() {
    db, err := sql.Open("sqlite3", "./db/users.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    createTable := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        fullname TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL
    );
    `
    _, err = db.Exec(createTable)
    if err != nil {
        log.Fatal(err)
    }

    log.Println("Database initialized and users table created.")
}
