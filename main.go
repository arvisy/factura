package main

import (
	"pair-project/config"
	"pair-project/handler"
	"pair-project/middleware"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	db := config.InitDB()

	userHandler := handler.NewUserHandler(db)
	user := e.Group("/users")
	{
		user.POST("/register", userHandler.Register)
		user.POST("/login", userHandler.Login)
		user.GET("/info", userHandler.GetInfoUser, middleware.Authentication, middleware.CustomerAuth)
		user.GET("/equipments", userHandler.GetAllEquipment)
		user.POST("/rents", userHandler.RentEquipment, middleware.Authentication, middleware.CustomerAuth)
		user.POST("/topup", userHandler.Topup, middleware.Authentication, middleware.CustomerAuth)
		user.POST("/payment", userHandler.Payment, middleware.Authentication, middleware.CustomerAuth)
		user.POST("/callback", userHandler.XenditCallback, middleware.Authentication, middleware.CustomerAuth)

		user.POST("/equipments", userHandler.CreateEquipment, middleware.Authentication, middleware.AdminAuth)
		user.DELETE("/equipments/:id", userHandler.DeleteEquipment, middleware.Authentication, middleware.AdminAuth)
		user.PUT("/equipments/:id", userHandler.UpdateEquipment, middleware.Authentication, middleware.AdminAuth)
	}

	e.Start(":8080")
}
