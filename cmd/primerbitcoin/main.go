package main

import (
	"context"
	"fmt"
	binance "github.com/adshao/go-binance/v2"
	"github.com/common-nighthawk/go-figure"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

var log = logrus.New()

func main() {

	// Set the formatter to include timestamps
	log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	// Load env vars from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	// Define necessary configuration options
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("SECRET_KEY")
	btcToBuy := "0.01"

	// Create a new Binance API client (USE TESTNET)
	binance.UseTestnet = true
	client := binance.NewClient(apiKey, apiSecret)

	// Create banner
	banner := figure.NewFigure("primerbitcoin!", "", true)
	banner.Print()

	// Schedule the cronjob to run every minute
	buyBTC(client, btcToBuy, "BTCUSDT")
	cronjob := func() {
		// Get the current time
		t := time.Now()

		// Check if the minute is divisible by 5
		if t.Minute()%5 == 0 {
			// Run the buyBTC function
			buyBTC(client, btcToBuy, "BTCUSDT")
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

// Define the cronjob function
func buyBTC(client *binance.Client, qty string, symbol string) {

	// Get price
	price, err := getPrice(client, symbol)
	if err != nil {
		log.Errorf("Could not get price for %s", symbol)
	}

	log.Infof("Buying %s of %s", qty, symbol)

	// Use the Binance API client to execute a market buy order for BTC
	order, err := client.NewCreateOrderService().Symbol("BTCUSDT").Side(binance.SideTypeBuy).Type(binance.OrderTypeMarket).Quantity(qty).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Infof("Bought %s BTC at price of %s USDT per BTC", order.ExecutedQuantity, price)
}

// Get price for symbol
func getPrice(client *binance.Client, symbol string) (string, error) {
	log.Infof("Getting price for %s", symbol)
	prices, err := client.NewListPricesService().Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return "", nil
	}

	for _, p := range prices {
		if p.Symbol == symbol {
			return p.Price, nil
		}
	}
	log.Panicf("Symbol %s not found", symbol)
	return "", nil
}
