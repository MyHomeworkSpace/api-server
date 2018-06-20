package api

import (
	"net/http"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"

	"github.com/labstack/echo"
)

// responses
type MultiPlannerAnnouncementResponse struct {
	Status        string                     `json:"status"`
	Announcements []data.PlannerAnnouncement `json:"announcements"`
}

type PlannerFridayResponse struct {
	Status string             `json:"status"`
	Friday data.PlannerFriday `json:"friday"`
}

type PlannerWeekInfoResponse struct {
	Status        string                     `json:"status"`
	Announcements []data.PlannerAnnouncement `json:"announcements"`
	Friday        data.PlannerFriday         `json:"friday"`
}

func InitPlannerAPI(e *echo.Echo) {
	e.GET("/planner/getWeekInfo/:date", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		startDate, err := time.Parse("2006-01-02", c.Param("date"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}
		endDate := startDate.Add(time.Hour * 24 * 7)

		user, err := Data_GetUserByID(GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("getting planner week information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		grade, err := Data_GetUserGrade(user)
		if err != nil {
			ErrorLog_LogError("getting planner week information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		announcementsGroups := Data_GetGradeAnnouncementGroups(grade)
		announcementsGroupsSQL := Data_GetAnnouncementGroupSQL(announcementsGroups)

		announcementRows, err := DB.Query("SELECT id, date, text, grade, `type` FROM announcements WHERE date >= ? AND date < ? AND ("+announcementsGroupsSQL+")", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		if err != nil {
			ErrorLog_LogError("getting announcement information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer announcementRows.Close()
		announcements := []data.PlannerAnnouncement{}
		for announcementRows.Next() {
			resp := data.PlannerAnnouncement{-1, "", "", -1, -1}
			announcementRows.Scan(&resp.ID, &resp.Date, &resp.Text, &resp.Grade, &resp.Type)
			if resp.Type == AnnouncementType_BreakStart {
				resp.Text = "Start of " + resp.Text
			} else if resp.Type == AnnouncementType_BreakEnd {
				resp.Text = "End of " + resp.Text
			}
			announcements = append(announcements, resp)
		}

		fridayDate := startDate.Add(time.Hour * 24 * 4)

		fridayRows, err := DB.Query("SELECT * FROM fridays WHERE date = ?", fridayDate)
		if err != nil {
			ErrorLog_LogError("getting friday information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer fridayRows.Close()
		friday := data.PlannerFriday{-1, "", -1}
		if fridayRows.Next() {
			fridayRows.Scan(&friday.ID, &friday.Date, &friday.Index)
		}

		return c.JSON(http.StatusOK, PlannerWeekInfoResponse{"ok", announcements, friday})
	})

	e.GET("/planner/announcements/getWeek/:date", func(c echo.Context) error {
		startDate, err := time.Parse("2006-01-02", c.Param("date"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}
		endDate := startDate.Add(time.Hour * 24 * 7)
		rows, err := DB.Query("SELECT id, date, text, grade, `type` FROM announcements WHERE date >= ? AND date < ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		if err != nil {
			ErrorLog_LogError("getting announcement information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		announcements := []data.PlannerAnnouncement{}
		for rows.Next() {
			resp := data.PlannerAnnouncement{-1, "", "", -1, -1}
			rows.Scan(&resp.ID, &resp.Date, &resp.Text, &resp.Grade, &resp.Type)
			announcements = append(announcements, resp)
		}
		return c.JSON(http.StatusOK, MultiPlannerAnnouncementResponse{"ok", announcements})
	})

	e.GET("/planner/fridays/get/:date", func(c echo.Context) error {
		rows, err := DB.Query("SELECT * FROM fridays WHERE date = ?", c.Param("date"))
		if err != nil {
			ErrorLog_LogError("getting friday information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		if rows.Next() {
			resp := data.PlannerFriday{-1, "", -1}
			rows.Scan(&resp.ID, &resp.Date, &resp.Index)
			return c.JSON(http.StatusOK, PlannerFridayResponse{"ok", resp})
		} else {
			return c.JSON(http.StatusOK, StatusResponse{"ok"})
		}
	})
}
