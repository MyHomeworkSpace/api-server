package api

import (
	"net/http"

	"github.com/labstack/echo"
)

type Notification struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
	Expiry  string `json:"expiry"`
}

type NotificationResponse struct {
	Status                string         `json:"status"`
	ReturnedNotifications []Notification `json:"notifications"`
}

func InitAnnoucncementsAPI(e *echo.Echo) {
	e.GET("/notifications/get", func(c echo.Context) error {
		rows, err := DB.Query("SELECT `id`, `content`, `expiry` FROM notifications WHERE expiry > NOW()")
		if err != nil {
			ErrorLog_LogError("getting annoucement", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		notifications := []Notification{}
		for rows.Next() {
			resp := Notification{-1, "", ""}
			rows.Scan(&resp.ID, &resp.Content, &resp.Expiry)
			notifications = append(notifications, resp)
		}
		return c.JSON(http.StatusOK, NotificationResponse{"ok", notifications})
	})

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
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
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
		if c.FormValue("id") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}
		_, err := DB.Exec("DELETE FROM notifications WHERE id = ?", c.FormValue("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
