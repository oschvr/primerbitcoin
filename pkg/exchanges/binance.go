package exchanges

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/shopspring/decimal"
	"log"
	"primerbitcoin/database"
	"primerbitcoin/pkg/config"
	"primerbitcoin/pkg/notifications"
	"primerbitcoin/pkg/utils"
	"strconv"
)

// MUST USE PRODUCTION API KEYS TO USE EURBTC

func getPriceForSymbol(client *binance.Client, symbol string) string {
	price, err := client.NewAveragePriceService().Symbol(symbol).Do(context.Background())
	if err != nil {
		utils.Logger.Errorf("Unable to retrieve price")
		return ""
	}
	return price.Price
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

// calculateAmount will calculate the amount of crypto per fiat configured
func calculateAmount(client *binance.Client, symbol string, quantity string) float64 {
	// Disclaimer
	utils.Logger.Info("Amounts are parsed and calculated up to the 8th decimal")

	// Get price for symbol
	price := getPriceForSymbol(client, symbol)

	// Calculate amount to buy based on price and order quantity
	parsedQuantity, err := strconv.ParseFloat(quantity, 64)
	if err != nil {
		utils.Logger.Errorf("Unable to parse quantity")
	}
	parsedPrice, err := strconv.ParseFloat(price, 64)
	if err != nil {
		utils.Logger.Errorf("Unable to parse price")
	}
	amount := parsedQuantity / parsedPrice

	return amount
}

// getLotSize retrieves the LOT_SIZE filter from the exchange information
func getLotSize(client *binance.Client, symbol string, quantity string) decimal.Decimal {

	// Get information from exchange for symbol
	info, err := client.NewExchangeInfoService().Symbol(symbol).Do(context.Background())
	if err != nil {
		utils.Logger.Errorf("Unable to retrieve exchange info")
	}

	var lotSize decimal.Decimal
	for _, info := range info.Symbols {
		if info.Symbol == symbol {
			for _, filter := range info.Filters {
				if filter["filterType"].(string) == string(binance.SymbolFilterTypeLotSize) {
					lotSize, _ = decimal.NewFromString(filter["stepSize"].(string))
					break
				}
			}
		}
	}
	if lotSize.IsZero() {
		utils.Logger.Fatal("Lot size not found for the symbol")
	}
	utils.Logger.Infof("Lot size: %v", lotSize)
	return lotSize
}

// estimateRunway will get the balance in fiat, the order config
// and estimate if next and how many orders are possible to execute
// If estimatedRunway is < 3, send a notification
func estimateRunway(client *binance.Client, cfg config.Config) (float64, bool) {
	// Get order settings
	var orderSettings = cfg.Order

	// Get balance
	balance := getBalance(client, orderSettings.Minor)
	parsedBalance, _ := strconv.ParseFloat(balance, 64)

	// Get price for symbol
	price := getPriceForSymbol(client, orderSettings.Symbol)

	// Get amount of crypto per configure order (fiat)
	orderAmount := calculateAmount(client, orderSettings.Symbol, orderSettings.Quantity)
	balanceAmount := calculateAmount(client, orderSettings.Symbol, balance)

	// Log
	utils.Logger.Infof(fmt.Sprintf("Order amount for %s%s in %s is %.8f", orderSettings.Quantity, orderSettings.Minor, orderSettings.Major, orderAmount))
	utils.Logger.Infof("Current balance: %s", balance)
	utils.Logger.Infof(fmt.Sprintf("Balance amount for %s%s in %s is %.8f", orderSettings.Quantity, orderSettings.Minor, orderSettings.Major, balanceAmount))
	utils.Logger.Infof("Price for %s : %s", orderSettings.Symbol, price)

	// estimate runway
	estimatedRunway := balanceAmount / orderAmount
	canRun := false

	switch {
	case estimatedRunway > 3:
		canRun = true
		return estimatedRunway, canRun
	case estimatedRunway <= 3 && estimatedRunway > 1:
		// Send warning to telegram
		notifications.SendTelegramMessage(fmt.Sprintf("[🟠 primerbitcoin] You're running low on balance (%.2f%s). Primerbitcoin will stop running after the next %.f runs.", parsedBalance, orderSettings.Minor, estimatedRunway))
		canRun = true
		utils.Logger.Infof("Primerbitcoin can run: %t", canRun)
		return estimatedRunway, canRun

	case estimatedRunway == 1:
		// Send alert to telegram
		notifications.SendTelegramMessage(fmt.Sprintf("[🔴 primerbitcoin] No more fiat balance to run (%.2f%s). Primerbitcoin will stop running after the next run", parsedBalance, orderSettings.Minor))
		canRun = true
		utils.Logger.Infof("Primerbitcoin can run: %t", canRun)
		return estimatedRunway, canRun

	case estimatedRunway < 1:

		// Send alert to telegram
		notifications.SendTelegramMessage(fmt.Sprintf("[🔴 primerbitcoin] No more fiat balance to run (%.2f%s). Primerbitcoin can't run", parsedBalance, orderSettings.Minor))
		canRun = false
		utils.Logger.Infof("Primerbitcoin can run: %t", canRun)
		return estimatedRunway, canRun

	default:
		return estimatedRunway, canRun
	}
}

// CreateOrder runs a custom buy/sell order
func CreateOrder(client *binance.Client, cfg config.Config) {
	// Get order settings
	var orderSettings = cfg.Order

	// Get price for symbol
	price := getPriceForSymbol(client, orderSettings.Symbol)

	_, canRun := estimateRunway(client, cfg)
	utils.Logger.Infof("Primerbitcoin can create orders: %t", canRun)
	if canRun == false {
		utils.Logger.Infof("Unable to create order. Not enough balance")
		return
	}

	//// Create instance of Order model
	// Need to understand why would I do this.
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

	// Get amount of crypto per configure order (fiat)
	orderAmount := calculateAmount(client, orderSettings.Symbol, orderSettings.Quantity)
	lotSize, _ := getLotSize(client, orderSettings.Symbol, orderSettings.Quantity).Float64()

	// Calculate quantity as per lot size allowed
	var quantityLotSize float64
	for i := 0.0; i <= orderAmount; i += lotSize {
		quantityLotSize += lotSize
	}

	// Quantity of Major, equivalent to minor amount, adhering to lot size
	quantity := fmt.Sprintf("%.8f", quantityLotSize)

	// Use the Binance API client to execute a market buy order for BTC
	order, err := client.NewCreateOrderService().Symbol(orderSettings.Symbol).Side(sideType).Type(binance.OrderTypeMarket).Quantity(quantity).Do(context.Background())
	if err != nil {
		utils.Logger.Errorf("Unable to execute order of %s%s for %s%s: %v", quantity, orderSettings.Major, orderSettings.Quantity, orderSettings.Minor, err)
		return
	}

	// Get balance
	balance := getBalance(client, orderSettings.Minor)
	parsedBalance, _ := strconv.ParseFloat(balance, 64)

	// Get executed quantity
	parsedExecutedQty, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
	parsedPrice, _ := strconv.ParseFloat(price, 64)
	totalExecuted := parsedExecutedQty * parsedPrice

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
	notifications.SendTelegramMessage(fmt.Sprintf("[🟢primerbitcoin] 🎉 You just bought %.4f of %s at price of %.2f in %s.\n 💸 Total spent : %.2f. \n 🏦Remaining balance of %.2f", parsedExecutedQty, orderSettings.Symbol, parsedPrice, "binance", totalExecuted, parsedBalance))

	utils.Logger.Infof("Order ID: %d. Bought %.4f of %s at price of %.2f in %s. 💸 Total spent : %.2f. Remaining balance of %.2f", order.OrderID, parsedExecutedQty, orderSettings.Symbol, parsedPrice, "binance", totalExecuted, parsedBalance)
}

// GetSymbolInfoFromExchange will retrieve all possible information for a given symbol from the exchange
func GetSymbolInfoFromExchange(client *binance.Client, symbol string) {
	info, err := client.NewExchangeInfoService().Symbol(symbol).Do(context.Background())
	if err != nil {
		utils.Logger.Errorf("Unable to retrieve exchange info")
	}
	utils.Logger.Infof(fmt.Sprintf("%v", info.Symbols))
}
