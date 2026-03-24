package testpkg2

// Price is a price type from testpkg2 (simulates common.Price / domain model)
type Price struct {
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

// ResponseBody uses testpkg2.Price
type ResponseBody struct {
	Price Price `json:"price"`
}
