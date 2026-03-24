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

// MapPriceRequest has a map field whose value type is Price,
// used to verify $ref rewriting under additionalProperties.
type MapPriceRequest struct {
	Items map[string]Price `json:"items"`
}
