package bitso

import (
	"database/sql"
	"fmt"
	bitsosdk "github.com/xiam/bitso-go/bitso"
	"primerbitcoin/database"
	"primerbitcoin/pkg/config"
	"primerbitcoin/pkg/notifications"
	"primerbitcoin/pkg/utils"
	"strconv"
)

func getBalance(client *bitsosdk.Client, minor string) float64 {
	utils.Logger.Info("Getting balance")
	balances, err := client.Balances(nil)
	if err != nil {
		utils.Logger.Fatalln("Error getting balances", err)
	}

	for _, balance := range balances {
		if balance.Currency.String() == minor {
			return balance.Available.Float64()
		} else {
			utils.Logger.Infof("Nothing found for %s", minor)
			return 0.0
		}
	}
	return 0.0
}

// Symbol == Book
func getPrice(client *bitsosdk.Client, major string, minor string) float64 {

	// New book for ticker
	book := bitsosdk.NewBook(bitsosdk.ToCurrency(major), bitsosdk.ToCurrency(minor))
	ticker, err := client.Ticker(book)
	if err != nil {
		utils.Logger.Fatalln("Error getting price, ", err)
	}

	price := ticker.Ask.Float64()
	return price
}

func calculateAmount(client *bitsosdk.Client, cfg config.Config) float64 {
	// Get order settings
	var orderSettings = cfg.Order

	// Disclaimer
	utils.Logger.Infof("Amounts are parsed and calculated up to the 8 decimal")

	// Calculate price
	price := getPrice(client, orderSettings.Major, orderSettings.Minor)
	quantity, err := strconv.ParseFloat(orderSettings.Quantity, 64)
	if err != nil {
		utils.Logger.Errorf("Unable to parse quantity")
	}
	amount := quantity / price

	return amount
}

func estimateRunway(client *bitsosdk.Client, cfg config.Config) (float64, bool) {
	// Get order settings
	var orderSettings = cfg.Order

	// Get balance
	balance := getBalance(client, orderSettings.Minor)
	utils.Logger.Infof("Balance of %s is %0.8f", orderSettings.Minor, balance)

	// Get price for symbol
	price := getPrice(client, orderSettings.Major, orderSettings.Minor)
	utils.Logger.Infof("Price is %0.2f%s for 1 %s", price, orderSettings.Minor, orderSettings.Major)

	// Calculate amount to buy
	amount := calculateAmount(client, cfg)
	utils.Logger.Infof("Amount to buy of %0.8f", amount)

	// Estimate runway
	// Minimum amount is 10mxn
	runway := balance / amount
	canRun := false

	switch {
	case runway > 3:
		canRun = true
		return runway, canRun
	case runway <= 3 && runway > 1:
		// Send warning to telegram
		notifications.SendTelegramMessage(fmt.Sprintf("[🟠 primerbitcoin] You're running low on balance (%.2f%s). Primerbitcoin will stop running after the next %.f runs.", balance, orderSettings.Minor, runway))
		canRun = true
		return runway, canRun

	case runway == 1:
		// Send alert to telegram
		notifications.SendTelegramMessage(fmt.Sprintf("[🔴 primerbitcoin] No more fiat balance to run (%.2f%s). Primerbitcoin will stop running after the next run", balance, orderSettings.Minor))
		canRun = true
		return runway, canRun

	case runway < 1:

		// Send alert to telegram
		notifications.SendTelegramMessage(fmt.Sprintf("[🔴 primerbitcoin] No more fiat balance to run (%.2f%s). Primerbitcoin can't run", balance, orderSettings.Minor))
		canRun = false
		return runway, canRun

	default:
		return runway, canRun
	}
}

// CreateOrder runs a custom buy/sell order
func CreateOrder(client *bitsosdk.Client, cfg config.Config) {
	// Get order settings
	var orderSettings = cfg.Order

	// Get balance
	balance := getBalance(client, orderSettings.Minor)
	utils.Logger.Infof("Balance of %s is %0.8f", orderSettings.Minor, balance)

	// Get price for symbol
	price := getPrice(client, orderSettings.Major, orderSettings.Minor)
	utils.Logger.Infof("Price is %0.2f%s for 1 %s", price, orderSettings.Minor, orderSettings.Major)

	// Estimate runway
	_, canRun := estimateRunway(client, cfg)

	if canRun == false {
		utils.Logger.Fatalf("Unable to create order. Not enough balance: %0.2f%s", balance, orderSettings.Minor)
		return
	}

	// Prepare database insert
	stmt, err := database.DB.Prepare("INSERT INTO orders(exchange, symbol, quantity, price, success, order_id) VALUES (?,?,?,?,?,?)")
	if err != nil {
		utils.Logger.Fatal("Unable to prepare order statement")
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	// Calculate amount to buy
	amount := calculateAmount(client, cfg)

	// Check if the quantity is more than the minimum
	quantity := amount * price

	utils.Logger.Infof("Amount to buy of %0.8f for %0.2f%s", amount, quantity, orderSettings.Minor)

	if quantity < 10.0 {
		utils.Logger.Fatalf("Error: %0.2f is less than the minimum 10.00%s", quantity, orderSettings.Minor)
		notifications.SendTelegramMessage(fmt.Sprintf("[🔴 primerbitcoin] %s: %0.2f is less than the minimum 10.00%s. Primerbitcoin can't run.", "bitso", quantity, orderSettings.Minor))
		return
	}

	// Define side, if its buy or sell
	var side bitsosdk.OrderSide
	if orderSettings.Side == "buy" || orderSettings.Side == "BUY" {
		side = bitsosdk.OrderSideBuy
	} else {
		side = bitsosdk.OrderSideSell
	}

	// Create book
	book := *bitsosdk.NewBook(bitsosdk.ToCurrency(orderSettings.Major), bitsosdk.ToCurrency(orderSettings.Minor))

	// Create order
	orderPlacement := &bitsosdk.OrderPlacement{
		Book:  book,
		Side:  side,
		Type:  bitsosdk.OrderTypeMarket,
		Minor: bitsosdk.ToMonetary(amount),
		Price: bitsosdk.ToMonetary(price),
	}

	// Place order
	order, err := client.PlaceOrder(orderPlacement)
	if err != nil {
		utils.Logger.Fatalf("Unable to place order, %s", err)
	}

	fmt.Println(order)
}
