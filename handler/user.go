package handler

import (
	"errors"
	"pair-project/model"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = []byte("your-secret-key")

type UserHandler struct {
	DB *gorm.DB
}

func NewUserHandler(db *gorm.DB) UserHandler {
	return UserHandler{DB: db}
}

func (u *UserHandler) Register(c echo.Context) error {
	newUser := new(model.Users)

	if err := c.Bind(newUser); err != nil {
		return c.JSON(400, echo.Map{
			"message": "invalid request",
		})
	}

	if newUser.Email == "" || newUser.Password == "" {
		return c.JSON(400, echo.Map{
			"message": "email and password are required",
		})
	}

	existingUser := new(model.Users)
	result := u.DB.Where("email = ?", newUser.Email).First(existingUser)
	if result.RowsAffected > 0 {
		return c.JSON(400, echo.Map{
			"message": "email already registered",
		})
	}

	var lastUserID int
	u.DB.Model(&model.Users{}).Select("user_id").Order("user_id desc").Limit(1).Scan(&lastUserID)
	newUser.UserID = lastUserID + 1

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.JSON(500, echo.Map{
			"message": "failed to hash password",
			"detail":  err.Error(),
		})
	}

	newUser.Password = string(hashedPassword)
	newUser.Role = "customer"

	result = u.DB.Create(&newUser)
	if result.Error != nil {
		return c.JSON(500, echo.Map{
			"message": "failed to register user",
			"detail":  result.Error.Error(),
		})
	}

	responseBody := model.Users{
		UserID:  newUser.UserID,
		Email:   newUser.Email,
		Deposit: newUser.Deposit,
	}

	return c.JSON(201, echo.Map{
		"message":   "success register",
		"user_info": responseBody,
	})
}

func (uh *UserHandler) Login(c echo.Context) error {
	loginRequest := new(model.Users)
	if err := c.Bind(loginRequest); err != nil {
		return c.JSON(400, echo.Map{
			"message": "invalid request",
		})
	}

	user := new(model.Users)
	result := uh.DB.Where("email = ?", loginRequest.Email).First(user)
	if result.Error != nil {
		return c.JSON(401, echo.Map{
			"message": "invalid email or password",
		})
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password))
	if err != nil {
		return c.JSON(401, echo.Map{
			"message": "invalid email or password",
		})
	}

	token, err := GenerateJWTToken(user)
	if err != nil {
		return c.JSON(500, echo.Map{
			"message": "failed to generate JWT token",
			"detail":  err.Error(),
		})
	}

	return c.JSON(200, echo.Map{
		"message": "login success",
		"token":   token,
	})
}

func GenerateJWTToken(user *model.Users) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["user_id"] = user.UserID
	claims["role"] = user.Role
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	signedToken, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (uh *UserHandler) GetInfoUser(c echo.Context) error {
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(401, echo.Map{
			"message": "unauthorized",
		})
	}

	userRole, ok := c.Get("role").(string)
	if !ok {
		return c.JSON(401, echo.Map{
			"message": "unauthorized",
		})
	}

	if userRole != "customer" {
		return c.JSON(403, echo.Map{
			"message": "forbidden",
		})
	}

	user := new(model.Users)
	result := uh.DB.First(&user, userID)
	if result.Error != nil {
		return c.JSON(500, echo.Map{
			"message": "failed to get user information",
			"detail":  result.Error.Error(),
		})
	}

	return c.JSON(200, echo.Map{
		"user_info": user,
	})
}

func (u *UserHandler) GetAllEquipment(c echo.Context) error {
	var equipments []model.Equipments
	result := u.DB.Find(&equipments)
	if result.Error != nil {
		return c.JSON(500, echo.Map{
			"message": "failed to fetch equipments",
			"detail":  result.Error.Error(),
		})
	}

	return c.JSON(200, echo.Map{
		"equipments": equipments,
	})
}

func (u *UserHandler) RentEquipment(c echo.Context) error {
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(401, echo.Map{
			"message": "unauthorized",
		})
	}

	rentRequest := new(model.RentEquipment)
	if err := c.Bind(rentRequest); err != nil {
		return c.JSON(400, echo.Map{
			"message": "invalid request",
		})
	}

	if rentRequest.RentEquipmentID == 0 || rentRequest.Quantity == 0 || rentRequest.StartDate == "" || rentRequest.EndDate == "" {
		return c.JSON(400, echo.Map{
			"message": "equipment_id, quantity, start_date, and end_date are required",
		})
	}

	var lastRentID int
	u.DB.Model(&model.Rents{}).Select("rent_id").Order("rent_id desc").Limit(1).Scan(&lastRentID)

	newRent := model.Rents{
		UserID:        userID,
		PaymentDate:   time.Now().Format("2006-01-02"),
		PaymentStatus: "pending",
	}

	newRent.RentID = lastRentID + 1

	result := u.DB.Create(&newRent)
	if result.Error != nil {
		return c.JSON(500, echo.Map{
			"message": "failed to create rent entry",
			"detail":  result.Error.Error(),
		})
	}

	var lastRentEquipmentID int
	u.DB.Model(&model.RentEquipment{}).Select("rent_equipment_id").Order("rent_equipment_id desc").Limit(1).Scan(&lastRentEquipmentID)

	equipment := new(model.Equipments)
	result = u.DB.First(&equipment, rentRequest.EquipmentID)
	if result.Error != nil {
		return c.JSON(500, echo.Map{
			"message": "failed to get equipment information",
			"detail":  result.Error.Error(),
		})
	}

	if equipment.Stock < rentRequest.Quantity {
		return c.JSON(400, echo.Map{
			"message": "insufficient stock for the requested quantity",
		})
	}

	totalRentalCost := rentRequest.Quantity * equipment.RentalCost

	newRentEquipment := model.RentEquipment{
		RentID:          newRent.RentID,
		EquipmentID:     rentRequest.EquipmentID,
		Quantity:        rentRequest.Quantity,
		StartDate:       rentRequest.StartDate,
		EndDate:         rentRequest.EndDate,
		TotalRentalCost: totalRentalCost,
	}

	newRentEquipment.RentEquipmentID = lastRentEquipmentID + 1

	result = u.DB.Create(&newRentEquipment)
	if result.Error != nil {
		return c.JSON(500, echo.Map{
			"message": "failed to rent equipment",
			"detail":  result.Error.Error(),
		})
	}

	equipment.Stock -= rentRequest.Quantity
	u.DB.Save(&equipment)

	return c.JSON(201, echo.Map{
		"message":   "equipment rented successfully",
		"rent_info": newRentEquipment,
		"equipment": equipment,
	})
}

