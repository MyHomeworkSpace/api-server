package api

import (
	"net/http"
	"strconv"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"

	"github.com/julienschmidt/httprouter"
)

// responses
type eventChangeResponse struct {
	Status      string            `json:"status"`
	EventChange *data.EventChange `json:"eventChange"`
}

func routeCalendarEventChangesGet(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("eventID") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	eventID := r.FormValue("eventID")

	rows, err := DB.Query("SELECT eventID, cancel, userID FROM calendar_event_changes WHERE eventID = ? AND userID = ?", eventID, c.User.ID)
	if err != nil {
		errorlog.LogError("getting calendar event change", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		writeJSON(w, http.StatusOK, eventChangeResponse{"ok", nil})
		return
	}

	eventChange := data.EventChange{}
	rows.Scan(&eventChange.EventID, &eventChange.Cancel, &eventChange.UserID)

	writeJSON(w, http.StatusOK, eventChangeResponse{"ok", &eventChange})
}

func routeCalendarEventChangesSet(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("eventID") == "" || r.FormValue("cancel") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	eventID := r.FormValue("eventID")

	cancel, err := strconv.ParseBool(r.FormValue("cancel"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	cancelInt := 0
	if cancel {
		cancelInt = 1
	}

	rows, err := DB.Query("SELECT eventID FROM calendar_event_changes WHERE eventID = ? AND userID = ?", eventID, c.User.ID)
	if err != nil {
		errorlog.LogError("getting calendar event change", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		// doesn't exist, add it
		_, err = DB.Exec(
			"INSERT INTO calendar_event_changes(eventID, cancel, userID) VALUES(?, ?, ?)",
			eventID, cancelInt, c.User.ID,
		)
		if err != nil {
			errorlog.LogError("inserting calendar event change", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}
	} else {
		// exists already, update it
		_, err = DB.Exec(
			"UPDATE calendar_event_changes SET cancel = ? WHERE eventID = ? AND userID = ?",
			cancelInt, eventID, c.User.ID,
		)
		if err != nil {
			errorlog.LogError("updating calendar event change", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}
