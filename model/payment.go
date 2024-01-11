package model

type Payment struct {
	PaymentID     int    `json:"payment_id"`
	RentID        int    `json:"rent_id"`
	PaymentMethod string `json:"payment_method"`
}
