package api

import (
	"net/http"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"

	"github.com/labstack/echo"
)

// responses
type PlannerWeekInfoResponse struct {
	Status        string                     `json:"status"`
	Announcements []data.PlannerAnnouncement `json:"announcements"`
}

func routePlannerGetWeekInfo(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	startDate, err := time.Parse("2006-01-02", ec.Param("date"))
	if err != nil {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}
	endDate := startDate.Add(time.Hour * 24 * 7)

	providers, err := data.GetProvidersForUser(c.User)
	if err != nil {
		ErrorLog_LogError("getting calendar providers", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	announcements := []data.PlannerAnnouncement{}

	for _, provider := range providers {
		providerData, err := provider.GetData(DB, c.User, time.UTC, startDate, endDate, data.ProviderDataAnnouncements)
		if err != nil {
			ErrorLog_LogError("getting calendar provider data", err)
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}

		announcements = append(announcements, providerData.Announcements...)
	}

	ec.JSON(http.StatusOK, PlannerWeekInfoResponse{"ok", announcements})
}
