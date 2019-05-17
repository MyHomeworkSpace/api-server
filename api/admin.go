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

func routeAdminGetAllFeedback(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}
	user, _ := Data_GetUserByID(GetSessionUserID(&ec))
	if user.Level < 1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
		return
	}

	rows, err := DB.Query("SELECT feedback.id, feedback.userId, feedback.type, feedback.text, feedback.screenshot, feedback.timestamp, users.name, users.email FROM feedback INNER JOIN users ON feedback.userId = users.id")
	if err != nil {
		ErrorLog_LogError("getting all feedback", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
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

	ec.JSON(http.StatusOK, FeedbacksResponse{"ok", feedbacks})
}

func routeAdminGetFeedbackScreenshot(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}
	user, _ := Data_GetUserByID(GetSessionUserID(&ec))
	if user.Level < 1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
		return
	}

	id, err := strconv.Atoi(ec.Param("id"))
	if err != nil {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_paramas"})
		return
	}

	rows, err := DB.Query("SELECT screenshot FROM feedback WHERE id = ?", id)
	if err != nil {
		ErrorLog_LogError("getting screenshot", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	if !rows.Next() {
		ec.JSON(http.StatusNotFound, ErrorResponse{"error", "not_found"})
		return
	}

	var screenshot64 string
	err = rows.Scan(&screenshot64)
	if err != nil {
		ErrorLog_LogError("getting screenshot", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	rows.Close()

	if screenshot64 == "" {
		ec.JSON(http.StatusNotFound, ErrorResponse{"error", "no_screenshot"})
		return
	}

	screenshot64 = strings.Replace(screenshot64, "data:image/png;base64,", "", 1)

	screenshot, err := base64.StdEncoding.DecodeString(screenshot64)
	if err != nil {
		ErrorLog_LogError("getting screenshot", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.Blob(http.StatusOK, "image/png;base64", screenshot)
}

func routeAdminGetUserCount(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}
	user, _ := Data_GetUserByID(GetSessionUserID(&ec))
	if user.Level < 1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
		return
	}

	rows, err := DB.Query("SELECT COUNT(*) FROM users")
	if err != nil {
		ErrorLog_LogError("getting user count", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	rows.Next()
	count := -1
	rows.Scan(&count)

	ec.JSON(http.StatusOK, UserCountResponse{"ok", count})
}

func InitAdminAPI(e *echo.Echo) {
	e.GET("/admin/getAllFeedback", Route(routeAdminGetAllFeedback))
	e.GET("/admin/getFeedbackScreenshot/:id", Route(routeAdminGetFeedbackScreenshot))
	e.GET("/admin/getUserCount", Route(routeAdminGetUserCount))
}
