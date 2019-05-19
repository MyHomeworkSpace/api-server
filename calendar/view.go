package calendar

import (
	"database/sql"
	"math"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools/manager"
)

// A ViewDay represents a day in a View.
type ViewDay struct {
	DayString     string                     `json:"day"`
	Announcements []data.PlannerAnnouncement `json:"announcements"`
	Events        []data.Event               `json:"events"`
}

// A View represents a view of a user's calendar over a certain period of time.
type View struct {
	Providers []ProviderInfo `json:"providers"`
	Days      []ViewDay      `json:"days"`
}

// A ProviderInfo struct represents information about an active calendar providers.
type ProviderInfo struct {
	Name string `json:"name"`
}

// GetView retrieves a CalendarView for the given user with the given parameters.
func GetView(db *sql.DB, user *data.User, location *time.Location, startTime time.Time, endTime time.Time) (View, error) {
	view := View{
		Providers: []ProviderInfo{},
		Days:      []ViewDay{},
	}

	providers := []data.Provider{
		// TODO: not hardcode this for dalton
		manager.GetSchoolByID("dalton").CalendarProvider(),
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

	// handle calendar providers
	for _, provider := range providers {
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

			view.Days[dayOffset].Events = append(view.Days[dayOffset].Events, event)
		}
	}

	return view, nil
}
