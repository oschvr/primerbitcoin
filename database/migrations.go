package database

import (
	"embed"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
	"os"
	"primerbitcoin/pkg/config"
	"primerbitcoin/pkg/utils"
)

//go:embed migrations/*.sql
var migrations embed.FS

func init() {
	// Load env vars
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
}

func Migrate(config config.Config) {

	// Embed migrations in binary
	migrationsEmbed, err := iofs.New(migrations, "migrations")
	if err != nil {
		utils.Logger.Panic("Unable to embed migrations in binary")
	}

	// Migration for SQLite db
	utils.Logger.Info("Migrating database")
	driver, err := sqlite.WithInstance(DB, &sqlite.Config{})
	if err != nil {
		utils.Logger.Error("Unable to initialize sqlite3 driver, ", err)
		panic(err)
	}

	// Define database URI
	dbUri := fmt.Sprintf("sqlite3://%s", config.Database.Host)
	utils.Logger.Infof("Running migrations against %s", dbUri)

	// Create a new migration source
	m, err := migrate.NewWithInstance("iofs", migrationsEmbed, os.Getenv("DATABASE_NAME"), driver)
	if err != nil {
		utils.Logger.Error("Unable to create migrations from source, ", err)
		panic(err)
	}

	// Run migrations (up)
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			utils.Logger.Info("No migrations to apply")
			return
		} else {
			utils.Logger.Error("Unable to run migrations", err)
			panic(err)
		}
		return
	}

	utils.Logger.Info("Migrations successful")
	return
}
