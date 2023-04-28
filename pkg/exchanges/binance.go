package exchanges

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"log"
	"primerbitcoin/database"
	"primerbitcoin/pkg/notifications"
	"primerbitcoin/pkg/utils"
	"strconv"
)

// CreateOrder runs a custom buy/sell order
func CreateOrder(client *binance.Client, quantity string, symbol string, side string, minor string) {

	// Get balance
	balance := getBalance(client, minor)

	// Get price
	price, err := getPrice(client, symbol)
	if err != nil {
		utils.Logger.Errorf("Could not get price for %s", symbol)
	}
	utils.Logger.Infof("Buying %s of %s", quantity, symbol)

	//// Create instance of Order model
	//model := models.Order{Exchange: "binance", Symbol: symbol, Price: price, Success: true, Side: side, Quantity: quantity}

	// Prepare database insert
	stmt, err := database.DB.Prepare("INSERT INTO orders(exchange, symbol, quantity, price, success, order_id) VALUES (?,?,?,?,?,?)")
	if err != nil {
		log.Panicf("Unable to prepare order statement")
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {

		}
	}(stmt)

	// Define side, if its buy or sell
	var sideType binance.SideType
	if side == "buy" || side == "BUY" {
		sideType = binance.SideTypeBuy
	} else {
		sideType = binance.SideTypeSell
	}

	// Use the Binance API client to execute a market buy order for BTC
	order, err := client.NewCreateOrderService().Symbol(symbol).Side(sideType).Type(binance.OrderTypeMarket).Quantity(quantity).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	// Calculate
	parsedBalance, _ := strconv.ParseFloat(balance, 64)
	parsedExecutedQty, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
	parsedPrice, _ := strconv.ParseFloat(price, 64)
	totalExecuted := parsedExecutedQty * parsedPrice
	newBalance := parsedBalance - totalExecuted

	// Execute sql statement
	row, err := stmt.Exec("BINANCE", order.Symbol, order.ExecutedQuantity, price, true, order.OrderID)
	if err != nil {
		utils.Logger.Warn("Unable to persist order in db, ", err)
	}

	// Check if order persisted
	orderPersisted, err := row.RowsAffected()
	if err != nil {
		utils.Logger.Warn("Unable to persist order in db, ", err)
	} else if orderPersisted >= 1 {
		utils.Logger.Infof("Order %d persisted in db.", order.OrderID)
	}

	// Send msg to telegram
	notifications.SendTelegramMessage(fmt.Sprintf("[🟢primerbitcoin] 🎉 You just bought %.4f of %s at price of %.2f in %s.\n 💸 Total spent : %.2f. \n 🏦Remaining balance of %.2f", parsedExecutedQty, symbol, parsedPrice, "binance", totalExecuted, newBalance))

	utils.Logger.Infof("Order ID: %d. Bought %.4f of %s at price of %.2f in %s. 💸 Total spent : %.2f. Remaining balance of %.2f", order.OrderID, parsedExecutedQty, symbol, parsedPrice, "binance", totalExecuted, newBalance)
}

// GetBalance will get the balances for the clients account
func getBalance(client *binance.Client, minor string) string {
	utils.Logger.Info("Getting balance")
	accountService := client.NewGetAccountService()
	account, err := accountService.Do(context.Background())
	if err != nil {
		utils.Logger.Panic("Unable to get account information", err)
	}

	for _, balance := range account.Balances {
		if balance.Asset == minor {
			utils.Logger.Infof("Balance of %s is %s", balance.Asset, balance.Free)
			return balance.Free
		}
	}
	return ""
}

// Get price for symbol
func getPrice(client *binance.Client, symbol string) (string, error) {
	utils.Logger.Infof("Getting price for %s", symbol)
	prices, err := client.NewListPricesService().Do(context.Background())
	if err != nil {
		utils.Logger.Error("Unable to get prices information, ", err)
		return "", err
	}

	for _, p := range prices {
		if p.Symbol == symbol {
			return p.Price, nil
		}
	}
	utils.Logger.Warnf("Symbol %s not found", symbol)
	return "", nil
}
