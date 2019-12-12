package api

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"net/http"
	"time"

	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/labstack/echo"
	"github.com/pquerna/otp/totp"
)

type enrollmentResponse struct {
	Status   string `json:"status"`
	Enrolled bool   `json:"enrolled"`
}

type totpSecretResponse struct {
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
	// are they enrolled already?
	enrolled, err := isUser2FAEnrolled(c.User.ID)
	if err != nil {
		errorlog.LogError("starting TOTP enrollment", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	if enrolled {
		writeJSON(w, http.StatusUnauthorized, errorResponse{"error", "already_enrolled"})
		return
	}

	// generate a new secret, store in redis
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "MyHomeworkSpace",
		AccountName: c.User.Email,
	})

	secret := key.Secret()
	redisKeyName := fmt.Sprintf("user:%d:totp_tmp", c.User.ID)

	redisResponse := RedisClient.Set(redisKeyName, secret, time.Hour)
	if redisResponse.Err() != nil {
		errorlog.LogError("starting TOTP enrollment", redisResponse.Err())
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// generate qr code and data url
	image, err := key.Image(200, 200)
	if err != nil {
		errorlog.LogError("starting TOTP enrollment", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	var pngImage bytes.Buffer

	err = png.Encode(&pngImage, image)
	if err != nil {
		errorlog.LogError("starting TOTP enrollment", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	encodedString := base64.StdEncoding.EncodeToString(pngImage.Bytes())

	imageURL := "data:image/png;base64," + encodedString

	writeJSON(w, http.StatusOK, totpSecretResponse{"ok", secret, imageURL})
}

func routeAuth2faCompleteEnroll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if r.FormValue("code") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	code := r.FormValue("code")

	// are they enrolled already?
	enrolled, err := isUser2FAEnrolled(c.User.ID)
	if err != nil {
		errorlog.LogError("completing TOTP enrollment", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	if enrolled {
		writeJSON(w, http.StatusUnauthorized, errorResponse{"error", "already_enrolled"})
		return
	}

	// do they have a stored secret in redis?
	redisKeyName := fmt.Sprintf("user:%d:totp_tmp", c.User.ID)
	secret, err := RedisClient.Get(redisKeyName).Result()
	if err != nil {
		errorlog.LogError("completing TOTP enrollment", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// verify the sample code
	validated := totp.Validate(code, secret)

	if !validated {
		writeJSON(w, http.StatusUnauthorized, errorResponse{"error", "bad_totp_code"})
		return
	}

	// store the secret and remove it from redis
	_, err = DB.Exec("INSERT INTO totp(userId, secret) VALUES(?, ?)", c.User.ID, secret)
	if err != nil {
		errorlog.LogError("starting TOTP enrollment", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	RedisClient.Del(redisKeyName)

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeAuth2faStatus(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	enrolled, err := isUser2FAEnrolled(c.User.ID)
	if err != nil {
		errorlog.LogError("getting TOTP enrollment status", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, enrollmentResponse{"ok", enrolled})
}

func routeAuth2faUnenroll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	userID := c.User.ID

	if r.FormValue("code") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	code := r.FormValue("code")

	// are they enrolled already?
	enrolled, err := isUser2FAEnrolled(c.User.ID)
	if err != nil {
		errorlog.LogError("handling TOTP unenrollment", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	if !enrolled {
		writeJSON(w, http.StatusUnauthorized, errorResponse{"error", "not_enrolled"})
		return
	}

	// get their secret
	rows, err := DB.Query("SELECT secret FROM totp WHERE userId = ?", userID)
	if err != nil {
		errorlog.LogError("handling TOTP unenrollment", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	secret := ""

	rows.Next()
	rows.Scan(&secret)

	// verify the sample code
	validated := totp.Validate(code, secret)

	if !validated {
		writeJSON(w, http.StatusUnauthorized, errorResponse{"error", "bad_totp_code"})
		return
	}

	// remove their secret
	_, err = DB.Exec("DELETE FROM totp WHERE userID = ?", userID)
	if err != nil {
		errorlog.LogError("handling TOTP unenrollment", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}
