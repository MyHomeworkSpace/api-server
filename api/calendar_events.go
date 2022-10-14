package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/MyHomeworkSpace/api-server/calendar"
	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"

	"github.com/julienschmidt/httprouter"
)

// structs for data
type CalendarEvent struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Start    int    `json:"start"`
	End      int    `json:"end"`
	Location string `json:"location"`
	Desc     string `json:"desc"`
	UserID   int    `json:"userId"`
}
type CalendarHWEvent struct {
	ID       int           `json:"id"`
	Homework data.Homework `json:"homework"`
	Start    int           `json:"start"`
	End      int           `json:"end"`
	Location string        `json:"location"`
	Desc     string        `json:"desc"`
	UserID   int           `json:"userId"`
}
type CalendarScheduleItem struct {
	ID           int    `json:"id"`
	TermID       int    `json:"termId"`
	ClassID      int    `json:"classId"`
	Name         string `json:"name"`
	OwnerID      int    `json:"ownerId"`
	OwnerName    string `json:"ownerName"`
	DayNumber    int    `json:"dayNumber"`
	Block        string `json:"block"`
	BuildingName string `json:"buildingName"`
	RoomNumber   string `json:"roomNumber"`
	Start        int    `json:"start"`
	End          int    `json:"end"`
	UserID       int    `json:"userId"`
}

// responses
type calendarWeekResponse struct {
	Status         string                     `json:"status"`
	Announcements  []data.PlannerAnnouncement `json:"announcements"`
	Events         []CalendarEvent            `json:"events"`
	HWEvents       []CalendarHWEvent          `json:"hwEvents"`
	ScheduleEvents [][]CalendarScheduleItem   `json:"scheduleEvents"`
}

/*
 * helpers
 */

func parseRecurFormInfo(r *http.Request) (bool, int, int, string, string) {
	if r.FormValue("recur") != "" {
		recurStr := r.FormValue("recur")
		recur, err := strconv.ParseBool(recurStr)
		if err != nil {
			return false, 0, 0, "", "invalid_params"
		}

		if !recur {
			return false, 0, 0, "", ""
		}

		if r.FormValue("recurFrequency") == "" || r.FormValue("recurInterval") == "" {
			return false, 0, 0, "", "missing_params"
		}

		recurFrequency, err := strconv.Atoi(r.FormValue("recurFrequency"))
		recurInterval, err1 := strconv.Atoi(r.FormValue("recurInterval"))
		recurUntil := ""
		if r.FormValue("recurUntil") != "" {
			_, err2 := time.Parse("2006-01-02", r.FormValue("recurUntil"))
			recurUntil = r.FormValue("recurUntil")
			if err2 != nil {
				return false, 0, 0, "", "invalid_params"
			}
		} else {
			// fill in a placeholder value because mysql wants one
			recurUntil = "2099-12-12"
		}

		if err != nil || err1 != nil {
			return false, 0, 0, "", "invalid_params"
		}

		if recurInterval < 1 {
			return false, 0, 0, "", "invalid_params"
		}

		return true, recurFrequency, recurInterval, recurUntil, ""
	}

	return false, 0, 0, "", ""
}

/*
 * routes
 */