// func (u *UserHandler) Topup(c echo.Context) error {

// }

func (u *UserHandler) CreateEquipment(c echo.Context) error {
	userRole, ok := c.Get("role").(string)
	if !ok {
		return c.JSON(401, echo.Map{
			"message": "unauthorized",
		})
	}

	if userRole != "admin" {
		return c.JSON(403, echo.Map{
			"message": "forbidden",
		})
	}

	newEquipment := new(model.Equipments)
	err := c.Bind(newEquipment)
	if err != nil {
		return c.JSON(400, echo.Map{
			"message": "invalid request",
		})
	}

	if newEquipment.Name == "" || newEquipment.Stock == 0 || newEquipment.RentalCost == 0 || newEquipment.Category == "" {
		return c.JSON(400, echo.Map{
			"message": "name, stock, rental_cost, and category are required",
		})
	}

	var lastEquipmentID int
	u.DB.Model(&model.Equipments{}).Select("equipment_id").Order("equipment_id desc").Limit(1).Scan(&lastEquipmentID)

	newEquipment.EquipmentID = lastEquipmentID + 1

	result := u.DB.Create(&newEquipment)
	if result.Error != nil {
		return c.JSON(500, echo.Map{
			"message": "failed to create equipment",
			"detail":  result.Error.Error(),
		})
	}

	return c.JSON(201, echo.Map{
		"message":   "equipment create successfully",
		"equipment": newEquipment,
	})
}

func (u *UserHandler) DeleteEquipment(c echo.Context) error {
	userRole, ok := c.Get("role").(string)
	if !ok {
		return c.JSON(401, echo.Map{
			"message": "unauthorized",
		})
	}

	if userRole != "admin" {
		return c.JSON(403, echo.Map{
			"message": "forbidden",
		})
	}

	equipmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(400, echo.Map{
			"message": "invalid equipment ID",
		})
	}

	existingEquipment := new(model.Equipments)
	result := u.DB.First(&existingEquipment, equipmentID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(404, echo.Map{
				"message": "equipment not found",
			})
		}
		return c.JSON(500, echo.Map{
			"message": "failed to retrieve equipment",
		})
	}

	result = u.DB.Delete(&existingEquipment, equipmentID)
	if result.Error != nil {
		c.JSON(500, echo.Map{
			"message": "failed to delete equipment",
		})
	}

	return c.JSON(200, echo.Map{
		"message":   "equipment deleted successfully",
		"equipment": existingEquipment,
	})

}

func (u *UserHandler) UpdateEquipment(c echo.Context) error {
	userRole, ok := c.Get("role").(string)
	if !ok {
		return c.JSON(401, echo.Map{
			"message": "unauthorized",
		})
	}

	if userRole != "admin" {
		return c.JSON(403, echo.Map{
			"message": "forbidden",
		})
	}

	equipmentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(400, echo.Map{
			"message": "invalid equipment ID",
		})
	}

	existingEquipment := new(model.Equipments)
	result := u.DB.First(&existingEquipment, equipmentID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(404, echo.Map{
				"message": "equipment not found",
			})
		}
		return c.JSON(500, echo.Map{
			"message": "failed to retrieve equipment",
		})
	}

	updateWorkout := new(model.Equipments)
	err = c.Bind(updateWorkout)
	if err != nil {
		return c.JSON(400, echo.Map{
			"message": "invalid request",
		})
	}

	if updateWorkout.Name == "" || updateWorkout.Stock == 0 || updateWorkout.RentalCost == 0 || updateWorkout.Category == "" {
		return c.JSON(400, echo.Map{
			"message": "name, stock, rental_cost, and category are required",
		})
	}

	existingEquipment.Name = updateWorkout.Name
	existingEquipment.Stock = updateWorkout.Stock
	existingEquipment.RentalCost = updateWorkout.RentalCost
	existingEquipment.Category = updateWorkout.Category

	result = u.DB.Save(existingEquipment)
	if result.Error != nil {
		return c.JSON(500, echo.Map{
			"message": "failed to update equipment",
		})
	}

	return c.JSON(200, echo.Map{
		"message":   "equipment updated successfully",
		"equipment": existingEquipment,
	})
}
