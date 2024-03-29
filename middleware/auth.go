package middleware

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

var jwtSecret = []byte("your-secret-key")

func Authentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tokenString := c.Request().Header.Get("Authorization")

		if tokenString == "" {
			return c.JSON(401, echo.Map{
				"message": "unauthorized",
			})
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.JSON(401, echo.Map{
				"message": "unauthorized",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.JSON(401, echo.Map{
				"message": "unauthorized",
			})
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			return c.JSON(401, echo.Map{
				"message": "unauthorized",
			})
		}

		role, ok := claims["role"].(string)
		if !ok {
			return c.JSON(401, echo.Map{
				"message": "unauthorized",
			})
		}

		c.Set("user_id", int(userID))
		c.Set("role", role)

		return next(c)
	}
}

func CustomerAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role := c.Get("role").(string)

		if role != "customer" {
			return c.JSON(403, echo.Map{
				"message": "forbidden",
			})
		}

		return next(c)
	}
}

func AdminAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role := c.Get("role").(string)

		if role != "admin" {
			return c.JSON(403, echo.Map{
				"message": "forbidden",
			})
		}

		return next(c)
	}
}
