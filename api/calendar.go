package api

import (
	"encoding/json"
	"math"
	"net/http"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"

	"github.com/MyHomeworkSpace/api-server/calendar"

	"github.com/labstack/echo"
)

// responses
type CalendarStatusResponse struct {
	Status    string `json:"status"`
	StatusNum int    `json:"statusNum"`
}

func routeCalendarGetStatus(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if len(c.User.Schools) == 0 {
		ec.JSON(http.StatusOK, CalendarStatusResponse{"ok", 0})
		return
	}

	schools, err := data.GetSchoolsForUser(c.User)
	if err != nil {
		ErrorLog_LogError("getting calendar status", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	needsUpdate, err := schools[0].NeedsUpdate(DB)
	if err != nil {
		ErrorLog_LogError("getting calendar status", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	statusNum := 1

	if needsUpdate {
		statusNum = 2
	}

	ec.JSON(http.StatusOK, CalendarStatusResponse{"ok", statusNum})
}

func routeCalendarGetView(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("start") == "" || ec.FormValue("end") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	timeZone, err := time.LoadLocation("America/New_York")
	if err != nil {
		ErrorLog_LogError("timezone info", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	startDate, err := time.ParseInLocation("2006-01-02", ec.FormValue("start"), timeZone)
	if err != nil {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}
	endDate, err := time.ParseInLocation("2006-01-02", ec.FormValue("end"), timeZone)
	if err != nil {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}

	if int(math.Floor(endDate.Sub(startDate).Hours()/24)) > 2*365 {
		// cap of 2 years between start and end
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}

	view, err := calendar.GetView(DB, c.User, timeZone, startDate, endDate)
	if err != nil {
		ErrorLog_LogError("getting calendar view", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, CalendarViewResponse{"ok", view})
}

func routeCalendarImport(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("password") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	schoolID := "dalton"

	school, err := MainRegistry.GetSchoolByID(schoolID)
	if err != nil {
		ErrorLog_LogError("importing schedule", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	tx, err := DB.Begin()
	if err != nil {
		ErrorLog_LogError("importing schedule", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// clear any existing school record
	_, err = tx.Exec("DELETE FROM schools WHERE schoolId = ? AND userId = ?", schoolID, c.User.ID)
	if err != nil {
		tx.Rollback()

		ErrorLog_LogError("importing schedule", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// enroll
	result, err := school.Enroll(tx, c.User, map[string]interface{}{
		"username": c.User.Username,
		"password": ec.FormValue("password"),
	})
	if err != nil {
		tx.Rollback()

		schoolError, ok := err.(data.SchoolError)
		if ok {
			// it wants to report an error code
			ec.JSON(http.StatusOK, ErrorResponse{"error", schoolError.Code})
			return
		}

		// server error
		ErrorLog_LogError("importing schedule", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	dataBytes, err := json.Marshal(result)
	if err != nil {
		tx.Rollback()

		ErrorLog_LogError("importing schedule", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	dataString := string(dataBytes)

	// save the new data
	_, err = tx.Exec("INSERT INTO schools(schoolId, data, userId) VALUES(?, ?, ?)", schoolID, dataString, c.User.ID)
	if err != nil {
		tx.Rollback()

		ErrorLog_LogError("importing schedule", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// go!
	err = tx.Commit()
	if err != nil {
		ErrorLog_LogError("importing schedule", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeCalendarResetSchedule(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	schoolID := "dalton"

	tx, err := DB.Begin()
	if err != nil {
		ErrorLog_LogError("clearing schedule from DB", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	schools, err := data.GetSchoolsForUser(c.User)
	if err != nil {
		tx.Rollback()

		ErrorLog_LogError("clearing schedule from DB", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	if len(schools) != 1 {
		tx.Rollback()

		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "no_schools"})
		return
	}

	err = schools[0].Unenroll(tx, c.User)
	if err != nil {
		tx.Rollback()

		ErrorLog_LogError("clearing schedule from DB", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	_, err = tx.Exec("DELETE FROM schools WHERE schoolId = ? AND userId = ?", schoolID, c.User.ID)
	if err != nil {
		tx.Rollback()

		ErrorLog_LogError("clearing schedule from DB", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	err = tx.Commit()
	if err != nil {
		ErrorLog_LogError("clearing schedule from DB", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}
