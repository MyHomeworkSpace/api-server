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

	"github.com/julienschmidt/httprouter"
)

type feedbacksResponse struct {
	Status    string          `json:"status"`
	Feedbacks []data.Feedback `json:"feedbacks"`
}

type userCountResponse struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

func routeAdminGetAllFeedback(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	rows, err := DB.Query("SELECT feedback.id, feedback.userId, feedback.type, feedback.text, feedback.screenshot, feedback.timestamp, feedback.userAgent, users.name, users.email FROM feedback INNER JOIN users ON feedback.userId = users.id")
	if err != nil {
		errorlog.LogError("getting all feedback", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	feedbacks := []data.Feedback{}

	defer rows.Close()

	for rows.Next() {
		resp := data.Feedback{}
		var screenshot string
		err = rows.Scan(&resp.ID, &resp.UserID, &resp.Type, &resp.Text, &screenshot, &resp.Timestamp, &resp.UserAgent, &resp.UserName, &resp.UserEmail)
		if err != nil {
			errorlog.LogError("scanning feedback", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return

		}
		if screenshot != "" {
			resp.HasScreenshot = true
		}
		feedbacks = append(feedbacks, resp)
	}

	writeJSON(w, http.StatusOK, feedbacksResponse{"ok", feedbacks})
}

func routeAdminGetFeedbackScreenshot(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	id, err := strconv.Atoi(p.ByName("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	rows, err := DB.Query("SELECT screenshot FROM feedback WHERE id = ?", id)
	if err != nil {
		errorlog.LogError("getting screenshot", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	if !rows.Next() {
		writeJSON(w, http.StatusNotFound, errorResponse{"error", "not_found"})
		return
	}

	var screenshot64 string
	err = rows.Scan(&screenshot64)
	if err != nil {
		errorlog.LogError("getting screenshot", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	rows.Close()

	if screenshot64 == "" {
		writeJSON(w, http.StatusNotFound, errorResponse{"error", "not_found"})
		return
	}

	screenshotEncodedBytes := []byte(strings.Replace(screenshot64, "data:image/png;base64,", "", 1))
	screenshotDecodedBytes := make([]byte, base64.StdEncoding.DecodedLen(len(screenshotEncodedBytes)))

	_, err = base64.StdEncoding.Decode(screenshotDecodedBytes, screenshotEncodedBytes)
	if err != nil {
		errorlog.LogError("getting screenshot", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "image/png")
	_, err = w.Write(screenshotDecodedBytes)
	if err != nil {
		errorlog.LogError("getting screenshot", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
}

func routeAdminGetUserCount(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	rows, err := DB.Query("SELECT COUNT(*) FROM users")
	if err != nil {
		errorlog.LogError("getting user count", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	rows.Next()
	count := -1
	rows.Scan(&count)

	writeJSON(w, http.StatusOK, userCountResponse{"ok", count})
}

func routeAdminSendEmail(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("template") == "" || r.FormValue("data") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	user := c.User

	if r.FormValue("userID") != "" {
		userID, err := strconv.Atoi(r.FormValue("userID"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
			return
		}

		userStruct, err := data.GetUserByID(userID)
		if err == data.ErrNotFound {
			writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
			return
		} else if err != nil {
			errorlog.LogError("sending email", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		user = &userStruct
	}

	data := map[string]interface{}{}
	err := json.Unmarshal([]byte(r.FormValue("data")), &data)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	err = email.Send("", user, r.FormValue("template"), data)
	if err != nil {
		errorlog.LogError("sending email", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeAdminTriggerError(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	errorlog.LogError("intentionally triggered error", errors.New("api: intentionally triggered error"))
	writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
}
