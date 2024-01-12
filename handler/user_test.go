package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pair-project/model"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRegister(t *testing.T) {
	e := echo.New()

	db, err := setupTestDB()
	if err != nil {
		t.Fatal(err)
	}

	userHandler := NewUserHandler(db)

	requestPayload := map[string]interface{}{
		"email":    "test2@example.com",
		"password": "testpassword",
	}

	payload, err := json.Marshal(requestPayload)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	err = userHandler.Register(c)

	assert.NoError(t, err)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "success register", responseBody["message"])

	userInfo := responseBody["user_info"].(map[string]interface{})
	assert.Equal(t, "test2@example.com", userInfo["email"])
	assert.Equal(t, float64(0), userInfo["deposit"])
}

func TestGetAllEquipment(t *testing.T) {
	e := echo.New()

	db, err := setupTestDB()
	if err != nil {
		t.Fatal(err)
	}

	userHandler := NewUserHandler(db)

	equipment := model.Equipments{
		Name:       "Laptop",
		Stock:      10,
		RentalCost: 50,
		Category:   "Electronics",
	}
	db.Create(&equipment)

	req := httptest.NewRequest(http.MethodGet, "/equipments", nil)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	err = userHandler.GetAllEquipment(c)

	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, responseBody["equipments"])

	equipments := responseBody["equipments"].([]interface{})
	assert.Len(t, equipments, 1)

	equipmentData := equipments[0].(map[string]interface{})
	assert.Equal(t, "Laptop", equipmentData["name"])
	assert.Equal(t, float64(10), equipmentData["stock"])
	assert.Equal(t, float64(50), equipmentData["rental_cost"])
	assert.Equal(t, "Electronics", equipmentData["category"])
}

func TestLogin_Success(t *testing.T) {
	e := echo.New()

	db, err := setupTestDB()
	if err != nil {
		t.Fatal(err)
	}

	userHandler := NewUserHandler(db)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatal(err)
	}
	user := model.Users{
		Email:    "test@example.com",
		Password: string(hashedPassword),
		Role:     "customer",
	}
	db.Create(&user)

	loginRequest := map[string]interface{}{
		"email":    "test@example.com",
		"password": "testpassword",
	}

	payload, err := json.Marshal(loginRequest)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	err = userHandler.Login(c)

	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, rec.Code)

	var responseBody map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "login success", responseBody["message"])

	assert.NotNil(t, responseBody["token"])
}

func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&model.Users{}, &model.Equipments{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
