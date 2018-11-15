package api

import (
	"net/http"

	"github.com/labstack/echo"
)

type UsersResponse struct {
	Status string `json:"status"`
	Users  []User `json:"users"`
}

type FeedbacksResponse struct {
	Status    string     `json:"status"`
	Feedbacks []Feedback `json:"feedbacks"`
}

func InitAdminAPI(e *echo.Echo) {
	e.GET("/admin/getAllUsers", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		user, _ := Data_GetUserByID(GetSessionUserID(&c))
		if user.Level < 1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
		}

		rows, err := DB.Query("SELECT ID, name, username, email, type, features, level, showMigrateMessage FROM users")
		if err != nil {
			ErrorLog_LogError("getting all users", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		users := []User{}
		for rows.Next() {
			resp := User{-1, "", "", "", "", "", -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Username, &resp.Email, &resp.Type, &resp.Features, &resp.Level, &resp.ShowMigrateMessage)
			users = append(users, resp)
		}

		return c.JSON(http.StatusOK, UsersResponse{"ok", users})
	})

	e.GET("/admin/getAllFeedback", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		user, _ := Data_GetUserByID(GetSessionUserID(&c))
		if user.Level < 1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
		}

		rows, err := DB.Query("SELECT feedback.id, feedback.userId, feedback.type, feedback.text, feedback.timestamp, users.name, users.email FROM feedback INNER JOIN users ON feedback.userId = users.id")
		if err != nil {
			ErrorLog_LogError("getting all feedback", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		feedbacks := []Feedback{}
		for rows.Next() {
			resp := Feedback{-1, -1, "", "", "", "", ""}
			rows.Scan(&resp.ID, &resp.UserID, &resp.Type, &resp.Text, &resp.Timestamp, &resp.UserName, &resp.UserEmail)
			feedbacks = append(feedbacks, resp)
		}

		return c.JSON(http.StatusOK, FeedbacksResponse{"ok", feedbacks})
	})
}
