package main

import (
	"fmt"
	binance "github.com/adshao/go-binance/v2"
	"github.com/common-nighthawk/go-figure"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"os"
	"primerbitcoin/database"
	"primerbitcoin/pkg/exchanges"
	"primerbitcoin/utils"
	"time"
)

func main() {

	// Execute Migrations
	database.Migrate()

	// Load env vars from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	// Define necessary configuration options
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("SECRET_KEY")
	btcToBuy := "0.001"

	// Create a new Binance API client (USE TESTNET)
	binance.UseTestnet = true
	client := binance.NewClient(apiKey, apiSecret)

	// Create banner
	banner := figure.NewFigure("primerbitcoin!", "", true)
	banner.Print()

	// Create scheduler
	scheduler := gocron.NewScheduler(time.UTC)

	// Configure job
	job, err := scheduler.Tag("primerbitcoin").Every(1).Minute().Do(func() {
		// Run Create Order
		exchanges.GetBalance(client)
		exchanges.CreateOrder(client, btcToBuy, "BTCUSDT", "BUY")
	})
	if err != nil {
		utils.Logger.Errorf("Unable to run cronjob %s", err)
	}

	utils.Logger.Infof("Running job %s", job.Tags())

	// Start scheduler
	scheduler.StartBlocking()
}
