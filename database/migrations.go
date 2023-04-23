package database

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

func Migrate() string {
	// Migration for SQLite db
	driver, err := sqlite.WithInstance(DB, &sqlite.Config{})
	if err != nil {
		panic(err)
	}

	// Create a new migration source
	s, err := (&file.File{}).Open("database/migrations")
	if err != nil {
		panic(err)
	}

	// Find bin path
	binPath, err := os.Executable()
	if err != nil {
		panic(err)
	}

	// Get db path
	binDir := filepath.Dir(binPath)
	dbPath, err := filepath.Abs(filepath.Join(binDir, os.Getenv("DATABASE_PATH")))

	fmt.Println(dbPath)

	// Apply all available migrations
	m, err := migrate.NewWithInstance("file", s, dbPath, driver)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("No migrations to apply.")
			return "No migrations to apply !"
		} else {
			panic(err)
		}
		return "Migrations applied"
	}
	return "Migration successful"
}
