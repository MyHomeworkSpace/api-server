package api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/email"
	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/labstack/echo"
)

type feedbacksResponse struct {
	Status    string          `json:"status"`
	Feedbacks []data.Feedback `json:"feedbacks"`
}

type userCountResponse struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

func routeAdminGetAllFeedback(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT feedback.id, feedback.userId, feedback.type, feedback.text, feedback.screenshot, feedback.timestamp, users.name, users.email FROM feedback INNER JOIN users ON feedback.userId = users.id")
	if err != nil {
		errorlog.LogError("getting all feedback", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	feedbacks := []data.Feedback{}
	for rows.Next() {
		resp := data.Feedback{-1, -1, "", "", "", "", "", false}
		var screenshot string
		rows.Scan(&resp.ID, &resp.UserID, &resp.Type, &resp.Text, &screenshot, &resp.Timestamp, &resp.UserName, &resp.UserEmail)
		if screenshot != "" {
			resp.HasScreenshot = true
		}
		feedbacks = append(feedbacks, resp)
	}

	ec.JSON(http.StatusOK, feedbacksResponse{"ok", feedbacks})
}

func routeAdminGetFeedbackScreenshot(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	id, err := strconv.Atoi(ec.Param("id"))
	if err != nil {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_paramas"})
		return
	}

	rows, err := DB.Query("SELECT screenshot FROM feedback WHERE id = ?", id)
	if err != nil {
		errorlog.LogError("getting screenshot", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	if !rows.Next() {
		ec.JSON(http.StatusNotFound, errorResponse{"error", "not_found"})
		return
	}

	var screenshot64 string
	err = rows.Scan(&screenshot64)
	if err != nil {
		errorlog.LogError("getting screenshot", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	rows.Close()

	if screenshot64 == "" {
		ec.JSON(http.StatusNotFound, errorResponse{"error", "no_screenshot"})
		return
	}

	screenshot64 = strings.Replace(screenshot64, "data:image/png;base64,", "", 1)

	screenshot, err := base64.StdEncoding.DecodeString(screenshot64)
	if err != nil {
		errorlog.LogError("getting screenshot", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.Blob(http.StatusOK, "image/png;base64", screenshot)
}

func routeAdminGetUserCount(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT COUNT(*) FROM users")
	if err != nil {
		errorlog.LogError("getting user count", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	rows.Next()
	count := -1
	rows.Scan(&count)

	ec.JSON(http.StatusOK, userCountResponse{"ok", count})
}

func routeAdminSendEmail(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("template") == "" || ec.FormValue("data") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	user := c.User

	if ec.FormValue("userID") != "" {
		userID, err := strconv.Atoi(ec.FormValue("userID"))
		if err != nil {
			ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
			return
		}

		userStruct, err := data.GetUserByID(userID)
		if err == data.ErrNotFound {
			ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
			return
		} else if err != nil {
			errorlog.LogError("sending email", err)
			ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		user = &userStruct
	}

	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(ec.FormValue("data")), &data)
	if err != nil {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	err = email.Send("", user, ec.FormValue("template"), data)
	if err != nil {
		errorlog.LogError("sending email", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeAdminTriggerError(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	errorlog.LogError("intentionally triggered error", errors.New("api: intentionally triggered error"))
	ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
}
