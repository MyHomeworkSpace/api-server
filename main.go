package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/api"
	"github.com/MyHomeworkSpace/api-server/auth"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

type ErrorResponse struct {
	Status string `json:"status"`
	Error string `json:"error"`
}

func main() {
	log.Println("MyHomeworkSpace API Server")

	InitConfig()
	InitDatabase()
	api.DB = DB
	auth.DB = DB

	e := echo.New()
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.CORS.Enabled {
				c.Response().Header().Set("Access-Control-Allow-Origin", config.CORS.Origin)
				c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if strings.HasPrefix(c.Request().URI(), "/api_tester") {
				return next(c)
			}
			_, err := c.Cookie("session")
			if err != nil {
				// user has no cookie, generate one
				cookie := new(echo.Cookie)
				cookie.SetName("session")
				cookie.SetPath("/")
				uid, err := auth.GenerateUID()
				if err != nil {
					return err
				}
				cookie.SetValue(uid)
				cookie.SetExpires(time.Now().Add(12 * 4 * 7 * 24 * time.Hour))
				c.SetCookie(cookie)
				return next(c)
			}

			// bypass csrf token for /auth/csrf
			csrfCookie, err := c.Cookie("csrfToken")
			if err != nil {
				// user has no cookie, generate one
				cookie := new(echo.Cookie)
				cookie.SetName("csrfToken")
				cookie.SetPath("/")
				uid, err := auth.GenerateRandomString(40)
				if err != nil {
					return err
				}
				cookie.SetValue(uid)
				cookie.SetExpires(time.Now().Add(12 * 4 * 7 * 24 * time.Hour))
				c.SetCookie(cookie)
				jsonResp := ErrorResponse{"error", "csrfToken_created"}
				return c.JSON(http.StatusBadRequest, jsonResp)
			}

			if strings.HasPrefix(c.Request().URI(), "/auth/csrf") {
				return next(c)
			}

			if csrfCookie.Value() != c.QueryParam("csrfToken") {
				jsonResp := ErrorResponse{"error", "csrfToken_invalid"}
				return c.JSON(http.StatusBadRequest, jsonResp)
			}

			return next(c)
		}
	})
	e.Static("/api_tester", "api_tester")
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "MyHomeworkSpace API Server")
	})

	api.Init(e) // API init delayed because router must be started first

	log.Printf("Listening on port %d", config.Server.Port)
	e.Run(standard.New(fmt.Sprintf(":%d", config.Server.Port)))
}
