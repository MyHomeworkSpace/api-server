package calendar

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools/manager"
	"github.com/MyHomeworkSpace/api-server/util"
)

// A ViewDay represents a day in a View.
type ViewDay struct {
	DayString     string                     `json:"day"`
	CurrentTerm   *Term                      `json:"currentTerm"`
	Announcements []data.PlannerAnnouncement `json:"announcements"`
	Events        []data.Event               `json:"events"`
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
		offBlockEndRows, err := db.Query("SELECT date FROM announcements WHERE ("+groupSQL+") AND `type` = 3 AND `text` = ? AND `date` > ?", block.Name, block.StartText)
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
func GetView(db *sql.DB, user *data.User, location *time.Location, grade int, announcementsGroupsSQL string, startTime time.Time, endTime time.Time) (View, error) {
	view := View{
		Days: []ViewDay{},
	}

	providers := []data.Provider{
		// TODO: not hardcode this for dalton
		manager.GetSchoolByID("dalton").CalendarProvider(),
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
	termRows, err := db.Query("SELECT id, termId, name, userId FROM calendar_terms WHERE userId = ? ORDER BY name ASC", user.ID)
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

	// if user is a senior, their classes end earlier
	lastDayOfClasses := Day_SchoolEnd
	if grade == 12 {
		lastDayOfClasses = Day_SeniorLastDay
	}

	// create days in array, set friday indices
	dayCount := int((endTime.Sub(startTime).Hours() / 24) + 0.5)
	currentDay := startTime
	viewIncludesSpecialAssessmentDay := false
	for i := 0; i < dayCount; i++ {
		view.Days = append(view.Days, ViewDay{
			DayString:     currentDay.Format("2006-01-02"),
			CurrentTerm:   nil,
			Announcements: []data.PlannerAnnouncement{},
			Events:        []data.Event{},
		})

		if currentDay.Add(time.Second).After(Day_SchoolStart) && currentDay.Before(lastDayOfClasses) {
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

		// do we have special assessments?
		for specialAssessmentDay, _ := range SpecialAssessmentDays {
			if view.Days[i].DayString == specialAssessmentDay {
				viewIncludesSpecialAssessmentDay = true
				break
			}
		}

		currentDay = currentDay.AddDate(0, 0, 1)
	}

	// get plain events
	plainEventRows, err := db.Query(
		"SELECT calendar_events.id, calendar_events.name, calendar_events.`start`, calendar_events.`end`, calendar_events.`desc`, calendar_events.userId, calendar_event_rules.id, calendar_event_rules.eventId, calendar_event_rules.frequency, calendar_event_rules.interval, calendar_event_rules.byDay, calendar_event_rules.byMonthDay, calendar_event_rules.byMonth, calendar_event_rules.until FROM calendar_events "+
			"LEFT JOIN calendar_event_rules ON calendar_events.id = calendar_event_rules.eventId "+
			"WHERE calendar_events.userId = ? AND ((calendar_events.`end` >= ? AND calendar_events.`start` <= ?) OR calendar_event_rules.frequency IS NOT NULL)",
		user.ID, startTime.Unix(), endTime.Unix(),
	)
	if err != nil {
		return View{}, err
	}
	defer plainEventRows.Close()

	for plainEventRows.Next() {
		event := data.Event{
			Type: data.EventTypePlain,
		}
		eventData := data.PlainEventData{}
		recurRule := data.RecurRule{
			ID: -1,
		}
		plainEventRows.Scan(
			&event.ID, &event.Name, &event.Start, &event.End, &eventData.Desc, &event.UserID,
			&recurRule.ID, &recurRule.EventID, &recurRule.Frequency, &recurRule.Interval, &recurRule.ByDayString, &recurRule.ByMonthDay, &recurRule.ByMonth, &recurRule.UntilString,
		)
		event.Data = eventData
		if recurRule.ID != -1 {
			event.RecurRule = &recurRule

			if event.RecurRule.UntilString == "2099-12-12" {
				// just a placeholder value for mysql, ignore it
				event.RecurRule.UntilString = ""
			} else {
				event.RecurRule.Until, err = time.Parse("2006-01-02", event.RecurRule.UntilString)
				if err != nil {
					return View{}, err
				}
			}
		}

		eventTimes := event.CalculateTimes(endTime)

		for _, eventTime := range eventTimes {
			dayOffset := int(math.Floor(eventTime.Sub(startTime).Hours() / 24))

			if dayOffset < 0 || dayOffset > len(view.Days)-1 {
				continue
			}

			view.Days[dayOffset].Events = append(view.Days[dayOffset].Events, event)
		}
	}

	// get homework events
	hwEventRows, err := db.Query("SELECT calendar_hwevents.id, homework.id, homework.name, homework.`due`, homework.`desc`, homework.`complete`, homework.classId, homework.userId, calendar_hwevents.`start`, calendar_hwevents.`end`, calendar_hwevents.userId FROM calendar_hwevents INNER JOIN homework ON calendar_hwevents.homeworkId = homework.id WHERE calendar_hwevents.userId = ? AND (calendar_hwevents.`end` >= ? AND calendar_hwevents.`start` <= ?)", user.ID, startTime.Unix(), endTime.Unix())
	if err != nil {
		return View{}, err
	}
	defer hwEventRows.Close()

	for hwEventRows.Next() {
		event := data.Event{
			Type: data.EventTypeHomework,
		}
		eventData := data.HomeworkEventData{}
		hwEventRows.Scan(&event.ID, &eventData.Homework.ID, &eventData.Homework.Name, &eventData.Homework.Due, &eventData.Homework.Desc, &eventData.Homework.Complete, &eventData.Homework.ClassID, &eventData.Homework.UserID, &event.Start, &event.End, &event.UserID)
		event.Data = eventData
		event.Name = eventData.Homework.Name

		eventStartTime := time.Unix(int64(event.Start), 0)
		dayOffset := int(math.Floor(eventStartTime.Sub(startTime).Hours() / 24))

		if dayOffset < 0 || dayOffset > len(view.Days)-1 {
			continue
		}

		view.Days[dayOffset].Events = append(view.Days[dayOffset].Events, event)
	}

	// get data from calendar providers
	for _, provider := range providers {
		providerData, err := provider.GetData(db, user, startTime, endTime, data.ProviderDataAll)
		if err != nil {
			return View{}, err
		}

		// add announcements
		for _, announcement := range providerData.Announcements {
			announcementDate, err := time.Parse("2006-01-02", announcement.Date)
			if err != nil {
				return View{}, err
			}
			dayOffset := int(math.Ceil(announcementDate.Sub(startTime).Hours() / 24))

			if dayOffset < 0 || dayOffset > len(view.Days)-1 {
				continue
			}

			view.Days[dayOffset].Announcements = append(view.Days[dayOffset].Announcements, announcement)
		}
	}

	// get schedule events
	for i := 0; i < dayCount; i++ {
		day := view.Days[i]

		if day.CurrentTerm == nil {
			continue
		}

		dayTime, _ := time.ParseInLocation("2006-01-02", day.DayString, location)
		dayOffset := int(dayTime.Unix())

		// is it candlelighting?
		if dayTime.Year() == Day_Candlelighting.Year() && dayTime.Month() == Day_Candlelighting.Month() && dayTime.Day() == Day_Candlelighting.Day() {
			itemsForPeriod := map[string][]data.Event{}
			seenClassIds := []int{}
			rows, err := db.Query("SELECT calendar_periods.id, calendar_classes.termId, calendar_classes.sectionId, calendar_classes.`name`, calendar_classes.ownerId, calendar_classes.ownerName, calendar_periods.dayNumber, calendar_periods.block, calendar_periods.buildingName, calendar_periods.roomNumber, calendar_periods.`start`, calendar_periods.`end`, calendar_periods.userId FROM calendar_periods INNER JOIN calendar_classes ON calendar_periods.classId = calendar_classes.sectionId WHERE calendar_periods.userId = ? AND (calendar_classes.termId = ? OR calendar_classes.termId = -1) AND calendar_periods.block IN ('C', 'D', 'H', 'G') GROUP BY calendar_periods.id, calendar_classes.termId, calendar_classes.name, calendar_classes.ownerId, calendar_classes.ownerName", user.ID, day.CurrentTerm.TermID)
			if err != nil {
				return View{}, err
			}
			defer rows.Close()
			for rows.Next() {
				event := data.Event{
					Type: data.EventTypeSchedule,
				}
				eventData := data.ScheduleEventData{}
				rows.Scan(&event.ID, &eventData.TermID, &eventData.ClassID, &event.Name, &eventData.OwnerID, &eventData.OwnerName, &eventData.DayNumber, &eventData.Block, &eventData.BuildingName, &eventData.RoomNumber, &event.Start, &event.End, &event.UserID)
				event.Data = eventData

				if !util.IntSliceContains(seenClassIds, eventData.ClassID) {
					_, sliceExists := itemsForPeriod[eventData.Block]
					if !sliceExists {
						itemsForPeriod[eventData.Block] = []data.Event{}
					}
					itemsForPeriod[eventData.Block] = append(itemsForPeriod[eventData.Block], event)
					seenClassIds = append(seenClassIds, eventData.ClassID)
				}
			}

			for _, specialScheduleItem := range SpecialSchedule_HS_Candlelighting {
				if specialScheduleItem.Block != "" {
					// it references a specific class, look it up
					items, haveItems := itemsForPeriod[specialScheduleItem.Block]
					if haveItems {
						for _, scheduleItem := range items {
							newEvent := scheduleItem
							newEvent.Start = dayOffset + specialScheduleItem.Start
							newEvent.End = dayOffset + specialScheduleItem.End

							// remove building + room number because we can't tell which one to use
							// (in the case of classes where the room changes, like most science classes)
							eventData := newEvent.Data.(data.ScheduleEventData)
							eventData.BuildingName = ""
							eventData.RoomNumber = ""
							newEvent.Data = eventData

							view.Days[i].Events = append(view.Days[i].Events, newEvent)
						}
					} else {
						// the person doesn't have a class for that period
						// just skip it
					}
				} else {
					// it's a fixed thing, just add it directly
					newEvent := data.Event{
						ID:    -1,
						Name:  specialScheduleItem.Name,
						Start: dayOffset + specialScheduleItem.Start,
						End:   dayOffset + specialScheduleItem.End,
						Type:  data.EventTypeSchedule,
						Data: data.ScheduleEventData{
							TermID:       day.CurrentTerm.TermID,
							ClassID:      -1,
							OwnerID:      -1,
							OwnerName:    "",
							DayNumber:    -1,
							Block:        "",
							BuildingName: "",
							RoomNumber:   "",
						},
						UserID: -1,
					}
					view.Days[i].Events = append(view.Days[i].Events, newEvent)
				}
			}
		}

		// check if it's an off day
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

		// calculate day index (1 = monday, 8 = friday 4)
		dayNumber := int(dayTime.Weekday())

		if dayTime.Weekday() == time.Friday {
			fridayNumber := -1
			for _, friday := range fridays {
				if day.DayString == friday.Date {
					fridayNumber = friday.Index
					break
				}
			}

			if fridayNumber != -1 {
				dayNumber = 4 + fridayNumber
			} else {
				continue
			}
		}

		if dayTime.Weekday() == time.Saturday || dayTime.Weekday() == time.Sunday {
			continue
		}

		rows, err := db.Query("SELECT calendar_periods.id, calendar_classes.termId, calendar_classes.sectionId, calendar_classes.`name`, calendar_classes.ownerId, calendar_classes.ownerName, calendar_periods.dayNumber, calendar_periods.block, calendar_periods.buildingName, calendar_periods.roomNumber, calendar_periods.`start`, calendar_periods.`end`, calendar_periods.userId FROM calendar_periods INNER JOIN calendar_classes ON calendar_periods.classId = calendar_classes.sectionId WHERE calendar_periods.userId = ? AND (calendar_classes.termId = ? OR calendar_classes.termId = -1) AND calendar_periods.dayNumber = ? GROUP BY calendar_periods.id, calendar_classes.termId, calendar_classes.name, calendar_classes.ownerId, calendar_classes.ownerName", user.ID, day.CurrentTerm.TermID, dayNumber)
		if err != nil {
			return View{}, err
		}
		defer rows.Close()
		for rows.Next() {
			event := data.Event{
				Type: data.EventTypeSchedule,
			}
			eventData := data.ScheduleEventData{}
			rows.Scan(&event.ID, &eventData.TermID, &eventData.ClassID, &event.Name, &eventData.OwnerID, &eventData.OwnerName, &eventData.DayNumber, &eventData.Block, &eventData.BuildingName, &eventData.RoomNumber, &event.Start, &event.End, &event.UserID)
			event.Data = eventData

			event.Start += dayOffset
			event.End += dayOffset

			view.Days[i].Events = append(view.Days[i].Events, event)
		}

		if dayTime.Weekday() == time.Thursday {
			// special case: assembly
			for eventIndex, event := range view.Days[i].Events {
				// check for an "HS House" event
				// starting 11:50, ending 12:50
				if strings.HasPrefix(event.Name, "HS House") && event.Start == int(dayTime.Unix())+42600 && event.End == int(dayTime.Unix())+46200 {
					// found it
					// now look up what type of assembly period it is this week
					assemblyType, foundType := AssemblyTypeList[dayTime.Format("2006-01-02")]

					if !foundType || assemblyType == AssemblyType_Assembly {
						// set name to assembly and room to Theater
						view.Days[i].Events[eventIndex].Name = "Assembly"
						eventData := view.Days[i].Events[eventIndex].Data.(data.ScheduleEventData)
						eventData.RoomNumber = "Theater"
						view.Days[i].Events[eventIndex].Data = eventData
					} else if assemblyType == AssemblyType_LongHouse {
						// set name to long house
						view.Days[i].Events[eventIndex].Name = "Long House"
					} else if assemblyType == AssemblyType_Lab {
						// just remove it
						view.Days[i].Events = append(view.Days[i].Events[:eventIndex], view.Days[i].Events[eventIndex+1:]...)
					}
				}
			}
		}
	}

	if viewIncludesSpecialAssessmentDay {
		// get a list of the user's calendar classes
		sectionIDs := []int{}
		classRows, err := db.Query("SELECT sectionId FROM calendar_classes WHERE userId = ? GROUP BY `sectionId`", user.ID)
		if err != nil {
			return View{}, err
		}
		defer classRows.Close()
		for classRows.Next() {
			sectionID := -1
			classRows.Scan(&sectionID)
			sectionIDs = append(sectionIDs, sectionID)
		}

		// find the applicable special assessments
		allSpecialAssessments := []*SpecialAssessmentInfo{}
		for _, sectionID := range sectionIDs {
			specialAssessment, foundAssessment := SpecialAssessmentList[sectionID]
			if !foundAssessment {
				// no assessment for this class
				continue
			}

			isDuplicate := false
			for _, alreadyFoundSpecialAssessment := range allSpecialAssessments {
				if specialAssessment == alreadyFoundSpecialAssessment {
					isDuplicate = true
					break
				}
			}
			if isDuplicate {
				continue
			}

			allSpecialAssessments = append(allSpecialAssessments, specialAssessment)
		}

		for i := 0; i < dayCount; i++ {
			day := view.Days[i]

			dayType := SpecialAssessmentType_Unknown

			for specialAssessmentDay, specialAssessmentDayType := range SpecialAssessmentDays {
				if day.DayString == specialAssessmentDay {
					dayType = specialAssessmentDayType
					break
				}
			}

			if dayType == SpecialAssessmentType_Unknown {
				continue
			}

			var assessmentForDay *SpecialAssessmentInfo
			for _, assessment := range allSpecialAssessments {
				if assessment.Subject == dayType {
					assessmentForDay = assessment
					break
				}
			}

			if assessmentForDay == nil {
				continue
			}

			event := data.Event{
				Type:   data.EventTypeSchedule,
				ID:     -1,
				Name:   fmt.Sprintf("Final - %s", assessmentForDay.ClassName),
				Start:  assessmentForDay.Start,
				End:    assessmentForDay.End,
				UserID: user.ID,
			}

			finalDay := startTime.Add(time.Duration(i) * 24 * time.Hour)

			// hacky time correction to shift the timezone properly
			startHour := int(math.Floor(float64(event.Start) / 60 / 60))
			startMin := int(math.Floor((float64(event.Start) - (float64(startHour) * 60 * 60)) / 60))

			event.Start = int(time.Date(finalDay.Year(), finalDay.Month(), finalDay.Day(), startHour, startMin, 0, 0, location).Unix())

			endHour := int(math.Floor(float64(event.End) / 60 / 60))
			endMin := int(math.Floor((float64(event.End) - (float64(endHour) * 60 * 60)) / 60))

			event.End = int(time.Date(finalDay.Year(), finalDay.Month(), finalDay.Day(), endHour, endMin, 0, 0, location).Unix())

			eventData := data.ScheduleEventData{
				TermID:       -1,
				ClassID:      -1,
				OwnerID:      -1,
				OwnerName:    assessmentForDay.TeacherName,
				DayNumber:    -1,
				Block:        "",
				BuildingName: "",
				RoomNumber:   assessmentForDay.RoomNumber,
			}
			event.Data = eventData

			view.Days[i].Events = append(view.Days[i].Events, event)
		}
	}

	return view, nil
}
