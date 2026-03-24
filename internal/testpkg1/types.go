package testpkg1

// Price is a price type from testpkg1 (simulates web/requests.Price)
type Price struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

// RequestBody uses testpkg1.Price
type RequestBody struct {
	Price Price `json:"price"`
}
