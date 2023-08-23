package calendar

import (
	"database/sql"
	"math"
	"strconv"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"

	set "github.com/deckarep/golang-set"
)

// A ViewDay represents a day in a View.
type ViewDay struct {
	DayString     string                     `json:"day"`
	Announcements []data.PlannerAnnouncement `json:"announcements"`
	Events        []data.Event               `json:"events"`
}

// A View represents a view of a user's calendar over a certain period of time.
type View struct {
	Providers       []ProviderInfo    `json:"providers"`
	SchoolsToUpdate []data.SchoolInfo `json:"schoolsToUpdate"`
	Days            []ViewDay         `json:"days"`
}

// A ProviderInfo struct represents information about an active calendar providers.
type ProviderInfo struct {
	Name string `json:"name"`
}

func doesEventInstanceOccurInTimeframe(instanceStart time.Time, duration time.Duration, startTime time.Time, endTime time.Time) bool {
	instanceEnd := instanceStart.Add(duration)

	// if the start time is in range, show the event
	if instanceStart.After(startTime) && instanceStart.Before(endTime) {
		return true
	}

	// if the end time is in range, show the event
	if instanceEnd.After(startTime) && instanceEnd.Before(endTime) {
		return true
	}

	// if the event encompasses our range, show it
	if instanceStart.Before(startTime) && instanceEnd.After(endTime) {
		return true
	}

	return false
}

func addEventToView(view *View, event data.Event, eventTime time.Time, eventDuration time.Duration, startTime time.Time, endTime time.Time) {
	dayOffset := int(math.Floor(eventTime.Sub(startTime).Hours() / 24))

	durationLeft := eventDuration
	currentStart := eventTime
	for {
		day := startTime.AddDate(0, 0, dayOffset)
		endOfDayTime := time.Date(
			day.Year(),
			day.Month(),
			day.Day()+1,
			0,
			0,
			0,
			0,
			day.Location(),
		)

		if dayOffset > len(view.Days)-1 {
			// no more chunks to add, we've run out of days
			break
		}

		if durationLeft <= 0 {
			// no more duration to add, we're done
			break
		}

		if dayOffset < 0 {
			// out of range, need to advance till we get in range
			// note that we still subtract time from the event's duration

			chunkDuration := endOfDayTime.Sub(currentStart)
			durationLeft -= chunkDuration

			currentStart = endOfDayTime
			dayOffset += 1
			continue
		}

		eventCopy := event

		eventCopy.Tags = map[data.EventTagType]interface{}{}
		for tag, value := range event.Tags {
			eventCopy.Tags[tag] = value
		}

		// calculate new end time for this chunk
		updatedEnd := currentStart.Add(durationLeft)
		if updatedEnd.After(endOfDayTime) {
			updatedEnd = endOfDayTime
		}

		// update how much time we have left to add
		chunkDuration := updatedEnd.Sub(currentStart)
		durationLeft -= chunkDuration

		// update chunk's start and end times
		eventCopy.Start = int(currentStart.Unix())
		eventCopy.End = int(updatedEnd.Unix())

		// leave full start/end in tags
		// note that a recurring event will also have OriginalStart and OriginalEnd, which takes precedent
		eventCopy.Tags[data.EventTagInstanceStart] = event.Start
		eventCopy.Tags[data.EventTagInstanceEnd] = event.End

		// if our start isn't the original start, then this is a continuation
		eventCopy.Tags[data.EventTagIsContinuation] = (currentStart != eventTime)

		// if there's more time left, then this event continues onto the next day
		eventCopy.Tags[data.EventTagContinues] = durationLeft > 0

		// actually add the event to the view
		view.Days[dayOffset].Events = append(view.Days[dayOffset].Events, eventCopy)

		currentStart = endOfDayTime
		dayOffset += 1
	}
}