func routeCalendarEventsGetWeek(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	startDate, err := time.Parse("2006-01-02", p.ByName("monday"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}
	endDate := startDate.Add(time.Hour * 24 * 7)

	view, err := calendar.GetView(DB, c.User, time.UTC, startDate, endDate)
	if err != nil {
		errorlog.LogError("getting calendar week", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	announcements := []data.PlannerAnnouncement{}
	plainEvents := []CalendarEvent{}
	hwEvents := []CalendarHWEvent{}
	var scheduleEvents [][]CalendarScheduleItem

	scheduleEvents = make([][]CalendarScheduleItem, 5)

	for dayIndex, day := range view.Days {
		for _, announcement := range day.Announcements {
			announcements = append(announcements, announcement)
		}
		for _, event := range day.Events {
			descriptionInterface, hasDescription := event.Tags[data.EventTagDescription]
			locationInterface, hasLocation := event.Tags[data.EventTagLocation]
			homeworkInterface, isHomework := event.Tags[data.EventTagHomework]
			_, isSchedule := event.Tags[data.EventTagClassID]
			if isHomework {
				homework := homeworkInterface.(data.Homework)
				hwEvent := CalendarHWEvent{
					ID:       event.ID,
					Homework: homework,
					Start:    event.Start,
					End:      event.End,
					UserID:   event.UserID,
				}
				if hasLocation {
					hwEvent.Location = locationInterface.(string)
				}
				if hasDescription {
					hwEvent.Desc = descriptionInterface.(string)
				}
				hwEvents = append(hwEvents, hwEvent)
			} else if isSchedule {
				dayTime, _ := time.Parse("2006-01-02", day.DayString)
				scheduleEvent := CalendarScheduleItem{
					ID:           event.ID,
					TermID:       event.Tags[data.EventTagTermID].(int),
					ClassID:      event.Tags[data.EventTagClassID].(int),
					Name:         event.Name,
					OwnerID:      event.Tags[data.EventTagOwnerID].(int),
					OwnerName:    event.Tags[data.EventTagOwnerName].(string),
					DayNumber:    event.Tags[data.EventTagDayNumber].(int),
					Block:        event.Tags[data.EventTagBlock].(string),
					BuildingName: event.Tags[data.EventTagBuildingName].(string),
					RoomNumber:   event.Tags[data.EventTagRoomNumber].(string),
					Start:        event.Start - int(dayTime.Unix()),
					End:          event.End - int(dayTime.Unix()),
					UserID:       event.UserID,
				}
				scheduleEvents[dayIndex] = append(scheduleEvents[dayIndex], scheduleEvent)
			} else {
				plainEvent := CalendarEvent{
					ID:     event.ID,
					Name:   event.Name,
					Start:  event.Start,
					End:    event.End,
					UserID: event.UserID,
				}
				if hasLocation {
					plainEvent.Location = locationInterface.(string)
				}
				if hasDescription {
					plainEvent.Desc = descriptionInterface.(string)
				}
				plainEvents = append(plainEvents, plainEvent)
			}
		}
	}

	writeJSON(w, http.StatusOK, calendarWeekResponse{
		Status:         "ok",
		Announcements:  announcements,
		Events:         plainEvents,
		HWEvents:       hwEvents,
		ScheduleEvents: scheduleEvents,
	})
}

func routeCalendarEventsAdd(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("name") == "" || r.FormValue("start") == "" || r.FormValue("end") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	recur, recurFrequency, recurInterval, recurUntil, errorCode := parseRecurFormInfo(r)

	if errorCode != "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", errorCode})
		return
	}

	start, err := strconv.Atoi(r.FormValue("start"))
	end, err2 := strconv.Atoi(r.FormValue("end"))
	if err != nil || err2 != nil || start > end {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// insert the event
	insertResult, err := DB.Exec(
		"INSERT INTO calendar_events(name, `start`, `end`, location, `desc`, userId) VALUES(?, ?, ?, ?, ?, ?)",
		r.FormValue("name"), start, end, r.FormValue("location"), r.FormValue("desc"), c.User.ID,
	)
	if err != nil {
		errorlog.LogError("adding calendar event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	eventID, err := insertResult.LastInsertId()
	if err != nil {
		errorlog.LogError("adding calendar event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// insert the recur rule if needed
	if recur {
		_, err = DB.Exec(
			"INSERT INTO calendar_event_rules(eventId, `frequency`, `interval`, byDay, byMonthDay, byMonth, `until`) VALUES(?, ?, ?, '', 0, 0, ?)",
			eventID, recurFrequency, recurInterval, recurUntil,
		)
		if err != nil {
			errorlog.LogError("adding calendar event", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeCalendarEventsEdit(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("id") == "" || r.FormValue("name") == "" || r.FormValue("start") == "" || r.FormValue("end") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	recur, recurFrequency, recurInterval, recurUntil, errorCode := parseRecurFormInfo(r)

	if errorCode != "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", errorCode})
		return
	}

	start, err := strconv.Atoi(r.FormValue("start"))
	end, err2 := strconv.Atoi(r.FormValue("end"))
	if err != nil || err2 != nil || start > end {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM calendar_events WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id"))
	if err != nil {
		errorlog.LogError("editing calendar event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "forbidden"})
		return
	}

	// update the event
	_, err = DB.Exec(
		"UPDATE calendar_events SET name = ?, `start` = ?, `end` = ?, location = ?, `desc` = ? WHERE id = ?",
		r.FormValue("name"), start, end, r.FormValue("location"), r.FormValue("desc"), r.FormValue("id"),
	)
	if err != nil {
		errorlog.LogError("editing calendar event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// is there a recur rule associated with this event?
	recurCheckStmt, err := DB.Query("SELECT COUNT(*) FROM calendar_event_rules WHERE eventId = ?", r.FormValue("id"))
	if err != nil {
		errorlog.LogError("editing calendar event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	ruleCount := -1
	recurCheckStmt.Next()
	recurCheckStmt.Scan(&ruleCount)

	if ruleCount < 0 || ruleCount > 1 {
		errorlog.LogError("editing calendar event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	recurRuleExists := (ruleCount == 1)

	if recur {
		// want recurrence
		if recurRuleExists {
			// we have a rule -> update it
			_, err = DB.Exec(
				"UPDATE calendar_event_rules SET `frequency` = ?, `interval` = ?, `until` = ? WHERE eventId = ?",
				recurFrequency, recurInterval, recurUntil, r.FormValue("id"),
			)
			if err != nil {
				errorlog.LogError("editing calendar event", err)
				writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
				return
			}
		} else {
			// no rule -> insert it
			_, err = DB.Exec(
				"INSERT INTO calendar_event_rules(eventId, `frequency`, `interval`, byDay, byMonthDay, byMonth, `until`) VALUES(?, ?, ?, '', 0, 0, ?)",
				r.FormValue("id"), recurFrequency, recurInterval, recurUntil,
			)
			if err != nil {
				errorlog.LogError("editing calendar event", err)
				writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
				return
			}
		}
	} else {
		// don't want recurrence
		if recurRuleExists {
			// we have a rule -> delete it
			_, err = DB.Exec(
				"DELETE FROM calendar_event_rules WHERE eventId = ?",
				r.FormValue("id"),
			)
			if err != nil {
				errorlog.LogError("editing calendar event", err)
				writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
				return
			}
		} else {
			// no rule -> do nothing
		}
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeCalendarEventsDelete(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("id") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM calendar_events WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting calendar event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	// delete the event
	_, err = DB.Exec(
		"DELETE FROM calendar_events WHERE id = ?",
		r.FormValue("id"),
	)
	if err != nil {
		errorlog.LogError("deleting calendar event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// delete any recur rules associated with the event
	_, err = DB.Exec(
		"DELETE FROM calendar_event_rules WHERE eventId = ?",
		r.FormValue("id"),
	)
	if err != nil {
		errorlog.LogError("deleting calendar event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeCalendarHWEventsAdd(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("homeworkId") == "" || r.FormValue("start") == "" || r.FormValue("end") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	start, err := strconv.Atoi(r.FormValue("start"))
	end, err2 := strconv.Atoi(r.FormValue("end"))
	if err != nil || err2 != nil || start > end {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// check you own the homework you're trying to associate this with
	rows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("homeworkId"))
	if err != nil {
		errorlog.LogError("adding calendar homework event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec(
		"INSERT INTO calendar_hwevents(homeworkId, `start`, `end`, location, `desc`, userId) VALUES(?, ?, ?, ?, ?, ?)",
		r.FormValue("homeworkId"), start, end, r.FormValue("location"), r.FormValue("desc"), c.User.ID,
	)
	if err != nil {
		errorlog.LogError("adding calendar homework event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeCalendarHWEventsEdit(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("id") == "" || r.FormValue("homeworkId") == "" || r.FormValue("start") == "" || r.FormValue("end") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	start, err := strconv.Atoi(r.FormValue("start"))
	end, err2 := strconv.Atoi(r.FormValue("end"))
	if err != nil || err2 != nil || start > end {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM calendar_hwevents WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id"))
	if err != nil {
		errorlog.LogError("editing calendar homework event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "forbidden"})
		return
	}

	// check you own the homework you're trying to associate this with
	rows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("homeworkId"))
	if err != nil {
		errorlog.LogError("adding calendar homework event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec(
		"UPDATE calendar_hwevents SET homeworkId = ?, `start` = ?, `end` = ?, location = ?, `desc` = ? WHERE id = ?",
		r.FormValue("homeworkId"), start, end, r.FormValue("location"), r.FormValue("desc"), r.FormValue("id"),
	)
	if err != nil {
		errorlog.LogError("editing calendar homework event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeCalendarHWEventsDelete(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("id") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM calendar_hwevents WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting calendar homework event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec("DELETE FROM calendar_hwevents WHERE id = ?", r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting calendar homework event", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}
