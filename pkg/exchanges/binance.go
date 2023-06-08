package exchanges

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"log"
	"primerbitcoin/database"
	"primerbitcoin/pkg/config"
	"primerbitcoin/pkg/notifications"
	"primerbitcoin/pkg/utils"
	"strconv"
)

// MUST USE PRODUCTION, LIVE API KEYS TO USE EURBTC

func getAvgPrice(client *binance.Client, symbol string) (string, error) {
	utils.Logger.Infof("Getting avg price for %s", symbol)
	price, err := client.NewAveragePriceService().Symbol(symbol).Do(context.Background())
	if err != nil {
		utils.Logger.Error("Unable to get average prices information, ", err)
		return "", err
	} else {
		return price.Price, nil
	}

}

func GetSymbolInfoFromExchange(client *binance.Client, symbol string) {
	info, err := client.NewExchangeInfoService().Symbol(symbol).Do(context.Background())
	if err != nil {
		utils.Logger.Errorf("Unable to retrieve exchange info")
	}
	utils.Logger.Infof(fmt.Sprintf("%v", info.Symbols))
}

func getPriceForSymbol(client *binance.Client, symbol string) string {
	price, err := client.NewAveragePriceService().Symbol(symbol).Do(context.Background())
	if err != nil {
		utils.Logger.Errorf("Unable to retrieve price")
		return ""
	}
	utils.Logger.Infof(fmt.Sprintf("Price for %s is %s", symbol, price.Price))
	return price.Price
}

func CalculateAmount(client *binance.Client, config config.Config) string {
	// Disclaimer
	utils.Logger.Info("Amounts are parsed and calculated up to the 8th decimal")

	// Get price for symbol
	price := getPriceForSymbol(client, config.Order.Symbol)

	// Calculate amount to buy based on price and order quantity
	parsedQuantity, _ := strconv.ParseFloat(config.Order.Quantity, 64)
	parsedPrice, _ := strconv.ParseFloat(price, 64)
	amount := parsedQuantity / parsedPrice

	// Log and return
	utils.Logger.Infof(fmt.Sprintf("Amount for %s%s in %s is %.8f", config.Order.Quantity, config.Order.Minor, config.Order.Major, amount))
	return fmt.Sprintf("%.8f", amount)
}

// CreateOrder runs a custom buy/sell order
func CreateOrder(client *binance.Client, cfg config.Config) {

	var orderSettings = cfg.Order

	// Get balance
	balance := getBalance(client, orderSettings.Minor)

	// Get price
	//price, err := getPrice(client, orderSettings.Symbol)
	price, err := client.NewListPricesService().Symbol(orderSettings.Symbol).Do(context.Background())
	if err != nil {
		utils.Logger.Fatalf("Could not get price for %s", orderSettings.Symbol)
	}
	//utils.Logger.Infof("Buying %s%s of %s", orderSettings.Amount, orderSettings.Minor, orderSettings.Major)

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
	if orderSettings.Side == "buy" || orderSettings.Side == "BUY" {
		sideType = binance.SideTypeBuy
	} else {
		sideType = binance.SideTypeSell
	}

	//Calculate quantity (qty = price/amount)
	//quantity := strconv.ParseFloat(price,64) / orderSettings.Amount

	parsedPrice, _ := strconv.ParseFloat("100", 64)
	parsedQuantity, _ := strconv.ParseFloat(orderSettings.Quantity, 64)
	quantity := fmt.Sprintf("%.4f", parsedPrice/parsedQuantity)
	utils.Logger.Infof("Buying %.4f%s of %s", parsedQuantity, orderSettings.Minor, orderSettings.Major)

	// Use the Binance API client to execute a market buy order for BTC
	order, err := client.NewCreateOrderService().Symbol(orderSettings.Symbol).Side(sideType).Type(binance.OrderTypeMarket).Quantity(quantity).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	// Calculate balance and executed
	parsedBalance, _ := strconv.ParseFloat(balance, 64)
	parsedExecutedQty, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
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
	notifications.SendTelegramMessage(fmt.Sprintf("[🟢primerbitcoin] 🎉 You just bought %.4f of %s at price of %.2f in %s.\n 💸 Total spent : %.2f. \n 🏦Remaining balance of %.2f", parsedExecutedQty, orderSettings.Symbol, parsedPrice, "binance", totalExecuted, newBalance))

	utils.Logger.Infof("Order ID: %d. Bought %.4f of %s at price of %.2f in %s. 💸 Total spent : %.2f. Remaining balance of %.2f", order.OrderID, parsedExecutedQty, orderSettings.Symbol, parsedPrice, "binance", totalExecuted, newBalance)
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
	prices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil {
		utils.Logger.Error("Unable to get prices information, ", err)
		return "", err
	}

	for _, p := range prices {
		if p.Symbol == symbol {
			return p.Price, nil
		}
	}
	utils.Logger.Fatalf("Symbol %s not found", symbol)
	return "", nil
}
