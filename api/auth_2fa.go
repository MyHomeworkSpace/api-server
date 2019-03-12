package api

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/pquerna/otp/totp"
)

type TOTPSecretResponse struct {
	Status   string `json:"status"`
	Secret   string `json:"secret"`
	ImageURL string `json:"imageURL"`
}

func isUser2FAEnrolled(userID int) (bool, error) {
	rows, err := DB.Query("SELECT COUNT(*) FROM totp WHERE userId = ?", userID)
	if err != nil {
		return false, err
	}

	secretCount := -1

	rows.Next()
	rows.Scan(&secretCount)

	if secretCount == 0 {
		return false, nil
	}

	return true, nil
}

func InitAuth2FAAPI(e *echo.Echo) {
	e.POST("/auth/2fa/beginEnroll", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		// are they enrolled already?
		enrolled, err := isUser2FAEnrolled(GetSessionUserID(&c))
		if enrolled {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "already_enrolled"})
		}

		// get their email
		user, err := Data_GetUserByID(GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("starting TOTP enrollment", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		// generate a new secret, store in redis
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "MyHomeworkSpace",
			AccountName: user.Email,
		})

		secret := key.Secret()
		redisKeyName := fmt.Sprintf("user:%d:totp_tmp", user.ID)

		redisResponse := RedisClient.Set(redisKeyName, secret, time.Hour)
		if redisResponse.Err() != nil {
			ErrorLog_LogError("starting TOTP enrollment", redisResponse.Err())
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		// generate qr code and data url
		image, err := key.Image(200, 200)
		if err != nil {
			ErrorLog_LogError("starting TOTP enrollment", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		var pngImage bytes.Buffer

		err = png.Encode(&pngImage, image)
		if err != nil {
			ErrorLog_LogError("starting TOTP enrollment", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		encodedString := base64.StdEncoding.EncodeToString(pngImage.Bytes())

		imageURL := "data:image/png;base64," + encodedString

		return c.JSON(http.StatusOK, TOTPSecretResponse{"ok", secret, imageURL})
	})

	e.POST("/auth/2fa/completeEnroll", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		userID := GetSessionUserID(&c)

		if c.FormValue("code") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		code := c.FormValue("code")

		// are they enrolled already?
		enrolled, err := isUser2FAEnrolled(GetSessionUserID(&c))
		if enrolled {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "already_enrolled"})
		}

		// do they have a stored secret in redis?
		redisKeyName := fmt.Sprintf("user:%d:totp_tmp", userID)
		secret, err := RedisClient.Get(redisKeyName).Result()
		if err != nil {
			ErrorLog_LogError("completing TOTP enrollment", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		// verify the sample code
		validated := totp.Validate(code, secret)

		if !validated {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "bad_totp_code"})
		}

		// store the secret and remove it from redis
		_, err = DB.Exec("INSERT INTO totp(userId, secret) VALUES(?, ?)", userID, secret)
		if err != nil {
			ErrorLog_LogError("starting TOTP enrollment", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		RedisClient.Del(redisKeyName)

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/auth/2fa/unenroll", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		userID := GetSessionUserID(&c)

		if c.FormValue("code") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		code := c.FormValue("code")

		// are they enrolled already?
		enrolled, err := isUser2FAEnrolled(GetSessionUserID(&c))
		if !enrolled {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "not_enrolled"})
		}

		// get their secret
		rows, err := DB.Query("SELECT secret FROM totp WHERE userId = ?", userID)
		if err != nil {
			ErrorLog_LogError("handling TOTP unenrollment", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		secret := ""

		rows.Next()
		rows.Scan(&secret)

		// verify the sample code
		validated := totp.Validate(code, secret)

		if !validated {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "bad_totp_code"})
		}

		// remove their secret
		_, err = DB.Exec("DELETE FROM totp WHERE userID = ?", userID)
		if err != nil {
			ErrorLog_LogError("handling TOTP unenrollment", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
