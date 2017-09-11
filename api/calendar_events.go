package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

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
	ID       int      `json:"id"`
	Homework Homework `json:"homework"`
	Start    int      `json:"start"`
	End      int      `json:"end"`
	UserID   int      `json:"userId"`
}

// responses
type CalendarWeekResponse struct {
	Status         string                   `json:"status"`
	Announcements  []PlannerAnnouncement    `json:"announcements"`
	CurrentTerm    *CalendarTerm            `json:"currentTerm"`
	Friday         PlannerFriday            `json:"friday"`
	Events         []CalendarEvent          `json:"events"`
	HWEvents       []CalendarHWEvent        `json:"hwEvents"`
	ScheduleEvents [][]CalendarScheduleItem `json:"scheduleEvents"`
}
type CalendarEventResponse struct {
	Status string          `json:"status"`
	Events []CalendarEvent `json:"events"`
}
type SingleCalendarEventResponse struct {
	Status string        `json:"status"`
	Event  CalendarEvent `json:"event"`
}

func InitCalendarEventsAPI(e *echo.Echo) {
	e.GET("/calendar/events/getWeek/:monday", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		userId := GetSessionUserID(&c)

		startDate, err := time.Parse("2006-01-02", c.Param("monday"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}
		endDate := startDate.Add(time.Hour * 24 * 7)

		// get announcements
		user, err := Data_GetUserByID(userId)
		if err != nil {
			log.Println("Error while getting planner week information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		grade, err := Data_GetUserGrade(user)
		if err != nil {
			log.Println("Error while getting planner week information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		announcementsGroups := Data_GetGradeAnnouncementGroups(grade)
		announcementsGroupsSQL := Data_GetAnnouncementGroupSQL(announcementsGroups)

		announcementRows, err := DB.Query("SELECT id, date, text, grade, `type` FROM announcements WHERE date >= ? AND date < ? AND ("+announcementsGroupsSQL+")", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		if err != nil {
			log.Println("Error while getting announcement information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer announcementRows.Close()
		announcements := []PlannerAnnouncement{}
		for announcementRows.Next() {
			resp := PlannerAnnouncement{-1, "", "", -1, -1}
			announcementRows.Scan(&resp.ID, &resp.Date, &resp.Text, &resp.Grade, &resp.Type)
			announcements = append(announcements, resp)
		}

		// get all terms for this user
		termRows, err := DB.Query("SELECT id, termId, name, userId FROM calendar_terms WHERE userId = ? ORDER BY name ASC", userId)
		if err != nil {
			log.Println("Error while getting term information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer termRows.Close()
		availableTerms := []CalendarTerm{}
		for termRows.Next() {
			term := CalendarTerm{}
			termRows.Scan(&term.ID, &term.TermID, &term.Name, &term.UserID)
			availableTerms = append(availableTerms, term)
		}

		// find the current term
		// TODO: be better at this and handle mid-week term switches (is that a thing?)
		var currentTerm *CalendarTerm
		if startDate.Add(time.Second).After(Day_SchoolStart) && startDate.Before(Day_SchoolEnd) {
			if startDate.After(Day_ExamRelief) {
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
			log.Println("Error while getting friday information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer fridayRows.Close()
		friday := PlannerFriday{-1, "", -1}
		if fridayRows.Next() {
			fridayRows.Scan(&friday.ID, &friday.Date, &friday.Index)
		}

		// get normal events
		eventRows, err := DB.Query("SELECT id, name, `start`, `end`, `desc`, userId FROM calendar_events WHERE userId = ? AND (`end` >= ? AND `start` <= ?)", userId, startDate.Unix(), endDate.Unix())
		if err != nil {
			log.Println("Error while getting calendar events: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer eventRows.Close()

		events := []CalendarEvent{}
		for eventRows.Next() {
			event := CalendarEvent{-1, "", -1, -1, "", -1}
			eventRows.Scan(&event.ID, &event.Name, &event.Start, &event.End, &event.Desc, &event.UserID)
			events = append(events, event)
		}

		// get homework events
		hwEventRows, err := DB.Query("SELECT calendar_hwevents.id, homework.id, homework.name, homework.`due`, homework.`desc`, homework.`complete`, homework.classId, homework.userId, calendar_hwevents.`start`, calendar_hwevents.`end`, calendar_hwevents.userId FROM calendar_hwevents INNER JOIN homework ON calendar_hwevents.homeworkId = homework.id WHERE calendar_hwevents.userId = ? AND (calendar_hwevents.`end` >= ? AND calendar_hwevents.`start` <= ?)", userId, startDate.Unix(), endDate.Unix())
		if err != nil {
			log.Println("Error while getting calendar events: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer hwEventRows.Close()

		hwEvents := []CalendarHWEvent{}
		for hwEventRows.Next() {
			hwEvent := CalendarHWEvent{-1, Homework{-1, "", "", "", -1, -1, -1}, -1, -1, -1}
			hwEventRows.Scan(&hwEvent.ID, &hwEvent.Homework.ID, &hwEvent.Homework.Name, &hwEvent.Homework.Due, &hwEvent.Homework.Desc, &hwEvent.Homework.Complete, &hwEvent.Homework.ClassID, &hwEvent.Homework.UserID, &hwEvent.Start, &hwEvent.End, &hwEvent.UserID)
			hwEvents = append(hwEvents, hwEvent)
		}

		// get schedule events
		var scheduleEvents [][]CalendarScheduleItem
		if currentTerm != nil {
			// there actually is school this week
			scheduleEvents = make([][]CalendarScheduleItem, 5)
			for dayIndex := 0; dayIndex < 5; dayIndex++ {
				dayNumber := dayIndex
				if dayNumber == 4 { // friday
					// use the current friday index
					if friday.Index == -1 {
						// there's no friday info for this week, so display a blank schedule
						continue
					}
					dayNumber = 4 + friday.Index - 1
				}

				// blackbaud day numbers are off by one because they treat sunday as 0
				dayNumber = dayNumber + 1

				dayEvents := make([]CalendarScheduleItem, 0)

				// fetch items for this day
				rows, err := DB.Query("SELECT calendar_periods.id, calendar_classes.termId, calendar_classes.sectionId, calendar_classes.`name`, calendar_classes.ownerId, calendar_classes.ownerName, calendar_periods.dayNumber, calendar_periods.block, calendar_periods.buildingName, calendar_periods.roomNumber, calendar_periods.`start`, calendar_periods.`end`, calendar_periods.userId FROM calendar_periods INNER JOIN calendar_classes ON calendar_periods.classId = calendar_classes.sectionId WHERE calendar_periods.userId = ? AND (calendar_classes.termId = ? OR calendar_classes.termId = -1) AND calendar_periods.dayNumber = ? GROUP BY calendar_periods.id, calendar_classes.termId, calendar_classes.name, calendar_classes.ownerId, calendar_classes.ownerName", userId, currentTerm.TermID, dayNumber)
				if err != nil {
					log.Println("Error while getting calendar events: ")
					log.Println(err)
					return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				}
				defer rows.Close()
				for rows.Next() {
					item := CalendarScheduleItem{}
					rows.Scan(&item.ID, &item.TermID, &item.ClassID, &item.Name, &item.OwnerID, &item.OwnerName, &item.DayNumber, &item.Block, &item.BuildingName, &item.RoomNumber, &item.Start, &item.End, &item.UserID)
					dayEvents = append(dayEvents, item)
				}

				scheduleEvents[dayIndex] = dayEvents
			}
		}

		return c.JSON(http.StatusOK, CalendarWeekResponse{
			Status:         "ok",
			Announcements:  announcements,
			CurrentTerm:    currentTerm,
			Friday:         friday,
			Events:         events,
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
			log.Println("Error while adding calendar event: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("name"), start, end, c.FormValue("desc"), GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while adding calendar event: ")
			log.Println(err)
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
			log.Println("Error while editing calendar event: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer idRows.Close()
		if !idRows.Next() {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("UPDATE calendar_events SET name = ?, `start` = ?, `end` = ?, `desc` = ? WHERE id = ?")
		if err != nil {
			log.Println("Error while editing calendar event: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("name"), start, end, c.FormValue("desc"), c.FormValue("id"))
		if err != nil {
			log.Println("Error while editing calendar event: ")
			log.Println(err)
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
			log.Println("Error while deleting calendar event: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer idRows.Close()
		if !idRows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("DELETE FROM calendar_events WHERE id = ?")
		if err != nil {
			log.Println("Error while deleting calendar event: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting calendar event: ")
			log.Println(err)
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
			log.Println("Error while adding calendar homework event:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		if !rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("INSERT INTO calendar_hwevents(homeworkId, `start`, `end`, userId) VALUES(?, ?, ?, ?)")
		if err != nil {
			log.Println("Error while adding calendar homework event:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("homeworkId"), start, end, GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while adding calendar homework event:")
			log.Println(err)
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
			log.Println("Error while editing calendar homework event: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer idRows.Close()
		if !idRows.Next() {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "forbidden"})
		}

		// check you own the homework you're trying to associate this with
		rows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("homeworkId"))
		if err != nil {
			log.Println("Error while adding calendar homework event:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		if !rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("UPDATE calendar_hwevents SET homeworkId = ?, `start` = ?, `end` = ? WHERE id = ?")
		if err != nil {
			log.Println("Error while editing calendar homework event: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("homeworkId"), start, end, c.FormValue("id"))
		if err != nil {
			log.Println("Error while editing calendar homework event: ")
			log.Println(err)
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
			log.Println("Error while deleting calendar homework event: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer idRows.Close()
		if !idRows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("DELETE FROM calendar_hwevents WHERE id = ?")
		if err != nil {
			log.Println("Error while deleting calendar homework event: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting calendar homework event: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
