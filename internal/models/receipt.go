package models

// Receipt represents a processed receipt with its details
type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

// Item represents a single item in a receipt
type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

// Response types for the API endpoints
type ReceiptID struct {
	ID string `json:"id"`
}

type Points struct {
	Points int64 `json:"points"`
}
