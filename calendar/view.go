package calendar

import (
	"database/sql"
	"math"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
)

// A ViewDay represents a day in a View.
type ViewDay struct {
	DayString     string                     `json:"day"`
	ShiftingIndex int                        `json:"shiftingIndex"` // if it's a shifting day, its current index (for example, friday 1/2/3/4)
	CurrentTerm   *Term                      `json:"currentTerm"`
	Announcements []data.PlannerAnnouncement `json:"announcements"`
	Events        []Event                    `json:"events"`
}

// A View represents a view of a user's calendar over a certain period of time.
type View struct {
	Days []ViewDay `json:"days"`
}

func getOffBlocksStartingBefore(db *sql.DB, before string, groupSQL string) ([]OffBlock, error) {
	// find the starts
	offBlockRows, err := db.Query("SELECT id, date, text, grade FROM announcements WHERE ("+groupSQL+") AND `type` = 2 AND `date` < ?", before)
	if err != nil {
		return nil, err
	}
	defer offBlockRows.Close()
	blocks := []OffBlock{}
	for offBlockRows.Next() {
		block := OffBlock{}
		offBlockRows.Scan(&block.StartID, &block.StartText, &block.Name, &block.Grade)
		blocks = append(blocks, block)
	}

	// find the matching ends
	for i, block := range blocks {
		offBlockEndRows, err := db.Query("SELECT date FROM announcements WHERE ("+groupSQL+") AND `type` = 3 AND `text` = ?", block.Name)
		if err != nil {
			return nil, err
		}
		defer offBlockEndRows.Close()
		if offBlockEndRows.Next() {
			offBlockEndRows.Scan(&blocks[i].EndText)
		}
	}

	// parse dates
	for i, block := range blocks {
		blocks[i].Start, err = time.Parse("2006-01-02", block.StartText)
		if err != nil {
			return nil, err
		}
		blocks[i].End, err = time.Parse("2006-01-02", block.EndText)
		if err != nil {
			return nil, err
		}
	}

	return blocks, err
}

