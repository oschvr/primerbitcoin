package database

import (
	"database/sql"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
	"primerbitcoin/pkg/utils"
)

var DB *sql.DB
var err error

func init() {

	// Assign helper variables
	dbName := os.Getenv("DATABASE_NAME")
	dbPath := filepath.Join(".", dbName)
	utils.Logger.Infof("Database file is %s", dbPath)

	// Check if database exists
	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		utils.Logger.Info("Database file doesn't exist, creating... ")
		file, err := os.Create(dbName)
		err = file.Close()
		if err != nil {
			utils.Logger.Panic("Unable to create database file, ", err)
			return
		}
	}

	// Create DB instance
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		utils.Logger.Panic("Unable to open database connection, ", err)
	}
}