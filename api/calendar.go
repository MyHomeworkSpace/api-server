package api

import (
	"math"
	"net/http"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"

	"github.com/MyHomeworkSpace/api-server/calendar"

	"github.com/labstack/echo"
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

func routeCalendarGetStatus(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if len(c.User.Schools) == 0 {
		ec.JSON(http.StatusOK, calendarStatusResponse{"ok", 0})
		return
	}

	schools, err := data.GetSchoolsForUser(c.User)
	if err != nil {
		errorlog.LogError("getting calendar status", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	needsUpdate, err := schools[0].NeedsUpdate(DB)
	if err != nil {
		errorlog.LogError("getting calendar status", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	statusNum := 1

	if needsUpdate {
		statusNum = 2
	}

	ec.JSON(http.StatusOK, calendarStatusResponse{"ok", statusNum})
}

func routeCalendarGetView(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("start") == "" || ec.FormValue("end") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	timeZone, err := time.LoadLocation("America/New_York")
	if err != nil {
		errorlog.LogError("timezone info", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	startDate, err := time.ParseInLocation("2006-01-02", ec.FormValue("start"), timeZone)
	if err != nil {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}
	endDate, err := time.ParseInLocation("2006-01-02", ec.FormValue("end"), timeZone)
	if err != nil {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	if int(math.Floor(endDate.Sub(startDate).Hours()/24)) > 2*365 {
		// cap of 2 years between start and end
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	view, err := calendar.GetView(DB, c.User, timeZone, startDate, endDate)
	if err != nil {
		errorlog.LogError("getting calendar view", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, calendarViewResponse{"ok", view})
}
