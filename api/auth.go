package api

import (
	"log"
	"net/http"

	"github.com/MyHomeworkSpace/api-server/auth"

	"github.com/labstack/echo"
)

type CSRFResponse struct {
	Status string `json:"status"`
	Token string `json:"token"`
}

func GetSessionUserID(c *echo.Context) int {
	cookie, err := (*c).Cookie("session")
	if err != nil {
		return -1
	}
	return auth.GetSession(cookie.Value()).UserId
}

func InitAuthAPI(e *echo.Echo) {
	e.GET("/auth/csrf", func(c echo.Context) error {
		cookie, _ := c.Cookie("csrfToken")
		jsonResp := CSRFResponse{"ok", cookie.Value()}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.POST("/auth/login", func(c echo.Context) error {
		if c.FormValue("username") == "" || c.FormValue("password") == "" {
			jsonResp := ErrorResponse{"error", "Missing required parameters."}
			return c.JSON(http.StatusUnprocessableEntity, jsonResp)
		}
		data, resp, err := auth.DaltonLogin(c.FormValue("username"), c.FormValue("password"))
		if resp != "" || err != nil {
			jsonResp := ErrorResponse{"error", resp}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		rows, err := DB.Query("SELECT id from users where username = ?", c.FormValue("username"))
		if err != nil {
			log.Println("Error while getting user information: ")
			log.Println(err)
			jsonResp := ErrorResponse{"error", "Internal server error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()
		session := auth.SessionInfo{-1, ""}
		if rows.Next() {
			// exists, use it
			userId := -1
			rows.Scan(&userId)
			session = auth.SessionInfo{userId, c.FormValue("username")}
		} else {
			// doesn't exist
			stmt, err := DB.Prepare("INSERT INTO users(name, username, email, type) VALUES(?, ?, ?, ?)")
			if err != nil {
				log.Println("Error while trying to set user information: ")
				log.Println(err)
				jsonResp := ErrorResponse{"error", "Internal server error"}
				return c.JSON(http.StatusInternalServerError, jsonResp)
			}
			res, err := stmt.Exec(data["fullname"], c.FormValue("username"), c.FormValue("username") + "@dalton.org", data["roles"].([]interface{})[0])
			if err != nil {
				log.Println("Error while trying to set user information: ")
				log.Println(err)
				jsonResp := ErrorResponse{"error", "Internal server error"}
				return c.JSON(http.StatusInternalServerError, jsonResp)
			}
			lastId, err := res.LastInsertId()
			if err != nil {
				log.Println("Error while trying to set user information: ")
				log.Println(err)
				jsonResp := ErrorResponse{"error", "Internal server error"}
				return c.JSON(http.StatusInternalServerError, jsonResp)
			}
			session = auth.SessionInfo{int(lastId), c.FormValue("username")}
		}
		cookie, _ := c.Cookie("session")
		auth.SetSession(cookie.Value(), session)
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.GET("/auth/me", func(c echo.Context) error {
		cookie, _ := c.Cookie("session")
		if auth.GetSession(cookie.Value()).UserId == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		rows, err := DB.Query("SELECT id, name, username, email, type, features FROM users WHERE id = ?", auth.GetSession(cookie.Value()).UserId)
		if err != nil {
			log.Println("Error while getting user information: ")
			log.Println(err)
			jsonResp := ErrorResponse{"error", "Internal server error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()
		if rows.Next() {
			// exists, use it
			jsonResp := UserResponse{"ok", -1, "", "", "", "", ""}
			rows.Scan(&jsonResp.ID, &jsonResp.Name, &jsonResp.Username, &jsonResp.Email, &jsonResp.Type, &jsonResp.Features)
			return c.JSON(http.StatusOK, jsonResp)
		} else {
			// doesn't exist
			jsonResp := ErrorResponse{"error", "Your user ID record is missing from the database. Please contact hello@myhomework.space for assistance."}
			return c.JSON(http.StatusOK, jsonResp)
		}
	})

	e.GET("/auth/logout", func(c echo.Context) error {
		cookie, _ := c.Cookie("session")
		if auth.GetSession(cookie.Value()).UserId == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		newSession := auth.SessionInfo{-1, ""}
		auth.SetSession(cookie.Value(), newSession)
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})
}
