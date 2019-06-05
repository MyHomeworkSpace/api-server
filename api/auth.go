package api

import (
	"net/http"
	"strings"

	"github.com/pquerna/otp/totp"

	"github.com/MyHomeworkSpace/api-server/auth"
	"github.com/MyHomeworkSpace/api-server/data"

	"github.com/labstack/echo"
)

type CSRFResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

type SessionResponse struct {
	Status  string `json:"status"`
	Session string `json:"session"`
}

type UserResponse struct {
	Status             string     `json:"status"`
	User               data.User  `json:"user"`
	Tabs               []data.Tab `json:"tabs"`
	ID                 int        `json:"id"`
	Name               string     `json:"name"`
	Username           string     `json:"username"`
	Email              string     `json:"email"`
	Type               string     `json:"type"`
	Features           string     `json:"features"`
	Level              int        `json:"level"`
	ShowMigrateMessage int        `json:"showMigrateMessage"`
}

type TwoFactorEnabled struct {
	Status   string `json:"status"`
	Secret   string `json:"emergency"`
	ImageURL string `json:"qr"`
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
	return GetSessionInfo(c).UserID
}

func GetSessionInfo(c *echo.Context) auth.SessionInfo {
	if HasAuthToken(c) {
		// we have an authorization header, use that
		token := GetAuthToken(c)
		if token == "" {
			return auth.SessionInfo{-1}
		}
		return auth.GetSessionFromAuthToken(token)
	} else {
		cookie, err := (*c).Cookie("session")
		if err != nil {
			return auth.SessionInfo{-1}
		}
		return auth.GetSession(cookie.Value)
	}
}

func IsInternalRequest(c *echo.Context) bool {
	remoteAddr := (*c).Request().RemoteAddr
	if ReverseProxyHeader != "" {
		if (*c).Request().Header.Get(ReverseProxyHeader) != "" {
			header := strings.Split((*c).Request().Header.Get(ReverseProxyHeader), ",")
			remoteAddr = strings.TrimSpace(header[len(header)-1])
		}
	}

	if strings.Split(remoteAddr, ":")[0] == "127.0.0.1" || strings.HasPrefix(remoteAddr, "[::1]") {
		return true
	} else {
		return false
	}
}

func InitAuthAPI(e *echo.Echo) {
	e.POST("/auth/clearMigrateFlag", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		stmt, err := DB.Prepare("UPDATE users SET showMigrateMessage = 0 WHERE id = ?")
		if err != nil {
			ErrorLog_LogError("clearing migration flag", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("clearing migration flag", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.GET("/auth/csrf", func(c echo.Context) error {
		cookie, _ := c.Cookie("csrfToken")
		return c.JSON(http.StatusOK, CSRFResponse{"ok", cookie.Value})
	})

	e.POST("/auth/login", func(c echo.Context) error {
		if c.FormValue("username") == "" || c.FormValue("password") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		username := strings.ToLower(c.FormValue("username"))
		password := c.FormValue("password")

		data, resp, err := auth.DaltonLogin(username, password)
		if resp != "" || err != nil {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", resp})
		}

		rows, err := DB.Query("SELECT id FROM users WHERE username = ?", c.FormValue("username"))
		if err != nil {
			ErrorLog_LogError("getting user information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		session := auth.SessionInfo{
			UserID: -1,
		}
		if rows.Next() {
			// exists, use it
			userID := -1

			rows.Scan(&userID)

			enrolled2fa, err := isUser2FAEnrolled(userID)
			if err != nil {
				ErrorLog_LogError("getting user enrollment status", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			if enrolled2fa {
				if c.FormValue("code") == "" {
					return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "totp_required"})
				}

				secretRows, err := DB.Query("SELECT secret FROM totp WHERE userId = ?", userID)
				if err != nil {
					ErrorLog_LogError("getting user 2fa secret", err)
					return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				}
				defer secretRows.Close()

				secret := ""

				secretRows.Next()
				secretRows.Scan(&secret)

				if !totp.Validate(c.FormValue("code"), secret) {
					return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "bad_totp_code"})
				}
			}

			session = auth.SessionInfo{
				UserID: userID,
			}
		} else {
			// doesn't exist, insert new record
			stmt, err := DB.Prepare("INSERT INTO users(name, username, email, type, showMigrateMessage) VALUES(?, ?, ?, ?, 0)")
			if err != nil {
				ErrorLog_LogError("trying to set user information", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}
			res, err := stmt.Exec(data["fullname"], username, username+"@dalton.org", data["roles"].([]string)[0])
			if err != nil {
				ErrorLog_LogError("trying to set user information", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}
			lastID, err := res.LastInsertId()
			if err != nil {
				ErrorLog_LogError("trying to set user information", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			// add default classes
			addClassesStmt, err := DB.Prepare("INSERT INTO `classes` (`name`, `userId`) VALUES ('Math', ?), ('History', ?), ('English', ?), ('Language', ?), ('Science', ?)")
			if err != nil {
				ErrorLog_LogError("trying to add default classes", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			_, err = addClassesStmt.Exec(int(lastID), int(lastID), int(lastID), int(lastID), int(lastID))
			if err != nil {
				ErrorLog_LogError("trying to add default classes", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			session = auth.SessionInfo{
				UserID: int(lastID),
			}
		}

		cookie, _ := c.Cookie("session")
		auth.SetSession(cookie.Value, session)

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.GET("/auth/me", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		user, err := Data_GetUserByID(GetSessionUserID(&c))
		if err == data.ErrNotFound {
			return c.JSON(http.StatusOK, ErrorResponse{"error", "user_record_missing"})
		} else if err != nil {
			ErrorLog_LogError("getting user information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		tabs, err := Data_GetTabsByUserID(user.ID)
		if err != nil {
			ErrorLog_LogError("getting user information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, UserResponse{
			Status: "ok",
			User:   user,
			Tabs:   tabs,

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
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		cookie, _ := c.Cookie("session")
		newSession := auth.SessionInfo{-1}
		auth.SetSession(cookie.Value, newSession)
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.GET("/auth/session", func(c echo.Context) error {
		cookie, err := c.Cookie("session")
		if err != nil {
			return c.JSON(http.StatusOK, SessionResponse{"ok", ""})
		}
		return c.JSON(http.StatusOK, SessionResponse{"ok", cookie.Value})
	})
}
