package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

var DB *sql.DB

func init() {
	var err error
	DB, err = sql.Open("sqlite3", os.Getenv("DATABASE_PATH"))
	if err != nil {
		log.Fatalln("Unable to open database connection", err)
	}
}
