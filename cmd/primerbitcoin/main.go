package main

import (
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	bitsosdk "github.com/xiam/bitso-go/bitso"
	"net/http"
	"os"
	"primerbitcoin/database"
	"primerbitcoin/pkg/bitso"
	"primerbitcoin/pkg/config"
	"primerbitcoin/pkg/metrics"
	"primerbitcoin/pkg/utils"
	"sync"
	"time"
)

var cfg config.Config
var version = "dev"

func main() {

	//Create global config
	config.DecodeConfig(&cfg)

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

	// Create waitGroup
	wg := &sync.WaitGroup{}
	wg.Add(1)

	// Add custom metrics
	metrics.RecordMetrics()

	// Handle metrics server
	http.Handle("/metrics", promhttp.Handler())

	// Start metrics server concurrently
	wg.Add(1)
	go func() {
		utils.Logger.Fatal(http.ListenAndServe(":2112", nil))
		wg.Done()
	}()

	// Define necessary configuration options
	apiKey := os.Getenv("API_KEY")
	apiSecret := os.Getenv("SECRET_KEY")

	/// BITSO ---------
	client := bitsosdk.NewClient(nil)
	client.SetAPIKey(apiKey)
	client.SetAPISecret(apiSecret)

	// Create scheduler
	scheduler := gocron.NewScheduler(time.UTC)

	if os.Getenv("PRODUCTION") == "true" {
		bitso.CreateOrder(client, cfg)
	} else {

		// Configure job
		job, err := scheduler.Tag(os.Getenv("APP_NAME")).Cron(cfg.Scheduler.Schedule).Do(func() {
			// Run Create Order
			bitso.CreateOrder(client, cfg)
		})

		if err != nil {
			utils.Logger.Errorf("Unable to run cronjob %s", err)
		}

		utils.Logger.Infof("Running job %s with cron schedule %s", job.Tags(), cfg.Scheduler.Schedule)

		// Start scheduler concurrently
		wg.Add(1)
		go func() {
			scheduler.StartBlocking()
			wg.Done()
		}()

		// Block indefinitely
		wg.Wait()
	}

	/// BINANCE ---------
	//// Create a new Binance API client (USE TESTNET)
	//isProd, _ := strconv.ParseBool(os.Getenv("PRODUCTION"))
	//binance.UseTestnet = isProd
	//
	//client := binance.NewClient(apiKey, apiSecret)

}
