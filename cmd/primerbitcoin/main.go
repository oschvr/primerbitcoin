package main

import (
	"fmt"
	binance "github.com/adshao/go-binance/v2"
	"github.com/common-nighthawk/go-figure"
	"github.com/joho/godotenv"
	"os"
	"primerbitcoin/database"
	"primerbitcoin/pkg/exchanges"
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

	// Run Create Order
	exchanges.GetBalance(client)
	exchanges.CreateOrder(client, btcToBuy, "BTCUSDT", "BUY")
	cronjob := func() {
		// Get the current time
		t := time.Now()

		// Check if the minute is divisible by 2c
		if t.Minute()%2 == 0 {
			// Run the buyBTC function
			exchanges.CreateOrder(client, btcToBuy, "BTCUSDT", "BUY")
		}
	}

	// Run the cronjob indefinitely
	for {
		// Get the current time
		t := time.Now()

		// Check if the current time is divisible by 1 minute
		if t.Second()%60 == 0 {
			// Run the cronjob function
			cronjob()
		}

		// Sleep for 1 second
		time.Sleep(time.Second)
	}
}
