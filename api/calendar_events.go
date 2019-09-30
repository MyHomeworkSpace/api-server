package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/MyHomeworkSpace/api-server/calendar"
	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"

	"github.com/labstack/echo"
)

// structs for data
type CalendarEvent struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Start  int    `json:"start"`
	End    int    `json:"end"`
	Desc   string `json:"desc"`
	UserID int    `json:"userId"`
}
type CalendarHWEvent struct {
	ID       int           `json:"id"`
	Homework data.Homework `json:"homework"`
	Start    int           `json:"start"`
	End      int           `json:"end"`
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

func parseRecurFormInfo(ec echo.Context) (bool, int, int, string, string) {
	if ec.FormValue("recur") != "" {
		recurStr := ec.FormValue("recur")
		recur, err := strconv.ParseBool(recurStr)
		if err != nil {
			return false, 0, 0, "", "invalid_params"
		}

		if !recur {
			return false, 0, 0, "", ""
		}

		if ec.FormValue("recurFrequency") == "" || ec.FormValue("recurInterval") == "" {
			return false, 0, 0, "", "missing_params"
		}

		recurFrequency, err := strconv.Atoi(ec.FormValue("recurFrequency"))
		recurInterval, err1 := strconv.Atoi(ec.FormValue("recurInterval"))
		recurUntil := ""
		if ec.FormValue("recurUntil") != "" {
			_, err2 := time.Parse("2006-01-02", ec.FormValue("recurUntil"))
			recurUntil = ec.FormValue("recurUntil")
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

func routeCalendarEventsGetWeek(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	startDate, err := time.Parse("2006-01-02", ec.Param("monday"))
	if err != nil {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}
	endDate := startDate.Add(time.Hour * 24 * 7)

	view, err := calendar.GetView(DB, c.User, time.UTC, startDate, endDate)
	if err != nil {
		errorlog.LogError("getting calendar week", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
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
			descriptionInterface, isPlain := event.Tags[data.EventTagDescription]
			homeworkInterface, isHomework := event.Tags[data.EventTagHomework]
			_, isSchedule := event.Tags[data.EventTagClassID]
			if isPlain {
				plainEvent := CalendarEvent{
					ID:     event.ID,
					Name:   event.Name,
					Start:  event.Start,
					End:    event.End,
					Desc:   descriptionInterface.(string),
					UserID: event.UserID,
				}
				plainEvents = append(plainEvents, plainEvent)
			} else if isHomework {
				homework := homeworkInterface.(data.Homework)
				hwEvent := CalendarHWEvent{
					ID:       event.ID,
					Homework: homework,
					Start:    event.Start,
					End:      event.End,
					UserID:   event.UserID,
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
			}
		}
	}

	ec.JSON(http.StatusOK, calendarWeekResponse{
		Status:         "ok",
		Announcements:  announcements,
		Events:         plainEvents,
		HWEvents:       hwEvents,
		ScheduleEvents: scheduleEvents,
	})
}

func routeCalendarEventsAdd(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("name") == "" || ec.FormValue("start") == "" || ec.FormValue("end") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	recur, recurFrequency, recurInterval, recurUntil, errorCode := parseRecurFormInfo(ec)

	if errorCode != "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", errorCode})
		return
	}

	start, err := strconv.Atoi(ec.FormValue("start"))
	end, err2 := strconv.Atoi(ec.FormValue("end"))
	if err != nil || err2 != nil || start > end {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// insert the event
	insertResult, err := DB.Exec(
		"INSERT INTO calendar_events(name, `start`, `end`, `desc`, userId) VALUES(?, ?, ?, ?, ?)",
		ec.FormValue("name"), start, end, ec.FormValue("desc"), c.User.ID,
	)
	if err != nil {
		errorlog.LogError("adding calendar event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	eventID, err := insertResult.LastInsertId()
	if err != nil {
		errorlog.LogError("adding calendar event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
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
			ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeCalendarEventsEdit(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("id") == "" || ec.FormValue("name") == "" || ec.FormValue("start") == "" || ec.FormValue("end") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	recur, recurFrequency, recurInterval, recurUntil, errorCode := parseRecurFormInfo(ec)

	if errorCode != "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", errorCode})
		return
	}

	start, err := strconv.Atoi(ec.FormValue("start"))
	end, err2 := strconv.Atoi(ec.FormValue("end"))
	if err != nil || err2 != nil || start > end {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM calendar_events WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("editing calendar event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "forbidden"})
		return
	}

	// update the event
	_, err = DB.Exec(
		"UPDATE calendar_events SET name = ?, `start` = ?, `end` = ?, `desc` = ? WHERE id = ?",
		ec.FormValue("name"), start, end, ec.FormValue("desc"), ec.FormValue("id"),
	)
	if err != nil {
		errorlog.LogError("editing calendar event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// is there a recur rule associated with this event?
	recurCheckStmt, err := DB.Query("SELECT COUNT(*) FROM calendar_event_rules WHERE eventId = ?", ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("editing calendar event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	ruleCount := -1
	recurCheckStmt.Next()
	recurCheckStmt.Scan(&ruleCount)

	if ruleCount < 0 || ruleCount > 1 {
		errorlog.LogError("editing calendar event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	recurRuleExists := (ruleCount == 1)

	if recur {
		// want recurrence
		if recurRuleExists {
			// we have a rule -> update it
			_, err = DB.Exec(
				"UPDATE calendar_event_rules SET `frequency` = ?, `interval` = ?, `until` = ? WHERE eventId = ?",
				recurFrequency, recurInterval, recurUntil, ec.FormValue("id"),
			)
			if err != nil {
				errorlog.LogError("editing calendar event", err)
				ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
				return
			}
		} else {
			// no rule -> insert it
			_, err = DB.Exec(
				"INSERT INTO calendar_event_rules(eventId, `frequency`, `interval`, byDay, byMonthDay, byMonth, `until`) VALUES(?, ?, ?, '', 0, 0, ?)",
				ec.FormValue("id"), recurFrequency, recurInterval, recurUntil,
			)
			if err != nil {
				errorlog.LogError("editing calendar event", err)
				ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
				return
			}
		}
	} else {
		// don't want recurrence
		if recurRuleExists {
			// we have a rule -> delete it
			_, err = DB.Exec(
				"DELETE FROM calendar_event_rules WHERE eventId = ?",
				ec.FormValue("id"),
			)
			if err != nil {
				errorlog.LogError("editing calendar event", err)
				ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
				return
			}
		} else {
			// no rule -> do nothing
		}
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeCalendarEventsDelete(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("id") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM calendar_events WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting calendar event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		ec.JSON(http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	// delete the event
	_, err = DB.Exec(
		"DELETE FROM calendar_events WHERE id = ?",
		ec.FormValue("id"),
	)
	if err != nil {
		errorlog.LogError("deleting calendar event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// delete any recur rules associated with the event
	_, err = DB.Exec(
		"DELETE FROM calendar_event_rules WHERE eventId = ?",
		ec.FormValue("id"),
	)
	if err != nil {
		errorlog.LogError("deleting calendar event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeCalendarHWEventsAdd(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("homeworkId") == "" || ec.FormValue("start") == "" || ec.FormValue("end") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	start, err := strconv.Atoi(ec.FormValue("start"))
	end, err2 := strconv.Atoi(ec.FormValue("end"))
	if err != nil || err2 != nil || start > end {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// check you own the homework you're trying to associate this with
	rows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("homeworkId"))
	if err != nil {
		errorlog.LogError("adding calendar homework event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		ec.JSON(http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec(
		"INSERT INTO calendar_hwevents(homeworkId, `start`, `end`, userId) VALUES(?, ?, ?, ?)",
		ec.FormValue("homeworkId"), start, end, c.User.ID,
	)
	if err != nil {
		errorlog.LogError("adding calendar homework event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeCalendarHWEventsEdit(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("id") == "" || ec.FormValue("homeworkId") == "" || ec.FormValue("start") == "" || ec.FormValue("end") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	start, err := strconv.Atoi(ec.FormValue("start"))
	end, err2 := strconv.Atoi(ec.FormValue("end"))
	if err != nil || err2 != nil || start > end {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM calendar_hwevents WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("editing calendar homework event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "forbidden"})
		return
	}

	// check you own the homework you're trying to associate this with
	rows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("homeworkId"))
	if err != nil {
		errorlog.LogError("adding calendar homework event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		ec.JSON(http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec(
		"UPDATE calendar_hwevents SET homeworkId = ?, `start` = ?, `end` = ? WHERE id = ?",
		ec.FormValue("homeworkId"), start, end, ec.FormValue("id"),
	)
	if err != nil {
		errorlog.LogError("editing calendar homework event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeCalendarHWEventsDelete(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("id") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM calendar_hwevents WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting calendar homework event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		ec.JSON(http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec("DELETE FROM calendar_hwevents WHERE id = ?", ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting calendar homework event", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}
