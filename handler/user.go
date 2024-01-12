package handler

import (
	"errors"
	"fmt"
	"net/http"
	"pair-project/model"
	"regexp"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/xendit/xendit-go"
	"github.com/xendit/xendit-go/invoice"
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

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(newUser.Email) {
		return c.JSON(400, echo.Map{
			"message": "invalid email format",
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

	err = sendConfirmationEmail(newUser.Email)
	if err != nil {
		// Jika gagal mengirim email, Anda dapat menangani kesalahan di sini
		return c.JSON(500, echo.Map{
			"message": "failed to send confirmation email",
			"detail":  err.Error(),
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

func sendConfirmationEmail(recipientEmail string) error {
	apiKey := "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"
	fromEmail := "ssmile2299@gmail.com"
	client := sendgrid.NewSendClient(apiKey)

	message := mail.NewSingleEmail(
		mail.NewEmail("Sender Name", fromEmail),
		"Registration Info",
		mail.NewEmail("Recipient Name", recipientEmail),
		"Thank you for registering! Your account has been successfully created.",
		"<p>Thank you for registering! Your account has been successfully created.</p>",
	)

	response, err := client.Send(message)
	if err != nil {
		return err
	}

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return nil
	} else {
		return fmt.Errorf("failed to send email, status code: %d", response.StatusCode)
	}
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

func (u *UserHandler) GetInfoUser(c echo.Context) error {
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
	result := u.DB.First(&user, userID)
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

	if rentRequest.EquipmentID == 0 || rentRequest.Quantity == 0 || rentRequest.StartDate == "" || rentRequest.EndDate == "" {
		return c.JSON(400, echo.Map{
			"message": "equipment_id, quantity, start_date, and end_date are required",
		})
	}

	newRent := model.Rents{
		UserID:        userID,
		PaymentDate:   time.Now().Format("2006-01-02"),
		PaymentStatus: "pending",
	}

	var existingPendingRent model.Rents

	result := u.DB.Where("user_id = ? AND payment_status = 'pending'", userID).First(&existingPendingRent)
	if result.Error == nil {
		newRent.RentID = existingPendingRent.RentID
	} else {
		var lastRentID int
		u.DB.Model(&model.Rents{}).Select("rent_id").Order("rent_id desc").Limit(1).Scan(&lastRentID)
		newRent.RentID = lastRentID + 1

		result = u.DB.Create(&newRent)
		if result.Error != nil {
			return c.JSON(500, echo.Map{
				"message": "failed to create rent entry",
				"detail":  result.Error.Error(),
			})
		}
	}

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

	var lastRentEquipmentID int
	u.DB.Model(&model.RentEquipment{}).Select("rent_equipment_id").Order("rent_equipment_id desc").Limit(1).Scan(&lastRentEquipmentID)
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

func (u *UserHandler) Payment(c echo.Context) error {
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"message": "unauthorized",
		})
	}

	var totalRentalCost int
	u.DB.
		Model(&model.RentEquipment{}).
		Joins("JOIN rents ON rent_equipments.rent_id = rents.rent_id").
		Joins("JOIN users ON rents.user_id = users.user_id").
		Select("COALESCE(SUM(rent_equipments.total_rental_cost), 0) AS total_rental_cost").
		Where("users.user_id = ?", userID).
		Scan(&totalRentalCost)

	xendit.Opt.SecretKey = "xnd_development_rlG0Cw5HcEjmlNu4dv4obsR46hiEKdzpoB1KwyGarmxl1KMVzBukIns0o94S"

	createInvoiceData := invoice.CreateParams{
		ExternalID: "your-external-id",
		Amount:     (float64(totalRentalCost)),
		PayerEmail: "user@example.com",
	}

	resp, err := invoice.Create(&createInvoiceData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "error creating invoice",
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "invoice created successfully",
		"invoice": resp,
	})
}

func (u *UserHandler) XenditCallback(c echo.Context) error {
	var payload map[string]interface{}
	if err := c.Bind(&payload); err != nil {
		fmt.Println("Error parsing Xendit callback payload:", err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "error parsing Xendit callback payload",
		})
	}

	fmt.Printf("Xendit Callback Payload: %+v\n", payload)

	userID, ok := payload["user_id"].(string)
	if !ok {
		fmt.Println("Error extracting user ID from Xendit callback payload")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "error extracting user ID from Xendit callback payload",
		})
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		fmt.Println("Error converting user ID to integer:", err)
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "error converting user ID to integer",
			"error":   err.Error(),
		})
	}

	var rent model.Rents
	if err := u.DB.Where("user_id = ?", userIDInt).First(&rent).Error; err != nil {
		fmt.Println("Error finding rent record:", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "error finding rent record",
			"error":   err.Error(),
		})
	}

	if rent.PaymentStatus != "success" {
		if err := u.DB.Model(&rent).Update("payment_status", "success").Error; err != nil {
			fmt.Println("Error updating payment status:", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"message": "error updating payment status",
				"error":   err.Error(),
			})
		}

		fmt.Println("Payment status updated to success. Additional logic executed.")
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Xendit callback processed successfully",
	})
}

func (u *UserHandler) Topup(c echo.Context) error {
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{
			"message": "unauthorized",
		})
	}

	var topupData model.Topup
	err := c.Bind(&topupData)
	if err != nil {
		return c.JSON(400, echo.Map{
			"message": "bad request",
		})
	}

	if topupData.TopupAmount == 0 {
		return c.JSON(400, echo.Map{
			"message": "topup_amount is required",
		})
	}

	newTopup := model.Topup{
		UserID:      userID,
		TopupAmount: topupData.TopupAmount,
		TopupDate:   time.Now().Format("2006-01-02 15:04:05"),
		Status:      "success",
	}

	lastTopup := model.Topup{}
	u.DB.Model(&model.Topup{}).Order("topup_id desc").First(&lastTopup)
	newTopup.TopupID = lastTopup.TopupID + 1

	result := u.DB.Create(&newTopup)
	if result.Error != nil {
		return c.JSON(500, echo.Map{
			"message": "internal server error",
		})
	}

	user := new(model.Users)
	result = u.DB.First(&user, userID)
	if result.Error != nil {
		return c.JSON(500, echo.Map{
			"message": "failed to get user information",
			"detail":  result.Error.Error(),
		})
	}

	user.Deposit += topupData.TopupAmount

	result = u.DB.Model(&model.Users{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"deposit": user.Deposit,
	})
	if result.Error != nil {
		return c.JSON(500, echo.Map{
			"message": "internal server error",
		})
	}

	return c.JSON(200, echo.Map{
		"message": "topup success",
		"user_id": userID,
		"deposit": user.Deposit,
	})

}

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

	result = u.DB.Model(&model.Equipments{}).Where("equipment_id = ?", equipmentID).Updates(existingEquipment)
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