// GetView retrieves a CalendarView for the given user with the given parameters.
func GetView(db *sql.DB, user *data.User, location *time.Location, startTime time.Time, endTime time.Time) (View, error) {
	view := View{
		Providers:       []ProviderInfo{},
		SchoolsToUpdate: []data.SchoolInfo{},
		Days:            []ViewDay{},
	}

	providers, err := data.GetProvidersForUser(db, user)
	if err != nil {
		return View{}, err
	}

	for _, schoolInfo := range user.Schools {
		if !schoolInfo.Enabled {
			continue
		}

		needsUpdate, err := schoolInfo.School.NeedsUpdate(db)
		if err != nil {
			return View{}, err
		}

		if needsUpdate {
			view.SchoolsToUpdate = append(view.SchoolsToUpdate, schoolInfo)
		}
	}

	// create days in array
	dayCount := int((endTime.Sub(startTime).Hours() / 24) + 0.5)
	currentDay := startTime
	for i := 0; i < dayCount; i++ {
		view.Days = append(view.Days, ViewDay{
			DayString:     currentDay.Format("2006-01-02"),
			Announcements: []data.PlannerAnnouncement{},
			Events:        []data.Event{},
		})

		currentDay = currentDay.AddDate(0, 0, 1)
	}

	// get plain events
	plainEventRows, err := db.Query(
		"SELECT calendar_events.id, calendar_events.name, calendar_events.`start`, calendar_events.`end`, calendar_events.location, calendar_events.`desc`, calendar_events.userId, calendar_event_rules.id, calendar_event_rules.eventId, calendar_event_rules.frequency, calendar_event_rules.interval, calendar_event_rules.byDay, calendar_event_rules.byMonthDay, calendar_event_rules.byMonth, calendar_event_rules.until FROM calendar_events "+
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
			StartTimezone: "America/New_York", // TODO: make this part of the event
			EndTimezone:   "America/New_York", // TODO: make this part of the event
			Tags:          map[data.EventTagType]interface{}{},
			Source:        -1,
		}
		location := ""
		desc := ""
		recurRule := data.RecurRule{
			ID: -1,
		}
		plainEventRows.Scan(
			&event.ID, &event.Name, &event.Start, &event.End, &location, &desc, &event.UserID,
			&recurRule.ID, &recurRule.EventID, &recurRule.Frequency, &recurRule.Interval, &recurRule.ByDayString, &recurRule.ByMonthDay, &recurRule.ByMonth, &recurRule.Until,
		)
		event.Tags[data.EventTagLocation] = location
		event.Tags[data.EventTagDescription] = desc
		if recurRule.ID != -1 {
			event.RecurRule = &recurRule

			if event.RecurRule.Until == "2099-12-12" {
				// just a placeholder value for mysql, ignore it
				event.RecurRule.Until = ""
			}

			event.Tags[data.EventTagCancelable] = true
			event.Tags[data.EventTagOriginalStart] = event.Start
			event.Tags[data.EventTagOriginalEnd] = event.End
		}

		eventTimes, err := event.CalculateTimes(endTime)
		if err != nil {
			return View{}, err
		}

		eventLength := time.Duration(event.End-event.Start) * time.Second

		for _, eventTime := range eventTimes {
			if !doesEventInstanceOccurInTimeframe(eventTime, eventLength, startTime, endTime) {
				continue
			}

			newTags := map[data.EventTagType]interface{}{}
			for tagType, tagValue := range event.Tags {
				newTags[tagType] = tagValue
			}
			event.Start = int(eventTime.Unix())
			event.End = int(eventTime.Add(eventLength).Unix())
			event.Tags = newTags
			event.UniqueID = "mhs-" + strconv.Itoa(event.ID) + "-" + eventTime.Format("2006-01-02")

			addEventToView(
				&view,
				event,
				eventTime,
				eventLength,
				startTime,
				endTime,
			)
		}
	}

	// get homework events
	hwEventRows, err := db.Query(
		"SELECT "+
			"calendar_hwevents.id, homework.id, homework.name, homework.`due`, homework.`desc`, homework.`complete`, homework.classId, homework.userId, calendar_hwevents.`start`, calendar_hwevents.`end`, calendar_hwevents.location, calendar_hwevents.`desc`, calendar_hwevents.userId, "+
			"classes.id, classes.name, classes.teacher, classes.color, classes.sortIndex, classes.userId "+
			"FROM calendar_hwevents "+
			"INNER JOIN homework ON calendar_hwevents.homeworkId = homework.id "+
			"INNER JOIN classes ON homework.classId = classes.id "+
			"WHERE calendar_hwevents.userId = ? AND (calendar_hwevents.`end` >= ? AND calendar_hwevents.`start` <= ?)",
		user.ID, startTime.Unix(), endTime.Unix(),
	)
	if err != nil {
		return View{}, err
	}
	defer hwEventRows.Close()

	for hwEventRows.Next() {
		event := data.Event{
			Tags:   map[data.EventTagType]interface{}{},
			Source: -1,
		}
		homework := data.Homework{}
		class := data.HomeworkClass{}
		location := ""
		desc := ""
		hwEventRows.Scan(
			&event.ID, &homework.ID, &homework.Name, &homework.Due, &homework.Desc, &homework.Complete, &homework.ClassID, &homework.UserID, &event.Start, &event.End, &location, &desc, &event.UserID,
			&class.ID, &class.Name, &class.Teacher, &class.Color, &class.SortIndex, &class.UserID,
		)
		event.UniqueID = "mhs-hw-" + strconv.Itoa(event.ID)
		event.Tags[data.EventTagHomework] = homework
		event.Tags[data.EventTagHomeworkClass] = class
		event.Tags[data.EventTagLocation] = location
		event.Tags[data.EventTagDescription] = desc
		event.Name = homework.Name

		eventStartTime := time.Unix(int64(event.Start), 0)
		dayOffset := int(math.Floor(eventStartTime.Sub(startTime).Hours() / 24))

		if dayOffset < 0 || dayOffset > len(view.Days)-1 {
			continue
		}

		view.Days[dayOffset].Events = append(view.Days[dayOffset].Events, event)
	}

	// handle calendar providers
	for providerIndex, provider := range providers {
		// add them to the list
		view.Providers = append(view.Providers, ProviderInfo{
			Name: provider.Name(),
		})

		// get data
		providerData, err := provider.GetData(db, user, location, startTime, endTime, data.ProviderDataAll)
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

		// add events
		for _, event := range providerData.Events {
			eventDate := time.Unix(int64(event.Start), 0)
			dayOffset := int(math.Floor(eventDate.Sub(startTime).Hours() / 24))

			if dayOffset < 0 || dayOffset > len(view.Days)-1 {
				continue
			}

			event.UniqueID = provider.ID() + "-" + event.UniqueID
			event.SeriesID = provider.ID() + "-" + event.SeriesID
			event.Source = providerIndex

			view.Days[dayOffset].Events = append(view.Days[dayOffset].Events, event)
		}
	}

	// apply any modifications made by the user
	eventChangeRows, err := db.Query("SELECT eventID, cancel FROM calendar_event_changes WHERE userID = ?", user.ID)
	if err != nil {
		return View{}, err
	}
	defer eventChangeRows.Close()

	cancellations := set.NewSet()

	for eventChangeRows.Next() {
		eventID, cancel := "", 0
		err = eventChangeRows.Scan(&eventID, &cancel)
		if err != nil {
			return View{}, err
		}

		if cancel == 1 {
			cancellations.Add(eventID)
		}
	}

	for dayIndex, day := range view.Days {
		for eventIndex, event := range day.Events {
			if cancellations.Contains(event.UniqueID) {
				view.Days[dayIndex].Events[eventIndex].Tags[data.EventTagCancelled] = true
			}
		}
	}

	return view, nil
}
