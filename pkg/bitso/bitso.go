package bitso

import (
	bitsosdk "github.com/xiam/bitso-go/bitso"
	"primerbitcoin/pkg/config"
	"primerbitcoin/pkg/utils"
)

func getFundings(client *bitsosdk.Client) {
	fundings, err := client.Fundings(nil)
	if err != nil {
		utils.Logger.Fatalln("Error getting fundings", err)
	}

	for _, funding := range fundings {
		utils.Logger.Printf("Funding: %s", funding)
	}
}

func getBalance(client *bitsosdk.Client, minor string) float64 {
	utils.Logger.Info("Getting balance")
	balances, err := client.Balances(nil)
	if err != nil {
		utils.Logger.Fatalln("Error getting balances", err)
	}

	for _, balance := range balances {
		if balance.Currency.String() == minor {
			return balance.Available.Float64()
		}
	}
	return 0.0
}

// Symbol == Book
func getPriceForSymbol(client *bitsosdk.Client, major string, minor string) float64 {

	// New book for ticker
	book := bitsosdk.NewBook(bitsosdk.ToCurrency(major), bitsosdk.ToCurrency(minor))
	ticker, err := client.Ticker(book)
	if err != nil {
		utils.Logger.Fatalln("Error getting price, ", err)
	}

	//utils.Logger.Infof("Price for book %s is %s %s for 1 %s", ticker.Book, ticker.Ask, minor, major)

	price := ticker.Ask.Float64()
	return price
}

func estimateRunway(client *bitsosdk.Client, cfg config.Config) (float64, bool) {
	// Get order settings
	var orderSettings = cfg.Order

	// Get balance
	balance := getBalance(client, orderSettings.Minor)
	utils.Logger.Infof("Balance of %s is %0.6f", orderSettings.Minor, balance)

	// Get price for symbol
	price := getPriceForSymbol(client, orderSettings.Major, orderSettings.Minor)
	utils.Logger.Infof("Price is %0.4f%s for 1 %s", price, orderSettings.Minor, orderSettings.Major)

	return 0.0, false

}

// CreateOrder runs a custom buy/sell order
func CreateOrder(client *bitsosdk.Client, cfg config.Config) {
	estimateRunway(client, cfg)
}
