package api

import (
	"net/http"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools/manager"

	"github.com/labstack/echo"
)

// responses
type PlannerWeekInfoResponse struct {
	Status        string                     `json:"status"`
	Announcements []data.PlannerAnnouncement `json:"announcements"`
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

		// TODO: merge the above into the calendar provider

		providers := []data.Provider{
			// TODO: not hardcode this for dalton
			manager.GetSchoolByID("dalton").CalendarProvider(),
		}

		for _, provider := range providers {
			providerData, err := provider.GetData(DB, &user, startDate, endDate, data.ProviderDataAnnouncements)
			if err != nil {
				ErrorLog_LogError("getting calendar provider data", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			for _, announcement := range providerData.Announcements {
				announcements = append(announcements, announcement)
			}
		}

		return c.JSON(http.StatusOK, PlannerWeekInfoResponse{"ok", announcements})
	})
}
