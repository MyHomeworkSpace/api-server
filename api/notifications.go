package api

import (
	"net/http"
	"strconv"

	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/labstack/echo"
)

type Notification struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	Expiry  string `json:"expiry"`
}

type notificationsResponse struct {
	Status        string         `json:"status"`
	Notifications []Notification `json:"notifications"`
}

func routeNotificationsAdd(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("expiry") == "" || ec.FormValue("content") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	_, err := DB.Exec("INSERT INTO notifications (content, expiry) VALUES (?, ?)", ec.FormValue("content"), ec.FormValue("expiry"))
	if err != nil {
		errorlog.LogError("adding notification", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeNotificationsDelete(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	idStr := ec.FormValue("id")
	if idStr == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	_, err = DB.Exec("DELETE FROM notifications WHERE id = ?", id)
	if err != nil {
		errorlog.LogError("deleting notification", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeNotificationsGet(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT `id`, `content`, `expiry` FROM notifications WHERE expiry > NOW()")
	if err != nil {
		errorlog.LogError("getting notification", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	notifications := []Notification{}
	for rows.Next() {
		resp := Notification{-1, "", ""}
		rows.Scan(&resp.ID, &resp.Content, &resp.Expiry)
		notifications = append(notifications, resp)
	}

	writeJSON(w, http.StatusOK, notificationsResponse{"ok", notifications})
}
