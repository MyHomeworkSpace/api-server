package api

import (
	"log"
	"net/http"
	"strconv"
	"strings"
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

		announcementRows, err := DB.Query("SELECT id, date, text, grade, `type` FROM announcements WHERE date >= ? AND date < ? AND ("+announcementsGroupsSQL+") AND type < 2", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
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

		// get off days in this week
		offDayRows, err := DB.Query("SELECT date FROM announcements WHERE date >= ? AND date < ? AND ("+announcementsGroupsSQL+") AND `type` = "+strconv.Itoa(AnnouncementType_FullOff), startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		if err != nil {
			log.Println("Error while getting off day information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer offDayRows.Close()
		offDays := []string{}
		for offDayRows.Next() {
			day := ""
			offDayRows.Scan(&day)
			offDays = append(offDays, day)
		}

		// get off blocks that might apply
		offBlocks, err := Data_GetOffBlocksStartingBefore(endDate.Format("2006-01-02"), announcementsGroups)
		if err != nil {
			log.Println("Error while getting off block information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
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

			// block out off days
			// TODO: can this be made more efficient?
			currentDay := startDate
			for dayIndex := 0; dayIndex < 5; dayIndex++ {
				dayStr := currentDay.Format("2006-01-02")
				if Util_StringSliceContains(offDays, dayStr) {
					// it's in the list of off days for this week, empty it
					scheduleEvents[dayIndex] = []CalendarScheduleItem{}
				}
				currentDay = currentDay.Add(24 * time.Hour)
			}

			// block out off blocks
			// TODO: can this be offloaded to mysql somehow, which would probably be faster?
			for _, block := range offBlocks {
				if !block.Start.After(endDate) || !endDate.After(block.End) {
					// there is overlap
					// find what day the overlap starts
					dayOverlapStarts := 0
					dayOverlapEnds := 7
					if !block.Start.Before(startDate) {
						dayOverlapStarts = int(block.Start.Sub(startDate) / (24 * time.Hour))
					}
					if block.End.Before(endDate) {
						dayOverlapEnds = (7 - int(endDate.Sub(block.End)/(24*time.Hour)))
					}
					if dayOverlapEnds > 7 {
						dayOverlapEnds = 7
					}

					// add the start announcement on that day
					announcements = append(announcements, PlannerAnnouncement{
						block.StartID,
						block.StartText,
						"Start of " + block.Name,
						block.Grade,
						AnnouncementType_BreakStart,
					})

					// add the end announcement on that day
					announcements = append(announcements, PlannerAnnouncement{
						block.EndID,
						block.EndText,
						"End of " + block.Name,
						block.Grade,
						AnnouncementType_BreakEnd,
					})

					// block out days
					currentDay = startDate.Add(time.Duration(dayOverlapStarts*24) * time.Hour)
					for dayIndex := dayOverlapStarts; dayIndex < (dayOverlapEnds + 1); dayIndex++ {
						if dayIndex < 5 && !currentDay.Equal(block.End) {
							scheduleEvents[dayIndex] = []CalendarScheduleItem{}
						}

						if !currentDay.Equal(block.Start) && !currentDay.Equal(block.End) {
							announcements = append(announcements, PlannerAnnouncement{
								block.StartID,
								currentDay.Format("2006-01-02"),
								block.Name,
								block.Grade,
								AnnouncementType_BreakStart,
							})
						}

						currentDay = currentDay.Add(24 * time.Hour)
					}
				}
			}

			// check for assembly on thursday
			if scheduleEvents[3] != nil {
				for i, event := range scheduleEvents[3] {
					// check for an "HS House" event
					// starting 11:50, ending 12:50
					if strings.HasPrefix(event.Name, "HS House") && event.Start == 42600 && event.End == 46200 {
						// found it
						scheduleEvents[3][i].Name = "Assembly"
					}
				}
			}

			// special exception: candlelighting
			if startDate.Before(Day_Candlelighting) && endDate.After(Day_Candlelighting) {
				dayNumber := (int(Day_Candlelighting.Weekday()) - 1)

				scheduleEvents[dayNumber] = []CalendarScheduleItem{}

				itemsForPeriod := map[string][]CalendarScheduleItem{}
				seenClassIds := []int{}

				rows, err := DB.Query("SELECT calendar_periods.id, calendar_classes.termId, calendar_classes.sectionId, calendar_classes.`name`, calendar_classes.ownerId, calendar_classes.ownerName, calendar_periods.dayNumber, calendar_periods.block, calendar_periods.buildingName, calendar_periods.roomNumber, calendar_periods.`start`, calendar_periods.`end`, calendar_periods.userId FROM calendar_periods INNER JOIN calendar_classes ON calendar_periods.classId = calendar_classes.sectionId WHERE calendar_periods.userId = ? AND (calendar_classes.termId = ? OR calendar_classes.termId = -1) AND calendar_periods.block IN ('C', 'D', 'H', 'G') GROUP BY calendar_periods.id, calendar_classes.termId, calendar_classes.name, calendar_classes.ownerId, calendar_classes.ownerName", userId, currentTerm.TermID)
				if err != nil {
					log.Println("Error while getting calendar events: ")
					log.Println(err)
					return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				}
				defer rows.Close()
				for rows.Next() {
					item := CalendarScheduleItem{}
					rows.Scan(&item.ID, &item.TermID, &item.ClassID, &item.Name, &item.OwnerID, &item.OwnerName, &item.DayNumber, &item.Block, &item.BuildingName, &item.RoomNumber, &item.Start, &item.End, &item.UserID)
					if !Util_IntSliceContains(seenClassIds, item.ClassID) {
						_, sliceExists := itemsForPeriod[item.Block]
						if !sliceExists {
							itemsForPeriod[item.Block] = []CalendarScheduleItem{}
						}
						itemsForPeriod[item.Block] = append(itemsForPeriod[item.Block], item)
						seenClassIds = append(seenClassIds, item.ClassID)
					}
				}

				for _, specialScheduleItem := range SpecialSchedule_HS_Candlelighting {
					if specialScheduleItem.Block != "" {
						// it references a specific class, look it up
						items, haveItems := itemsForPeriod[specialScheduleItem.Block]
						if haveItems {
							for _, scheduleItem := range items {
								newItem := scheduleItem
								newItem.Start = specialScheduleItem.Start
								newItem.End = specialScheduleItem.End
								// remove building + room number because we can't tell which one to use
								// (in the case of classes where the room changes, like most science classes)
								newItem.BuildingName = ""
								newItem.RoomNumber = ""
								scheduleEvents[dayNumber] = append(scheduleEvents[dayNumber], newItem)
							}
						} else {
							// the person doesn't have a class for that period
							// just skip it
						}
					} else {
						// it's a fixed thing, just add it directly
						newItem := CalendarScheduleItem{
							ID:           -1,
							TermID:       currentTerm.TermID,
							ClassID:      -1,
							Name:         specialScheduleItem.Name,
							OwnerID:      -1,
							OwnerName:    "",
							DayNumber:    dayNumber,
							Block:        "",
							BuildingName: "",
							RoomNumber:   "",
							Start:        specialScheduleItem.Start,
							End:          specialScheduleItem.End,
							UserID:       -1,
						}
						scheduleEvents[dayNumber] = append(scheduleEvents[dayNumber], newItem)
					}
				}
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
