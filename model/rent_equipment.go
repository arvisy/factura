package model

type RentEquipment struct {
	RentEquipmentID int    `json:"rent_Equipment_id"`
	RentID          int    `json:"rent_id"`
	EquipmentID     int    `json:"equipment_id"`
	Quantity        int    `json:"quantity"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	TotalRentalCost int    `json:"total_rental_cost"`
}
