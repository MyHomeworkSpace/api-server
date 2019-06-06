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

type EnrollmentResponse struct {
	Status   string `json:"status"`
	Enrolled bool   `json:"enrolled"`
}

type TOTPSecretResponse struct {
	Status   string `json:"status"`
	Secret   string `json:"secret"`
	ImageURL string `json:"imageURL"`
}

/*
 * helpers
 */

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

/*
 * routes
 */

func routeAuth2faBeginEnroll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	// are they enrolled already?
	enrolled, err := isUser2FAEnrolled(GetSessionUserID(&ec))
	if err != nil {
		ErrorLog_LogError("starting TOTP enrollment", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	if enrolled {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "already_enrolled"})
		return
	}

	// get their email
	user, err := Data_GetUserByID(GetSessionUserID(&ec))
	if err != nil {
		ErrorLog_LogError("starting TOTP enrollment", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
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
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// generate qr code and data url
	image, err := key.Image(200, 200)
	if err != nil {
		ErrorLog_LogError("starting TOTP enrollment", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	var pngImage bytes.Buffer

	err = png.Encode(&pngImage, image)
	if err != nil {
		ErrorLog_LogError("starting TOTP enrollment", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	encodedString := base64.StdEncoding.EncodeToString(pngImage.Bytes())

	imageURL := "data:image/png;base64," + encodedString

	ec.JSON(http.StatusOK, TOTPSecretResponse{"ok", secret, imageURL})
}

func routeAuth2faCompleteEnroll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	userID := GetSessionUserID(&ec)

	if ec.FormValue("code") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	code := ec.FormValue("code")

	// are they enrolled already?
	enrolled, err := isUser2FAEnrolled(GetSessionUserID(&ec))
	if err != nil {
		ErrorLog_LogError("completing TOTP enrollment", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	if enrolled {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "already_enrolled"})
		return
	}

	// do they have a stored secret in redis?
	redisKeyName := fmt.Sprintf("user:%d:totp_tmp", userID)
	secret, err := RedisClient.Get(redisKeyName).Result()
	if err != nil {
		ErrorLog_LogError("completing TOTP enrollment", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// verify the sample code
	validated := totp.Validate(code, secret)

	if !validated {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "bad_totp_code"})
		return
	}

	// store the secret and remove it from redis
	_, err = DB.Exec("INSERT INTO totp(userId, secret) VALUES(?, ?)", userID, secret)
	if err != nil {
		ErrorLog_LogError("starting TOTP enrollment", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	RedisClient.Del(redisKeyName)

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeAuth2faStatus(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	enrolled, err := isUser2FAEnrolled(GetSessionUserID(&ec))
	if err != nil {
		ErrorLog_LogError("getting TOTP enrollment status", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, EnrollmentResponse{"ok", enrolled})
}

func routeAuth2faUnenroll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	userID := GetSessionUserID(&ec)

	if ec.FormValue("code") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	code := ec.FormValue("code")

	// are they enrolled already?
	enrolled, err := isUser2FAEnrolled(GetSessionUserID(&ec))
	if err != nil {
		ErrorLog_LogError("handling TOTP unenrollment", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	if !enrolled {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "not_enrolled"})
		return
	}

	// get their secret
	rows, err := DB.Query("SELECT secret FROM totp WHERE userId = ?", userID)
	if err != nil {
		ErrorLog_LogError("handling TOTP unenrollment", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	secret := ""

	rows.Next()
	rows.Scan(&secret)

	// verify the sample code
	validated := totp.Validate(code, secret)

	if !validated {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "bad_totp_code"})
		return
	}

	// remove their secret
	_, err = DB.Exec("DELETE FROM totp WHERE userID = ?", userID)
	if err != nil {
		ErrorLog_LogError("handling TOTP unenrollment", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}
