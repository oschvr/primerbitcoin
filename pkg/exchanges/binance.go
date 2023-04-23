package exchanges

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"log"
	"primerbitcoin/database"
	"primerbitcoin/utils"
	"strconv"
)

// CreateOrder runs a custom buy/sell order
func CreateOrder(client *binance.Client, qty string, symbol string, side string) {

	// Get price
	price, err := getPrice(client, symbol)
	if err != nil {
		utils.Logger.Errorf("Could not get price for %s", symbol)
	}

	utils.Logger.Infof("Buying %s of %s", qty, symbol)

	// Prepare database insert
	query, err := database.DB.Prepare("INSERT INTO orders(symbol, quantity, price, success) VALUES (?,?,?,?)")
	if err != nil {
		log.Panicf("Unable to prepare order statement")
	}
	defer func(query *sql.Stmt) {
		err := query.Close()
		if err != nil {

		}
	}(query)

	// Define side, if its buy or sell
	var sideType binance.SideType
	if side == "buy" || side == "BUY" {
		sideType = binance.SideTypeBuy
	} else {
		sideType = binance.SideTypeSell
	}

	// Use the Binance API client to execute a market buy order for BTC
	order, err := client.NewCreateOrderService().Symbol(symbol).Side(sideType).Type(binance.OrderTypeMarket).Quantity(qty).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	// Execute sql statement
	_, err = query.Exec(symbol, qty, price, true)
	if err != nil {
		log.Panicf("Unable to persist order in db")
	}

	parsedQty, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
	parsedPrice, _ := strconv.ParseFloat(price, 64)
	total := parsedQty * parsedPrice

	utils.Logger.Infof("Bought %.4f BTC at price of %.4f USDT per BTC. Cost: %.4f", parsedQty, parsedPrice, total)
}

// GetBalance will get the balances for the clients account
func GetBalance(client *binance.Client) {
	utils.Logger.Info("Getting balance")
	accountService := client.NewGetAccountService()
	account, err := accountService.Do(context.Background())
	if err != nil {
		log.Panicf("Unable to get account information")
	}

	for _, balance := range account.Balances {
		fmt.Printf("%s: %s\n", balance.Asset, balance.Free)
	}
}

// Get price for symbol
func getPrice(client *binance.Client, symbol string) (string, error) {
	utils.Logger.Infof("Getting price for %s", symbol)
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
