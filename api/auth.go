package api

import (
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/pquerna/otp/totp"

	"github.com/MyHomeworkSpace/api-server/auth"
	"github.com/MyHomeworkSpace/api-server/config"
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

func isInternalRequest(c *echo.Context) bool {
	remoteAddr := (*c).Request().RemoteAddr
	if config.GetCurrent().Server.ReverseProxyHeader != "" {
		if (*c).Request().Header.Get(config.GetCurrent().Server.ReverseProxyHeader) != "" {
			header := strings.Split((*c).Request().Header.Get(config.GetCurrent().Server.ReverseProxyHeader), ",")
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

func routeAuthCreateAccount(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	/*
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
	*/
}

func routeAuthCsrf(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	cookie, _ := ec.Cookie("csrfToken")
	ec.JSON(http.StatusOK, CSRFResponse{"ok", cookie.Value})
}

func routeAuthLogin(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("email") == "" || ec.FormValue("password") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	email := ec.FormValue("email")
	password := ec.FormValue("password")

	// we check if they're already in our db
	userRows, err := DB.Query("SELECT id FROM users WHERE email = ?", email)
	if err != nil {
		ErrorLog_LogError("getting user information", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer userRows.Close()
	if userRows.Next() {
		// email is registered
		// this is the fun part

		userID := -1
		userRows.Scan(&userID)

		user, err := data.GetUserByID(userID)
		if err == data.ErrNotFound {
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "user_record_missing"})
			return
		} else if err != nil {
			ErrorLog_LogError("getting user information", err)
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}

		needsConversion := false
		// first we check for the easy path: they have a hash stored with us
		if user.PasswordHash != "" {
			// they do
			err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
			if err == bcrypt.ErrMismatchedHashAndPassword {
				// bye
				ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "creds_incorrect"})
				return
			} else if err != nil {
				ErrorLog_LogError("user login", err)
				ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				return
			}

			// if we got here, no error -> password correct
		} else {
			// they do not, are they a dalton member? (that is, do they have a username?)
			if user.Username != "" {
				// they are
				// this means we must authenticate with dalton
				_, resp, err := auth.DaltonLogin(user.Username, password)
				if resp != "" || err != nil {
					ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", resp})
					return
				}

				// the sign-in worked
				// flag the account for conversion after passing 2fa
				needsConversion = true
			}
		}

		// now we check for totp
		enrolled2fa, err := isUser2FAEnrolled(userID)
		if err != nil {
			ErrorLog_LogError("getting user 2fa enrollment status", err)
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

		if needsConversion {
			// if we got here, they signed in with dalton

			// generate their hash
			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				ErrorLog_LogError("converting Dalton user", err)
				ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				return
			}

			// save their hash
			_, err = DB.Exec("UPDATE users SET password = ? WHERE id = ?", string(hash), userID)
			if err != nil {
				ErrorLog_LogError("converting Dalton user", err)
				ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				return
			}
		}

		// if we've made it this far, they're signed in
		session := auth.SessionInfo{
			UserID: userID,
		}
		cookie, _ := ec.Cookie("session")
		auth.SetSession(cookie.Value, session)

		ec.JSON(http.StatusOK, StatusResponse{"ok"})
	} else {
		// email is not registered, bye
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "no_account"})
	}
}

func routeAuthMe(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	tabs, err := data.GetTabsByUserID(c.User.ID)
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
