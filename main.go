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
	"github.com/labstack/echo/middleware"
)

type ErrorResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type CSRFResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func main() {
	log.Println("MyHomeworkSpace API Server")

	InitConfig()
	InitDatabase()
	InitRedis()

	api.AuthURLBase = config.Server.AuthURLBase
	api.DB = DB
	api.RedisClient = RedisClient
	api.ReverseProxyHeader = config.Server.ReverseProxyHeader
	api.WhitelistEnabled = config.Whitelist.Enabled
	api.WhitelistFile = config.Whitelist.WhitelistFile
	api.WhitelistBlockMsg = config.Whitelist.BlockMessage
	auth.DB = DB
	auth.RedisClient = RedisClient

	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.CORS.Enabled {
				c.Response().Header().Set("Access-Control-Allow-Origin", config.CORS.Origin)
				c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if strings.HasPrefix(c.Request().URL.Path, "/api_tester") {
				return next(c)
			}
			if strings.HasPrefix(c.Request().URL.Path, "/application/requestAuth") {
				return next(c)
			}
			_, err := c.Cookie("session")
			if err != nil {
				// user has no cookie, generate one
				cookie := new(http.Cookie)
				cookie.Name = "session"
				cookie.Path = "/"
				uid, err := auth.GenerateUID()
				if err != nil {
					return err
				}
				cookie.Value = uid
				cookie.Expires = time.Now().Add(7 * 24 * time.Hour)
				c.SetCookie(cookie)
			}

			// bypass csrf if they send an authorization header
			if c.Request().Header.Get("Authorization") != "" {
				return next(c)
			}

			// bypass csrf for special internal api (this requires the ip to be localhost so it's still secure)
			if strings.HasPrefix(c.Request().URL.Path, "/schedule/internal") {
				return next(c)
			}

			csrfCookie, err := c.Cookie("csrfToken")
			csrfToken := ""
			hasNoToken := false
			if err != nil {
				// user has no cookie, generate one
				cookie := new(http.Cookie)
				cookie.Name = "csrfToken"
				cookie.Path = "/"
				uid, err := auth.GenerateRandomString(40)
				if err != nil {
					return err
				}
				cookie.Value = uid
				cookie.Expires = time.Now().Add(12 * 4 * 7 * 24 * time.Hour)
				c.SetCookie(cookie)

				hasNoToken = true
				csrfToken = cookie.Value

				// let the next if block handle this
			} else {
				csrfToken = csrfCookie.Value
			}

			// bypass csrf token for /auth/csrf
			if strings.HasPrefix(c.Request().URL.Path, "/auth/csrf") {
				// did we just make up a token?
				if hasNoToken {
					// if so, return it
					// auth.go won't know the new token yet
					return c.JSON(http.StatusOK, CSRFResponse{"ok", csrfToken})
				} else {
					return next(c)
				}
			}

			if csrfToken != c.QueryParam("csrfToken") || hasNoToken {
				return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "csrfToken_invalid"})
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
	e.Start(fmt.Sprintf(":%d", config.Server.Port))
}
