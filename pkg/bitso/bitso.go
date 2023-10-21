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
	balances, err := client.Balances(nil)
	if err != nil {
		utils.Logger.Fatalln("Error getting balances", err)
	}

	for _, balance := range balances {
		if balance.Currency.String() == minor {
			return balance.Available.Float64()
		}
	}

	utils.Logger.Infof("Nothing found for %s", minor)
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

func calculateQuantity(client *bitsosdk.Client, cfg config.Config) float64 {
	// Get order settings
	var orderSettings = cfg.Order

	// Calculate price
	price := getPrice(client, orderSettings.Major, orderSettings.Minor)
	amount, err := strconv.ParseFloat(orderSettings.Amount, 64)
	if err != nil {
		utils.Logger.Errorf("Unable to parse amount")
	}
	quantity := amount / price

	return quantity
}

func estimateRunway(client *bitsosdk.Client, cfg config.Config) (float64, bool) {
	// Get order settings
	var orderSettings = cfg.Order

	// Get balance
	balance := getBalance(client, orderSettings.Minor)

	// Calculate quantity to buy
	quantity := calculateQuantity(client, cfg)
	utils.Logger.Infof("Calculated quantity based on price: %0.8f%s", quantity, orderSettings.Major)

	// Estimate runway
	// Minimum amount is 10mxn
	runway := balance / quantity
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

	// Disclaimer
	utils.Logger.Infof("Quantities are calculated up to the 8 decimal")

	// Get order settings
	var orderSettings = cfg.Order

	// Get balance
	balance := getBalance(client, orderSettings.Minor)
	utils.Logger.Infof("Current balance of %s is %0.2f", orderSettings.Minor, balance)

	// Get price for symbol
	price := getPrice(client, orderSettings.Major, orderSettings.Minor)
	utils.Logger.Infof("Price is %0.2f%s for 1 %s", price, orderSettings.Minor, orderSettings.Major)

	// Estimate runway
	_, canRun := estimateRunway(client, cfg)

	if canRun == false {
		utils.Logger.Errorf("Unable to create order. Not enough balance: %0.2f%s", balance, orderSettings.Minor)
		return
	}

	// Calculate quantity to buy
	quantity := calculateQuantity(client, cfg)

	// Check if the amount is more than the minimum
	amount := quantity * price

	// Prepare database insert
	stmt, err := database.DB.Prepare("INSERT INTO orders(exchange, symbol, quantity, amount, price, success, order_id) VALUES (?,?,?,?,?,?,?)")
	if err != nil {
		utils.Logger.Errorf("Unable to prepare order statement")
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			utils.Logger.Errorf("Unable to run order statement")
		}
	}(stmt)

	utils.Logger.Infof("Amount to buy of %0.2f for %0.8f%s", amount, quantity, orderSettings.Major)

	if amount < 10.0 {
		utils.Logger.Fatalf("Error: %0.2f is less than the minimum 10.00%s", amount, orderSettings.Minor)
		notifications.SendTelegramMessage(fmt.Sprintf("[🔴 primerbitcoin] %s: %0.2f is less than the minimum 10.00%s. Primerbitcoin can't run.", "bitso", amount, orderSettings.Minor))
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
		Major: bitsosdk.ToMonetary(quantity),
		Price: bitsosdk.ToMonetary(price),
	}

	// Place order
	order, err := client.PlaceOrder(orderPlacement)
	if err != nil {
		utils.Logger.Fatalf("Unable to place order, %s", err)
	}

	// Persist order in db
	if order != "" {
		// Execute sql statement
		_, err := stmt.Exec("BITSO", orderSettings.Book, quantity, amount, price, true, order)
		if err != nil {
			utils.Logger.Warn("Unable to persist order in db, ", err)
		}
	}

	// Refresh balance
	balance = getBalance(client, orderSettings.Minor)

	// Send msg to telegram
	notifications.SendTelegramMessage(fmt.Sprintf("[🟢primerbitcoin] 🎉 You just bought %.8f of %s at price of %.2f %s in %s.\n 💸 Total spent : %.2f \n 🏦Remaining balance: %.2f", quantity, orderSettings.Major, price, orderSettings.Minor, "bitso", amount, balance))

	utils.Logger.Infof("Order ID: %s. Bought %.8f of %s at price of %.2f %s in %s. 💸 Total spent : %.2f. Remaining balance: %.2f", order, quantity, orderSettings.Major, price, orderSettings.Minor, "bitso", amount, balance)

}
