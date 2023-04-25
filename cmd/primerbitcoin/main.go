package main

import (
	binance "github.com/adshao/go-binance/v2"
	"github.com/common-nighthawk/go-figure"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"os"
	"primerbitcoin/database"
	"primerbitcoin/pkg/exchanges"
	"primerbitcoin/pkg/utils"
	"time"
)

func main() {

	// Load env vars from .env file
	err := godotenv.Load()
	if err != nil {
		utils.Logger.Panic("Error loading .env file, ", err)
		return
	}

	// Execute Migrations
	database.Migrate()

	// Create banner
	banner := figure.NewFigure(os.Getenv("APP_NAME"), "", true)
	banner.Print()

	// Define necessary configuration options
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("SECRET_KEY")
	amount := "0.001"

	// Create a new Binance API client (USE TESTNET)
	binance.UseTestnet = true
	client := binance.NewClient(apiKey, apiSecret)

	// Create scheduler
	scheduler := gocron.NewScheduler(time.UTC)

	// Configure job
	job, err := scheduler.Tag(os.Getenv("APP_NAME")).Every(1).Minute().Do(func() {
		// Run Create Order
		exchanges.CreateOrder(client, amount, "BTCUSDT", "BUY")
	})
	if err != nil {
		utils.Logger.Errorf("Unable to run cronjob %s", err)
	}

	utils.Logger.Infof("Running job %s", job.Tags())

	// Start scheduler
	scheduler.StartBlocking()
}
