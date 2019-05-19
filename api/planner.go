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

		// TODO: merge the above into the calendar provider

		grade, err := Data_GetUserGrade(user)
		if err != nil {
			ErrorLog_LogError("getting planner week information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		announcementsGroups := Data_GetGradeAnnouncementGroups(grade)
		announcementsGroupsSQL := Data_GetAnnouncementGroupSQL(announcementsGroups)

		providers := []data.Provider{
			// TODO: not hardcode this for dalton
			manager.GetSchoolByID("dalton").CalendarProvider(),
		}

		announcements := []data.PlannerAnnouncement{}

		for _, provider := range providers {
			providerData, err := provider.GetData(DB, &user, time.UTC, grade, announcementsGroupsSQL, startDate, endDate, data.ProviderDataAnnouncements)
			if err != nil {
				ErrorLog_LogError("getting calendar provider data", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			announcements = append(announcements, providerData.Announcements...)
		}

		return c.JSON(http.StatusOK, PlannerWeekInfoResponse{"ok", announcements})
	})
}
