package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/labstack/echo"
)

var DB *sql.DB

type StatusResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Status string `json:"status"`
	Error string `json:"error"`
}

type UserResponse struct {
	Status string `json:"status"`
	ID int `json:"id"`
	Name string `json:"name"`
	Username string `json:"username"`
	Email string `json:"email"`
	Type string `json:"type"`
	Features string `json:"features"`
}

func Init(e *echo.Echo) {
	e.GET("/status", func(c echo.Context) error {
		return c.String(http.StatusOK, "Alive")
	})

	InitAuthAPI(e)
	InitPlannerAPI(e)

	log.Println("API endpoints ready.")
}
