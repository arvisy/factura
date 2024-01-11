package model

type Users struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Deposit  int    `json:"deposit"`
	Role     string `json:"role"`
}
