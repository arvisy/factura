package model

type Topup struct {
	TopupID     int    `json:"topup_id"`
	UserID      int    `json:"user_id"`
	TopupAmount int    `json:"topup_amount"`
	TopupDate   string `json:"topup_date"`
	Status      string `json:"status"`
}
