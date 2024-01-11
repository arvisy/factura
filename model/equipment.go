package model

type Equipments struct {
	EquipmentID int    `json:"equipment_id"`
	Name        string `json:"name"`
	Stock       int    `json:"stock"`
	RentalCost  int    `json:"rental_cost"`
	Category    string `json:"category"`
}
