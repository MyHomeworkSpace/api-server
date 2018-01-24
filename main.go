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
			if config.CORS.Enabled && len(config.CORS.Origins) > 0 {
				foundOrigin := ""
				for _, origin := range config.CORS.Origins {
					if origin == c.Request().Header.Get("Origin") {
						foundOrigin = origin
					}
				}

				if foundOrigin == "" {
					foundOrigin = config.CORS.Origins[0]
				}

				c.Response().Header().Set("Access-Control-Allow-Origin", foundOrigin)
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

			// check if they have an authorization header
			if c.Request().Header.Get("Authorization") != "" {
				// get the token
				headerParts := strings.Split(c.Request().Header.Get("Authorization"), " ")
				if len(headerParts) == 2 {
					authToken := headerParts[1]
					
					// look up token
					rows, err := DB.Query("SELECT applications.cors FROM application_authorizations INNER JOIN applications ON application_authorizations.applicationId = applications.id WHERE application_authorizations.token = ?", authToken)
					if err == nil {
						// IMPORTANT: if there's an error with the token, we just continue with the request
						// this is for backwards compatibility with old versions, where the token would always bypass csrf and only be checked when authentication was needed
						// this is ok because if someone is able to add a new header, it should not be in a scenario where csrf would be a useful defense
						// TODO: it would be much cleaner to just fail here if the token is bad. do any applications actually rely on this behavior?

						defer rows.Close()
						if rows.Next() {
							cors := ""
							err = rows.Scan(&cors)

							if err == nil && cors != "" {
								c.Response().Header().Set("Access-Control-Allow-Origin", cors)
								c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
							}
						}
					}
				}

				// also bypass csrf
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
