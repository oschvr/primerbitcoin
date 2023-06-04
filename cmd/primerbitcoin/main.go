package main

import (
	"fmt"
	binance "github.com/adshao/go-binance/v2"
	"github.com/common-nighthawk/go-figure"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"os"
	"primerbitcoin/database"
	"primerbitcoin/pkg/config"
	"primerbitcoin/pkg/exchanges"
	"primerbitcoin/pkg/utils"
	"time"
)

var cfg config.Config

func main() {

	//Create global config
	config.DecodeConfig(&cfg)

	//Get latest tag
	version := utils.GetLatestTag()

	// Load env vars from .env file
	err := godotenv.Load()
	if err != nil {
		utils.Logger.Panic("Error loading .env file, ", err)
		return
	}

	// Execute Migrations
	database.Migrate(cfg)

	// Create banner
	banner := figure.NewFigure(fmt.Sprintf("%s:%s", os.Getenv("APP_NAME"), version), "", true)
	banner.Print()

	// Define necessary configuration options
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("SECRET_KEY")
	// Create a new Binance API client (USE TESTNET)
	binance.UseTestnet = true
	client := binance.NewClient(apiKey, apiSecret)

	// Create scheduler
	scheduler := gocron.NewScheduler(time.UTC)

	// Configure job
	job, err := scheduler.Tag(os.Getenv("APP_NAME")).Cron(cfg.Scheduler.Schedule).Do(func() {
		// Run Create Order
		exchanges.CreateOrder(client, cfg.Order.Amount, cfg.Order.Symbol, cfg.Order.Side, cfg.Order.Minor)
	})

	if err != nil {
		utils.Logger.Errorf("Unable to run cronjob %s", err)
	}

	utils.Logger.Infof("Running job %s with cron schedule %s", job.Tags(), cfg.Scheduler.Schedule)

	// Start scheduler
	scheduler.StartBlocking()
}
