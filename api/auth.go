package api

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/MyHomeworkSpace/api-server/auth"

	"github.com/labstack/echo"
)

type CSRFResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

type UserResponse struct {
	Status             string `json:"status"`
	User               User   `json:"user"`
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Username           string `json:"username"`
	Email              string `json:"email"`
	Type               string `json:"type"`
	Features           string `json:"features"`
	Level              int    `json:"level"`
	ShowMigrateMessage int    `json:"showMigrateMessage"`
}

func HasAuthToken(c *echo.Context) bool {
	return (*c).Request().Header.Get("Authorization") != ""
}

func GetAuthToken(c *echo.Context) string {
	headerParts := strings.Split((*c).Request().Header.Get("Authorization"), " ")
	if len(headerParts) != 2 {
		return ""
	} else {
		return headerParts[1]
	}
}

func GetSessionUserID(c *echo.Context) int {
	return GetSessionInfo(c).UserId
}

func GetSessionInfo(c *echo.Context) auth.SessionInfo {
	if HasAuthToken(c) {
		// we have an authorization header, use that
		token := GetAuthToken(c)
		if token == "" {
			return auth.SessionInfo{-1, ""}
		}
		return auth.GetSessionFromAuthToken(token)
	} else {
		cookie, err := (*c).Cookie("session")
		if err != nil {
			return auth.SessionInfo{-1, ""}
		}
		return auth.GetSession(cookie.Value)
	}
}

func InitAuthAPI(e *echo.Echo) {
	e.POST("/auth/clearMigrateFlag", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}

		stmt, err := DB.Prepare("UPDATE users SET showMigrateMessage = 0 WHERE id = ?")
		if err != nil {
			log.Println("Error while clearing migration flag: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = stmt.Exec(GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while clearing migration flag: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.GET("/auth/csrf", func(c echo.Context) error {
		cookie, _ := c.Cookie("csrfToken")
		jsonResp := CSRFResponse{"ok", cookie.Value}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.POST("/auth/login", func(c echo.Context) error {
		if c.FormValue("username") == "" || c.FormValue("password") == "" {
			jsonResp := ErrorResponse{"error", "Missing required parameters."}
			return c.JSON(http.StatusUnprocessableEntity, jsonResp)
		}
		data, resp, err := auth.DaltonLogin(strings.ToLower(c.FormValue("username")), c.FormValue("password"))
		if resp != "" || err != nil {
			jsonResp := ErrorResponse{"error", resp}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if WhitelistEnabled {
			file, err := os.Open(WhitelistFile)
			if err != nil {
				log.Println("Error while getting whitelist: ")
				log.Println(err)
				jsonResp := ErrorResponse{"error", "Internal server error"}
				return c.JSON(http.StatusInternalServerError, jsonResp)
			}
			scanner := bufio.NewScanner(file)
			found := false
			for scanner.Scan() {
				if strings.ToLower(scanner.Text()) == strings.ToLower(c.FormValue("username")) {
					found = true
					break
				}
			}
			file.Close()
			if !found {
				log.Printf("Blocked signin attempt by %s because they aren't on the whitelist\n", c.FormValue("username"))
				jsonResp := ErrorResponse{"error", WhitelistBlockMsg}
				return c.JSON(http.StatusInternalServerError, jsonResp)
			}
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
			stmt, err := DB.Prepare("INSERT INTO users(name, username, email, type, showMigrateMessage) VALUES(?, ?, ?, ?, 0)")
			if err != nil {
				log.Println("Error while trying to set user information: ")
				log.Println(err)
				jsonResp := ErrorResponse{"error", "Internal server error"}
				return c.JSON(http.StatusInternalServerError, jsonResp)
			}
			res, err := stmt.Exec(data["fullname"], c.FormValue("username"), c.FormValue("username")+"@dalton.org", data["roles"].([]interface{})[0])
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
			// add default classes
			addClassesStmt, err := DB.Prepare("INSERT INTO `classes` (`name`, `userId`) VALUES ('Math', ?), ('History', ?), ('English', ?), ('Language', ?), ('Science', ?)")
			if err != nil {
				log.Println("Error while trying to add default classes: ")
				log.Println(err)
				jsonResp := ErrorResponse{"error", "Internal server error"}
				return c.JSON(http.StatusInternalServerError, jsonResp)
			}
			_, err = addClassesStmt.Exec(int(lastId), int(lastId), int(lastId), int(lastId), int(lastId))
			if err != nil {
				log.Println("Error while trying to add default classes: ")
				log.Println(err)
				jsonResp := ErrorResponse{"error", "Internal server error"}
				return c.JSON(http.StatusInternalServerError, jsonResp)
			}
			session = auth.SessionInfo{int(lastId), c.FormValue("username")}
		}
		cookie, _ := c.Cookie("session")
		auth.SetSession(cookie.Value, session)
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.GET("/auth/me", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		user, err := Data_GetUserByID(GetSessionUserID(&c))
		if err == ErrDataNotFound {
			return c.JSON(http.StatusOK, ErrorResponse{"error", "user_record_missing"})
		} else if err != nil {
			log.Println("Error while getting user information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		return c.JSON(http.StatusOK, UserResponse{
			Status: "ok",
			User:   user,

			// these are set for backwards compatibility
			ID:                 user.ID,
			Name:               user.Name,
			Username:           user.Username,
			Email:              user.Email,
			Type:               user.Type,
			Features:           user.Features,
			Level:              user.Level,
			ShowMigrateMessage: user.ShowMigrateMessage,
		})
	})

	e.GET("/auth/logout", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		cookie, _ := c.Cookie("session")
		newSession := auth.SessionInfo{-1, ""}
		auth.SetSession(cookie.Value, newSession)
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})
}
