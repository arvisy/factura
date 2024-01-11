package model

type Rents struct {
	RentID        int    `json:"rent_id"`
	UserID        int    `json:"user_id"`
	PaymentDate   string `json:"payment_date"`
	PaymentStatus string `json:"payment_status"`
}
