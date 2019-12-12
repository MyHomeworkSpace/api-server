package api

import (
	"net/http"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"

	"github.com/julienschmidt/httprouter"
)

// responses
type plannerWeekInfoResponse struct {
	Status        string                     `json:"status"`
	Announcements []data.PlannerAnnouncement `json:"announcements"`
}

func routePlannerGetWeekInfo(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	startDate, err := time.Parse("2006-01-02", p.ByName("date"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}
	endDate := startDate.Add(time.Hour * 24 * 7)

	providers, err := data.GetProvidersForUser(c.User)
	if err != nil {
		errorlog.LogError("getting calendar providers", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	announcements := []data.PlannerAnnouncement{}

	for _, provider := range providers {
		providerData, err := provider.GetData(DB, c.User, time.UTC, startDate, endDate, data.ProviderDataAnnouncements)
		if err != nil {
			errorlog.LogError("getting calendar provider data", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		announcements = append(announcements, providerData.Announcements...)
	}

	writeJSON(w, http.StatusOK, plannerWeekInfoResponse{"ok", announcements})
}
