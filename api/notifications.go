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
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	_, err := DB.Exec("INSERT INTO notifications (content, expiry) VALUES (?, ?)", ec.FormValue("content"), ec.FormValue("expiry"))
	if err != nil {
		errorlog.LogError("adding notification", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeNotificationsDelete(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	idStr := ec.FormValue("id")
	if idStr == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	_, err = DB.Exec("DELETE FROM notifications WHERE id = ?", id)
	if err != nil {
		errorlog.LogError("deleting notification", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeNotificationsGet(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT `id`, `content`, `expiry` FROM notifications WHERE expiry > NOW()")
	if err != nil {
		errorlog.LogError("getting notification", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	notifications := []Notification{}
	for rows.Next() {
		resp := Notification{-1, "", ""}
		rows.Scan(&resp.ID, &resp.Content, &resp.Expiry)
		notifications = append(notifications, resp)
	}

	ec.JSON(http.StatusOK, notificationsResponse{"ok", notifications})
}
