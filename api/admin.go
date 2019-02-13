package api

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
)

type FeedbacksResponse struct {
	Status    string     `json:"status"`
	Feedbacks []Feedback `json:"feedbacks"`
}

type UserCountResponse struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

func InitAdminAPI(e *echo.Echo) {
	e.GET("/admin/getAllFeedback", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		user, _ := Data_GetUserByID(GetSessionUserID(&c))
		if user.Level < 1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
		}

		rows, err := DB.Query("SELECT feedback.id, feedback.userId, feedback.type, feedback.text, feedback.screenshot, feedback.timestamp, users.name, users.email FROM feedback INNER JOIN users ON feedback.userId = users.id")
		if err != nil {
			ErrorLog_LogError("getting all feedback", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		feedbacks := []Feedback{}
		for rows.Next() {
			resp := Feedback{-1, -1, "", "", "", "", "", false}
			var screenshot string
			rows.Scan(&resp.ID, &resp.UserID, &resp.Type, &resp.Text, &screenshot, &resp.Timestamp, &resp.UserName, &resp.UserEmail)
			if screenshot != "" {
				resp.HasScreenshot = true
			}
			feedbacks = append(feedbacks, resp)
		}

		return c.JSON(http.StatusOK, FeedbacksResponse{"ok", feedbacks})
	})

	e.GET("/admin/getFeedbackScreenshot/:id", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		user, _ := Data_GetUserByID(GetSessionUserID(&c))
		if user.Level < 1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
		}

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_paramas"})
		}

		rows, err := DB.Query("SELECT screenshot FROM feedback WHERE id = ?", id)
		if err != nil {
			ErrorLog_LogError("getting screenshot", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		if !rows.Next() {
			return c.JSON(http.StatusNotFound, ErrorResponse{"error", "not_found"})
		}

		var screenshot64 string
		err = rows.Scan(&screenshot64)
		if err != nil {
			ErrorLog_LogError("getting screenshot", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		rows.Close()

		if screenshot64 == "" {
			return c.JSON(http.StatusNotFound, ErrorResponse{"error", "no_screenshot"})
		}

		screenshot64 = strings.Replace(screenshot64, "data:image/png;base64,", "", 1)

		screenshot, err := base64.StdEncoding.DecodeString(screenshot64)
		if err != nil {
			ErrorLog_LogError("getting screenshot", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.Blob(http.StatusOK, "image/png;base64", screenshot)
	})

	e.GET("/admin/getUserCount", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		user, _ := Data_GetUserByID(GetSessionUserID(&c))
		if user.Level < 1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
		}

		rows, err := DB.Query("SELECT COUNT(*) FROM users")
		if err != nil {
			ErrorLog_LogError("getting user count", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		rows.Next()
		count := -1
		rows.Scan(&count)

		return c.JSON(http.StatusOK, UserCountResponse{"ok", count})
	})
}
