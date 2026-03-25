package collision_b

// Price simulates a type like common.Price (domain layer)
type Price struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// ResponseBody uses collision_b.Price
type ResponseBody struct {
	Price Price `json:"price"`
}
