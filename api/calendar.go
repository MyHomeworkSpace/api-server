package api

import (
	"math"
	"net/http"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"

	"github.com/MyHomeworkSpace/api-server/calendar"

	"github.com/julienschmidt/httprouter"
)

// responses
type calendarStatusResponse struct {
	Status    string `json:"status"`
	StatusNum int    `json:"statusNum"`
}
type calendarViewResponse struct {
	Status string        `json:"status"`
	View   calendar.View `json:"view"`
}

func routeCalendarGetStatus(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	schools, err := data.GetSchoolsForUser(c.User)
	if err != nil {
		errorlog.LogError("getting calendar status", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	if len(schools) == 0 {
		writeJSON(w, http.StatusOK, calendarStatusResponse{"ok", 0})
		return
	}

	needsUpdate, err := schools[0].NeedsUpdate(DB)
	if err != nil {
		errorlog.LogError("getting calendar status", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	statusNum := 1

	if needsUpdate {
		statusNum = 2
	}

	writeJSON(w, http.StatusOK, calendarStatusResponse{"ok", statusNum})
}

func routeCalendarGetView(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("start") == "" || r.FormValue("end") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	timeZone, err := time.LoadLocation("America/New_York")
	if err != nil {
		errorlog.LogError("timezone info", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	startDate, err := time.ParseInLocation("2006-01-02", r.FormValue("start"), timeZone)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}
	endDate, err := time.ParseInLocation("2006-01-02", r.FormValue("end"), timeZone)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	if int(math.Floor(endDate.Sub(startDate).Hours()/24)) > 2*365 {
		// cap of 2 years between start and end
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	view, err := calendar.GetView(DB, c.User, timeZone, startDate, endDate)
	if err != nil {
		errorlog.LogError("getting calendar view", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, calendarViewResponse{"ok", view})
}