// GetView retrieves a CalendarView for the given user with the given parameters.
func GetView(db *sql.DB, userID int, announcementsGroupsSQL string, startTime time.Time, endTime time.Time) (View, error) {
	view := View{
		Days: []ViewDay{},
	}

	// get announcements for time period
	announcementRows, err := db.Query("SELECT id, date, text, grade, `type` FROM announcements WHERE date >= ? AND date <= ? AND ("+announcementsGroupsSQL+") AND type < 2", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
	if err != nil {
		return View{}, err
	}
	defer announcementRows.Close()
	announcements := []data.PlannerAnnouncement{}
	for announcementRows.Next() {
		resp := data.PlannerAnnouncement{}
		announcementRows.Scan(&resp.ID, &resp.Date, &resp.Text, &resp.Grade, &resp.Type)
		announcements = append(announcements, resp)
	}

	// get all friday information for time period
	fridayRows, err := db.Query("SELECT * FROM fridays WHERE date >= ? AND date <= ?", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
	if err != nil {
		return View{}, err
	}
	defer fridayRows.Close()
	fridays := []data.PlannerFriday{}
	for fridayRows.Next() {
		friday := data.PlannerFriday{}
		fridayRows.Scan(&friday.ID, &friday.Date, &friday.Index)
		fridays = append(fridays, friday)
	}

	// get terms for user
	termRows, err := db.Query("SELECT id, termId, name, userId FROM calendar_terms WHERE userId = ? ORDER BY name ASC", userID)
	if err != nil {
		return View{}, err
	}
	defer termRows.Close()
	availableTerms := []Term{}
	for termRows.Next() {
		term := Term{}
		termRows.Scan(&term.ID, &term.TermID, &term.Name, &term.UserID)
		availableTerms = append(availableTerms, term)
	}

	// get off blocks for time period
	offBlocks, err := getOffBlocksStartingBefore(db, endTime.Format("2006-01-02"), announcementsGroupsSQL)
	if err != nil {
		return View{}, err
	}

	// generate list of all off days in time period
	offDays := []string{}

	for _, announcement := range announcements {
		if announcement.Type == AnnouncementType_FullOff {
			offDays = append(offDays, announcement.Date)
		}
	}

	for _, offBlock := range offBlocks {
		offDayCount := int(math.Ceil(offBlock.End.Sub(offBlock.Start).Hours() / 24))
		offDay := offBlock.Start
		announcements = append(announcements, data.PlannerAnnouncement{
			ID:    offBlock.StartID,
			Date:  offBlock.StartText,
			Text:  "Start of " + offBlock.Name,
			Grade: offBlock.Grade,
			Type:  AnnouncementType_BreakStart,
		})
		for i := 0; i < offDayCount; i++ {
			if i != 0 {
				announcements = append(announcements, data.PlannerAnnouncement{
					ID:    offBlock.StartID,
					Date:  offDay.Format("2006-01-02"),
					Text:  offBlock.Name,
					Grade: offBlock.Grade,
					Type:  AnnouncementType_BreakStart,
				})
			}
			offDays = append(offDays, offDay.Format("2006-01-02"))
			offDay = offDay.Add(24 * time.Hour)
		}
		announcements = append(announcements, data.PlannerAnnouncement{
			ID:    offBlock.EndID,
			Date:  offBlock.EndText,
			Text:  "End of " + offBlock.Name,
			Grade: offBlock.Grade,
			Type:  AnnouncementType_BreakEnd,
		})
	}

	// create days in array, set friday indices
	dayCount := int(math.Ceil(endTime.Sub(startTime).Hours() / 24))
	currentDay := startTime
	for i := 0; i < dayCount; i++ {
		view.Days = append(view.Days, ViewDay{
			DayString:     currentDay.Format("2006-01-02"),
			ShiftingIndex: -1,
			CurrentTerm:   nil,
			Announcements: []data.PlannerAnnouncement{},
			Events:        []Event{},
		})

		if currentDay.Add(time.Second).After(Day_SchoolStart) && currentDay.Before(Day_SchoolEnd) {
			if currentDay.After(Day_ExamRelief) {
				// it's the second term
				view.Days[i].CurrentTerm = &availableTerms[1]
			} else {
				// it's the first term
				view.Days[i].CurrentTerm = &availableTerms[0]
			}
		}

		for _, announcement := range announcements {
			if view.Days[i].DayString == announcement.Date {
				view.Days[i].Announcements = append(view.Days[i].Announcements, announcement)
			}
		}

		if currentDay.Weekday() == time.Friday {
			for _, friday := range fridays {
				if view.Days[i].DayString == friday.Date {
					view.Days[i].ShiftingIndex = friday.Index
					break
				}
			}
		}

		currentDay = currentDay.Add(24 * time.Hour)
	}

	// get plain events
	plainEventRows, err := db.Query("SELECT id, name, `start`, `end`, `desc`, userId FROM calendar_events WHERE userId = ? AND (`end` >= ? AND `start` <= ?)", userID, startTime.Unix(), endTime.Unix())
	if err != nil {
		return View{}, err
	}
	defer plainEventRows.Close()

	for plainEventRows.Next() {
		event := Event{
			Type: PlainEvent,
		}
		data := PlainEventData{}
		plainEventRows.Scan(&event.ID, &event.Name, &event.Start, &event.End, &data.Desc, &event.UserID)
		event.Data = data

		eventStartTime := time.Unix(int64(event.Start), 0)
		dayOffset := int(math.Floor(eventStartTime.Sub(startTime).Hours() / 24))

		if dayOffset < 0 || dayOffset > len(view.Days)-1 {
			continue
		}

		view.Days[dayOffset].Events = append(view.Days[dayOffset].Events, event)
	}

	// get homework events
	hwEventRows, err := db.Query("SELECT calendar_hwevents.id, homework.id, homework.name, homework.`due`, homework.`desc`, homework.`complete`, homework.classId, homework.userId, calendar_hwevents.`start`, calendar_hwevents.`end`, calendar_hwevents.userId FROM calendar_hwevents INNER JOIN homework ON calendar_hwevents.homeworkId = homework.id WHERE calendar_hwevents.userId = ? AND (calendar_hwevents.`end` >= ? AND calendar_hwevents.`start` <= ?)", userID, startTime.Unix(), endTime.Unix())
	if err != nil {
		return View{}, err
	}
	defer hwEventRows.Close()

	for hwEventRows.Next() {
		event := Event{
			Type: HomeworkEvent,
		}
		data := HomeworkEventData{}
		hwEventRows.Scan(&event.ID, &data.Homework.ID, &data.Homework.Name, &data.Homework.Due, &data.Homework.Desc, &data.Homework.Complete, &data.Homework.ClassID, &data.Homework.UserID, &event.Start, &event.End, &event.UserID)
		event.Data = data
		event.Name = data.Homework.Name

		eventStartTime := time.Unix(int64(event.Start), 0)
		dayOffset := int(math.Floor(eventStartTime.Sub(startTime).Hours() / 24))

		if dayOffset < 0 || dayOffset > len(view.Days)-1 {
			continue
		}

		view.Days[dayOffset].Events = append(view.Days[dayOffset].Events, event)
	}

	// get schedule events
	for i := 0; i < dayCount; i++ {
		day := view.Days[i]

		if day.CurrentTerm == nil {
			continue
		}

		dayTime, _ := time.Parse("2006-01-02", day.DayString)

		dayNumber := int(dayTime.Weekday())

		if dayTime.Weekday() == time.Friday {
			if day.ShiftingIndex != -1 {
				dayNumber = 4 + day.ShiftingIndex
			} else {
				continue
			}
		}

		isOff := false

		for _, offDay := range offDays {
			if day.DayString == offDay {
				isOff = true
				break
			}
		}

		if isOff {
			continue
		}

		if dayTime.Weekday() == time.Saturday || dayTime.Weekday() == time.Sunday {
			continue
		}

		rows, err := db.Query("SELECT calendar_periods.id, calendar_classes.termId, calendar_classes.sectionId, calendar_classes.`name`, calendar_classes.ownerId, calendar_classes.ownerName, calendar_periods.dayNumber, calendar_periods.block, calendar_periods.buildingName, calendar_periods.roomNumber, calendar_periods.`start`, calendar_periods.`end`, calendar_periods.userId FROM calendar_periods INNER JOIN calendar_classes ON calendar_periods.classId = calendar_classes.sectionId WHERE calendar_periods.userId = ? AND (calendar_classes.termId = ? OR calendar_classes.termId = -1) AND calendar_periods.dayNumber = ? GROUP BY calendar_periods.id, calendar_classes.termId, calendar_classes.name, calendar_classes.ownerId, calendar_classes.ownerName", userID, day.CurrentTerm.TermID, dayNumber)
		if err != nil {
			return View{}, err
		}
		defer rows.Close()
		for rows.Next() {
			event := Event{
				Type: ScheduleEvent,
			}
			data := ScheduleEventData{}
			rows.Scan(&event.ID, &data.TermID, &data.ClassID, &event.Name, &data.OwnerID, &data.OwnerName, &data.DayNumber, &data.Block, &data.BuildingName, &data.RoomNumber, &event.Start, &event.End, &event.UserID)
			event.Data = data

			event.Start += int(dayTime.Unix())
			event.End += int(dayTime.Unix())

			view.Days[i].Events = append(view.Days[i].Events, event)
		}

		if dayTime.Weekday() == time.Thursday {
			// special case: assembly
			for eventIndex, event := range view.Days[i].Events {
				// check for an "HS House" event
				// starting 11:50, ending 12:50
				if strings.HasPrefix(event.Name, "HS House") && event.Start == int(dayTime.Unix())+42600 && event.End == int(dayTime.Unix())+46200 {
					// found it
					view.Days[i].Events[eventIndex].Name = "Assembly"
				}
			}
		}
	}

	return view, nil
}
