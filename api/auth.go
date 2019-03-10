package api

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/pquerna/otp/totp"

	"github.com/MyHomeworkSpace/api-server/auth"

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
	Status             string `json:"status"`
	User               User   `json:"user"`
	Grade              int    `json:"grade"`
	Tabs               []Tab  `json:"tabs"`
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Username           string `json:"username"`
	Email              string `json:"email"`
	Type               string `json:"type"`
	Features           string `json:"features"`
	Level              int    `json:"level"`
	ShowMigrateMessage int    `json:"showMigrateMessage"`
	TwoFactorVerified  int    `json:"twoFactorVerified"`
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

	e.POST("/auth/enableTOTP", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		rows, err := DB.Query("SELECT email, twoFactorVerified FROM users WHERE id = ?", GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("getting user email and 2fa status", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		var userEmail string
		var verified int

		if verified == 1 {
			return c.JSON(http.StatusAlreadyReported, ErrorResponse{"error", "already_enabled"})
		}

		if rows.Next() {
			rows.Scan(&userEmail, &verified)
		} else {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if verified == 1 {
			return c.JSON(http.StatusAlreadyReported, ErrorResponse{"error", "already_enabled"})
		}

		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "MyHomeworkSpace",
			AccountName: userEmail,
		})

		if err != nil {
			ErrorLog_LogError("generating 2FA key", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		_, err = DB.Exec("UPDATE users SET twoFactorSecret = ? WHERE id = ?", key.Secret(), GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("setting 2fa key in db", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		image, err := key.Image(200, 200)
		if err != nil {
			ErrorLog_LogError("generating 2fa QR code", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		var pngImage bytes.Buffer

		err = png.Encode(&pngImage, image)
		if err != nil {
			ErrorLog_LogError("encoding 2fa qr code to png", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		encodedString := base64.StdEncoding.EncodeToString(pngImage.Bytes())

		imageurl := "data:image/png;base64," + encodedString

		return c.JSON(http.StatusOK, TwoFactorEnabled{"ok", key.Secret(), imageurl})
	})

	e.POST("/auth/verifyTOTP", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if c.FormValue("code") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		rows, err := DB.Query("SELECT twoFactorSecret, twoFactorVerified FROM users WHERE id = ?", GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("getting 2fa details from DB", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		var secret string
		var verified int

		if rows.Next() {
			rows.Scan(&secret, &verified)
		} else {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if verified == 1 {
			return c.JSON(http.StatusAlreadyReported, ErrorResponse{"error", "already_verified"})
		}
		if totp.Validate(c.FormValue("code"), secret) {
			_, err = DB.Exec("UPDATE users SET twoFactorVerified = 1 WHERE id = ?", GetSessionUserID(&c))
			if err != nil {
				ErrorLog_LogError("getting 2fa details from DB", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			return c.JSON(http.StatusOK, StatusResponse{"ok"})
		}
		return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "incorrect_code"})
	})

	e.POST("/auth/disableTOTP", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if c.FormValue("code") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		rows, err := DB.Query("SELECT twoFactorSecret, twoFactorVerified from users where id = ?", GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("getting user information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		var twoFactorSecret string
		var twoFactorVerified int

		if rows.Next() {
			rows.Scan(&twoFactorSecret, &twoFactorVerified)
		} else {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if twoFactorVerified == 0 {
			return c.JSON(http.StatusFailedDependency, ErrorResponse{"error", "two_factor_not_enabled"})
		}

		if totp.Validate(c.FormValue("code"), twoFactorSecret) {
			_, err = DB.Exec("UPDATE users SET twoFactorSecret = NULL, twoFactorVerified = 0 WHERE id = ?", GetSessionUserID(&c))
			return c.JSON(http.StatusOK, StatusResponse{"ok"})
		}

		return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "two_factor_incorrect"})
	})

	e.POST("/auth/login", func(c echo.Context) error {
		if c.FormValue("username") == "" || c.FormValue("password") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}
		data, resp, err := auth.DaltonLogin(strings.ToLower(c.FormValue("username")), c.FormValue("password"))
		if resp != "" || err != nil {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", resp})
		}
		if WhitelistEnabled {
			file, err := os.Open(WhitelistFile)
			if err != nil {
				ErrorLog_LogError("getting whitelist", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
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
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", WhitelistBlockMsg})
			}
		}
		rows, err := DB.Query("SELECT id, twoFactorSecret, twoFactorVerified from users where username = ?", c.FormValue("username"))
		if err != nil {
			ErrorLog_LogError("getting user information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		session := auth.SessionInfo{-1}
		if rows.Next() {
			// exists, use it
			userID := -1
			var twoFactorVerified int
			var twoFactorSecret string
			rows.Scan(&userID, &twoFactorSecret, &twoFactorVerified)
			if twoFactorVerified == 1 {
				if c.FormValue("code") == "" {
					return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "two_factor_required"})
				}
				if !totp.Validate(c.FormValue("code"), twoFactorSecret) {
					return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "two_factor_incorrect"})
				}
			}
			session = auth.SessionInfo{userID}
		} else {
			// doesn't exist
			stmt, err := DB.Prepare("INSERT INTO users(name, username, email, type, showMigrateMessage) VALUES(?, ?, ?, ?, 0)")
			if err != nil {
				ErrorLog_LogError("trying to set user information", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}
			res, err := stmt.Exec(data["fullname"], c.FormValue("username"), c.FormValue("username")+"@dalton.org", data["roles"].([]string)[0])
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
			session = auth.SessionInfo{int(lastID)}
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
		if err == ErrDataNotFound {
			return c.JSON(http.StatusOK, ErrorResponse{"error", "user_record_missing"})
		} else if err != nil {
			ErrorLog_LogError("getting user information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		grade, err := Data_GetUserGrade(user)
		if err != nil {
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
			Grade:  grade,
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
			TwoFactorVerified:  user.TwoFactorVerified,
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
