package models

type Order struct {
	Exchange string
	Symbol   string
	Quantity string
	Price    string
	Side     string
	OrderId  int
	Success  bool
}
