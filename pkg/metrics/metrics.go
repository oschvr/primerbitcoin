package metrics

import (
	"database/sql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"primerbitcoin/database"
	"primerbitcoin/pkg/utils"
	"strconv"
	"time"
)

func getTotalQtyBought() float64 {
	// Prepare query
	stmt, err := database.DB.Prepare("SELECT sum(quantity) FROM orders")
	if err != nil {
		utils.Logger.Error("Unable to get total qty bought, ", err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			utils.Logger.Error("Unable to create total qty statement, ", err)
		}
	}(stmt)
	var result string

	// Query db
	err = stmt.QueryRow().Scan(&result)
	if err != nil {
		utils.Logger.Error("Unable to run total qty query, ", err)
	}

	// Parse result
	parsedQty, err := strconv.ParseFloat(result, 64)
	if err != nil {
		utils.Logger.Error("Unable to parse total qty query, ", err)
	}

	return parsedQty
}

func getTotalAmountSpent() float64 {
	// Prepare query
	stmt, err := database.DB.Prepare("SELECT sum(amount) FROM orders")
	if err != nil {
		utils.Logger.Error("Unable to get total amount spent, ", err)
	}
	defer func(stmt *sql.Stmt) {
		err := stmt.Close()
		if err != nil {
			utils.Logger.Error("Unable to create total amount spent statement, ", err)
		}
	}(stmt)
	var result string

	// Query db
	err = stmt.QueryRow().Scan(&result)
	if err != nil {
		utils.Logger.Error("Unable to run total amount spent query, ", err)
	}

	// Parse result
	parsedAmount, err := strconv.ParseFloat(result, 64)
	if err != nil {
		utils.Logger.Error("Unable to parse total amount spent query, ", err)
	}

	return parsedAmount
}

// RecordMetrics creates custom metrics for the prometheus metrics server
func RecordMetrics() {

	// Gauge - It represents a single numerical value that can arbitrarily go up and down.
	// Counter - It is a cumulative metric that represents a single monotonically increasing counter. It can only go up and be reset to zero on restart
	// Histogram - To measure latency or response sizes, we typically use Histogram.

	pbTotalQuantityBought := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "primerbitcoin_quantity_bought_total",
		Help: "Total quantity of Major currency bought",
	})
	go func() {
		qty := getTotalQtyBought()
		pbTotalQuantityBought.Add(qty)
	}()

	pbTotalAmountSpent := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "primerbitcoin_amount_spent_total",
		Help: "Total amount of minor currency spent",
	})
	go func() {
		amt := getTotalAmountSpent()
		pbTotalAmountSpent.Add(amt)
	}()

	opsProcessed := promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}
