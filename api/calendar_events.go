package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/MyHomeworkSpace/api-server/calendar"
	"github.com/MyHomeworkSpace/api-server/data"

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

// responses
type CalendarWeekResponse struct {
	Status         string                     `json:"status"`
	Announcements  []data.PlannerAnnouncement `json:"announcements"`
	CurrentTerm    *calendar.Term             `json:"currentTerm"`
	Friday         data.PlannerFriday         `json:"friday"`
	Events         []CalendarEvent            `json:"events"`
	HWEvents       []CalendarHWEvent          `json:"hwEvents"`
	ScheduleEvents [][]CalendarScheduleItem   `json:"scheduleEvents"`
}
type CalendarEventResponse struct {
	Status string          `json:"status"`
	Events []CalendarEvent `json:"events"`
}
type SingleCalendarEventResponse struct {
	Status string        `json:"status"`
	Event  CalendarEvent `json:"event"`
}
type CalendarViewResponse struct {
	Status string        `json:"status"`
	View   calendar.View `json:"view"`
}

func InitCalendarEventsAPI(e *echo.Echo) {
	e.GET("/calendar/events/getWeek/:monday", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		userID := GetSessionUserID(&c)

		user, err := Data_GetUserByID(userID)
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

		startDate, err := time.Parse("2006-01-02", c.Param("monday"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}
		endDate := startDate.Add(time.Hour * 24 * 7)

		view, err := calendar.GetView(DB, userID, time.UTC, announcementsGroupsSQL, startDate, endDate)
		if err != nil {
			ErrorLog_LogError("getting calendar week", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		// get all terms for this user
		termRows, err := DB.Query("SELECT id, termId, name, userId FROM calendar_terms WHERE userId = ? ORDER BY name ASC", userID)
		if err != nil {
			ErrorLog_LogError("getting term information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer termRows.Close()
		availableTerms := []calendar.Term{}
		for termRows.Next() {
			term := calendar.Term{}
			termRows.Scan(&term.ID, &term.TermID, &term.Name, &term.UserID)
			availableTerms = append(availableTerms, term)
		}

		// find the current term
		// TODO: be better at this and handle mid-week term switches (is that a thing?)
		var currentTerm *calendar.Term
		if startDate.Add(time.Second).After(calendar.Day_SchoolStart) && startDate.Before(calendar.Day_SchoolEnd) {
			if startDate.After(calendar.Day_ExamRelief) {
				// it's the second term
				currentTerm = &availableTerms[1]
			} else {
				// it's the first term
				currentTerm = &availableTerms[0]
			}
		}

		// get friday info
		fridayDate := startDate.Add(time.Hour * 24 * 4)

		fridayRows, err := DB.Query("SELECT * FROM fridays WHERE date = ?", fridayDate)
		if err != nil {
			ErrorLog_LogError("getting friday information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer fridayRows.Close()
		friday := data.PlannerFriday{-1, "", -1}
		if fridayRows.Next() {
			fridayRows.Scan(&friday.ID, &friday.Date, &friday.Index)
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
				if event.Type == calendar.PlainEvent {
					data := event.Data.(calendar.PlainEventData)
					plainEvent := CalendarEvent{
						ID:     event.ID,
						Name:   event.Name,
						Start:  event.Start,
						End:    event.End,
						Desc:   data.Desc,
						UserID: event.UserID,
					}
					plainEvents = append(plainEvents, plainEvent)
				} else if event.Type == calendar.HomeworkEvent {
					data := event.Data.(calendar.HomeworkEventData)
					hwEvent := CalendarHWEvent{
						ID:       event.ID,
						Homework: data.Homework,
						Start:    event.Start,
						End:      event.End,
						UserID:   event.UserID,
					}
					hwEvents = append(hwEvents, hwEvent)
				} else if event.Type == calendar.ScheduleEvent {
					dayTime, _ := time.Parse("2006-01-02", day.DayString)
					data := event.Data.(calendar.ScheduleEventData)
					scheduleEvent := CalendarScheduleItem{
						ID:           event.ID,
						TermID:       data.TermID,
						ClassID:      data.ClassID,
						Name:         event.Name,
						OwnerID:      data.OwnerID,
						OwnerName:    data.OwnerName,
						DayNumber:    data.DayNumber,
						Block:        data.Block,
						BuildingName: data.BuildingName,
						RoomNumber:   data.RoomNumber,
						Start:        event.Start - int(dayTime.Unix()),
						End:          event.End - int(dayTime.Unix()),
						UserID:       event.UserID,
					}
					scheduleEvents[dayIndex] = append(scheduleEvents[dayIndex], scheduleEvent)
				}
			}
		}

		return c.JSON(http.StatusOK, CalendarWeekResponse{
			Status:         "ok",
			Announcements:  announcements,
			CurrentTerm:    currentTerm,
			Friday:         friday,
			Events:         plainEvents,
			HWEvents:       hwEvents,
			ScheduleEvents: scheduleEvents,
		})
	})

	// normal events
	e.POST("/calendar/events/add", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("name") == "" || c.FormValue("start") == "" || c.FormValue("end") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		start, err := strconv.Atoi(c.FormValue("start"))
		end, err2 := strconv.Atoi(c.FormValue("end"))
		if err != nil || err2 != nil || start > end {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		stmt, err := DB.Prepare("INSERT INTO calendar_events(name, `start`, `end`, `desc`, userId) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			ErrorLog_LogError("adding calendar event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("name"), start, end, c.FormValue("desc"), GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("adding calendar event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
	e.POST("/calendar/events/edit", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("id") == "" || c.FormValue("name") == "" || c.FormValue("start") == "" || c.FormValue("end") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		start, err := strconv.Atoi(c.FormValue("start"))
		end, err2 := strconv.Atoi(c.FormValue("end"))
		if err != nil || err2 != nil || start > end {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		// check if you are allowed to edit the given id
		idRows, err := DB.Query("SELECT id FROM calendar_events WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("editing calendar event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer idRows.Close()
		if !idRows.Next() {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("UPDATE calendar_events SET name = ?, `start` = ?, `end` = ?, `desc` = ? WHERE id = ?")
		if err != nil {
			ErrorLog_LogError("editing calendar event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("name"), start, end, c.FormValue("desc"), c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("editing calendar event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
	e.POST("/calendar/events/delete", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("id") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		// check if you are allowed to edit the given id
		idRows, err := DB.Query("SELECT id FROM calendar_events WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("deleting calendar event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer idRows.Close()
		if !idRows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("DELETE FROM calendar_events WHERE id = ?")
		if err != nil {
			ErrorLog_LogError("deleting calendar event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("deleting calendar event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	// homework events
	e.POST("/calendar/hwEvents/add", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("homeworkId") == "" || c.FormValue("start") == "" || c.FormValue("end") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		start, err := strconv.Atoi(c.FormValue("start"))
		end, err2 := strconv.Atoi(c.FormValue("end"))
		if err != nil || err2 != nil || start > end {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		// check you own the homework you're trying to associate this with
		rows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("homeworkId"))
		if err != nil {
			ErrorLog_LogError("adding calendar homework event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		if !rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("INSERT INTO calendar_hwevents(homeworkId, `start`, `end`, userId) VALUES(?, ?, ?, ?)")
		if err != nil {
			ErrorLog_LogError("adding calendar homework event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("homeworkId"), start, end, GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("adding calendar homework event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
	e.POST("/calendar/hwEvents/edit", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("id") == "" || c.FormValue("homeworkId") == "" || c.FormValue("start") == "" || c.FormValue("end") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		start, err := strconv.Atoi(c.FormValue("start"))
		end, err2 := strconv.Atoi(c.FormValue("end"))
		if err != nil || err2 != nil || start > end {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		// check if you are allowed to edit the given id
		idRows, err := DB.Query("SELECT id FROM calendar_hwevents WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("editing calendar homework event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer idRows.Close()
		if !idRows.Next() {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "forbidden"})
		}

		// check you own the homework you're trying to associate this with
		rows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("homeworkId"))
		if err != nil {
			ErrorLog_LogError("adding calendar homework event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		if !rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("UPDATE calendar_hwevents SET homeworkId = ?, `start` = ?, `end` = ? WHERE id = ?")
		if err != nil {
			ErrorLog_LogError("editing calendar homework event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("homeworkId"), start, end, c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("editing calendar homework event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
	e.POST("/calendar/hwEvents/delete", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("id") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		// check if you are allowed to edit the given id
		idRows, err := DB.Query("SELECT id FROM calendar_hwevents WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("deleting calendar homework event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer idRows.Close()
		if !idRows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("DELETE FROM calendar_hwevents WHERE id = ?")
		if err != nil {
			ErrorLog_LogError("deleting calendar homework event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("deleting calendar homework event", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
