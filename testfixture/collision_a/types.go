package collision_a

// Price simulates a type like web/requests.Price (API layer)
type Price struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

// RequestBody uses collision_a.Price
type RequestBody struct {
	Price Price `json:"price"`
}
