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

/*
 * routes
 */
func routeAuthClearMigrateFlag(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	_, err := DB.Exec("UPDATE users SET showMigrateMessage = 0 WHERE id = ?", c.User.ID)
	if err != nil {
		ErrorLog_LogError("clearing migration flag", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeAuthCsrf(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	cookie, _ := ec.Cookie("csrfToken")
	ec.JSON(http.StatusOK, CSRFResponse{"ok", cookie.Value})
}

func routeAuthLogin(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("username") == "" || ec.FormValue("password") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	username := strings.ToLower(ec.FormValue("username"))
	password := ec.FormValue("password")

	data, resp, err := auth.DaltonLogin(username, password)
	if resp != "" || err != nil {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", resp})
		return
	}

	rows, err := DB.Query("SELECT id FROM users WHERE username = ?", ec.FormValue("username"))
	if err != nil {
		ErrorLog_LogError("getting user information", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
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
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}

		if enrolled2fa {
			if ec.FormValue("code") == "" {
				ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "totp_required"})
				return
			}

			secretRows, err := DB.Query("SELECT secret FROM totp WHERE userId = ?", userID)
			if err != nil {
				ErrorLog_LogError("getting user 2fa secret", err)
				ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				return
			}
			defer secretRows.Close()

			secret := ""

			secretRows.Next()
			secretRows.Scan(&secret)

			if !totp.Validate(ec.FormValue("code"), secret) {
				ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "bad_totp_code"})
				return
			}
		}

		session = auth.SessionInfo{
			UserID: userID,
		}
	} else {
		// doesn't exist, insert new record
		res, err := DB.Exec(
			"INSERT INTO users(name, username, email, type, showMigrateMessage) VALUES(?, ?, ?, ?, 0)",
			data["fullname"], username, username+"@dalton.org", data["roles"].([]string)[0],
		)
		if err != nil {
			ErrorLog_LogError("trying to set user information", err)
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			ErrorLog_LogError("trying to set user information", err)
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}

		// add default classes
		_, err = DB.Exec(
			"INSERT INTO `classes` (`name`, `userId`) VALUES ('Math', ?), ('History', ?), ('English', ?), ('Language', ?), ('Science', ?)",
			int(lastID), int(lastID), int(lastID), int(lastID), int(lastID),
		)
		if err != nil {
			ErrorLog_LogError("trying to add default classes", err)
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}

		session = auth.SessionInfo{
			UserID: int(lastID),
		}
	}

	cookie, _ := ec.Cookie("session")
	auth.SetSession(cookie.Value, session)

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeAuthMe(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	tabs, err := Data_GetTabsByUserID(c.User.ID)
	if err != nil {
		ErrorLog_LogError("getting user information", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, UserResponse{
		Status: "ok",
		User:   *c.User,
		Tabs:   tabs,

		// these are set for backwards compatibility
		ID:                 c.User.ID,
		Name:               c.User.Name,
		Username:           c.User.Username,
		Email:              c.User.Email,
		Type:               c.User.Type,
		Features:           c.User.Features,
		Level:              c.User.Level,
		ShowMigrateMessage: c.User.ShowMigrateMessage,
	})
}

func routeAuthLogout(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	cookie, _ := ec.Cookie("session")
	newSession := auth.SessionInfo{-1}
	auth.SetSession(cookie.Value, newSession)
	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeAuthSession(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	cookie, err := ec.Cookie("session")
	if err != nil {
		ec.JSON(http.StatusOK, SessionResponse{"ok", ""})
		return
	}
	ec.JSON(http.StatusOK, SessionResponse{"ok", cookie.Value})
}
