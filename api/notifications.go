package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"
)

type Notification struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	Expiry  string `json:"expiry"`
}

type NotificationsResponse struct {
	Status        string         `json:"status"`
	Notifications []Notification `json:"notifications"`
}

func InitNotificationsAPI(e *echo.Echo) {
	e.POST("/notifications/add", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		user, _ := Data_GetUserByID(GetSessionUserID(&c))
		if user.Level < 1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
		}

		if c.FormValue("expiry") == "" || c.FormValue("content") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		_, err := DB.Exec("INSERT INTO notifications (content, expiry) VALUES (?, ?)", c.FormValue("content"), c.FormValue("expiry"))
		if err != nil {
			ErrorLog_LogError("adding notification", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/notifications/delete", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		user, _ := Data_GetUserByID(GetSessionUserID(&c))
		if user.Level < 1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
		}

		idStr := c.FormValue("id")
		if idStr == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		_, err = DB.Exec("DELETE FROM notifications WHERE id = ?", id)
		if err != nil {
			ErrorLog_LogError("deleting notification", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.GET("/notifications/get", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		rows, err := DB.Query("SELECT `id`, `content`, `expiry` FROM notifications WHERE expiry > NOW()")
		if err != nil {
			ErrorLog_LogError("getting notification", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		notifications := []Notification{}
		for rows.Next() {
			resp := Notification{-1, "", ""}
			rows.Scan(&resp.ID, &resp.Content, &resp.Expiry)
			notifications = append(notifications, resp)
		}

		return c.JSON(http.StatusOK, NotificationsResponse{"ok", notifications})
	})
}
