package data

import (
	"time"
)

// The available announcement types.
const (
	AnnouncementTypeText       = 0 // just informative
	AnnouncementTypeFullOff    = 1 // no classes at all
	AnnouncementTypeBreakStart = 2 // start of a break (inclusive of that day!)
	AnnouncementTypeBreakEnd   = 3 // end of a break (exclusive of that day!)
)

// A RecurFrequency describes how often an event recurs.
type RecurFrequency int

// Define the default RecurFrequencies.
const (
	RecurFrequencyDaily RecurFrequency = iota
	RecurFrequencyWeekly
	RecurFrequencyMonthly
	RecurFrequencyYearly
)

// An EventType represents the type of a calendar event: plain, homework, or schedule.
type EventType int

// The available event types.
const (
	EventTypePlain EventType = iota
	EventTypeHomework
	EventTypeSchedule
)

// An Event is an event on a user's calendar. It could be from their schedule, homework, or manually added.
type Event struct {
	ID        int         `json:"id"`
	Name      string      `json:"name"`
	Start     int         `json:"start"`
	End       int         `json:"end"`
	RecurRule *RecurRule  `json:"recurRule"`
	Type      EventType   `json:"type"`
	Data      interface{} `json:"data"`
	UserID    int         `json:"userId"`
}

// PlainEventData stores additional data associated with a plain event.
type PlainEventData struct {
	Desc string `json:"desc"`
}

// HomeworkEventData stores additional data associated with a homework event.
type HomeworkEventData struct {
	Homework Homework `json:"homework"`
}

// ScheduleEventData stores additional data associated with a schedule event.
type ScheduleEventData struct {
	TermID       int    `json:"termId"`
	ClassID      int    `json:"classId"`
	OwnerID      int    `json:"ownerId"`
	OwnerName    string `json:"ownerName"`
	DayNumber    int    `json:"dayNumber"`
	Block        string `json:"block"`
	BuildingName string `json:"buildingName"`
	RoomNumber   string `json:"roomNumber"`
}

// An OffBlock is a period of time that's marked off on a calendar, such as a holiday.
type OffBlock struct {
	StartID   int       `json:"startId"`
	EndID     int       `json:"endId"`
	Start     time.Time `json:"-"`
	End       time.Time `json:"-"`
	StartText string    `json:"start"`
	EndText   string    `json:"end"`
	Name      string    `json:"name"`
	Grade     int       `json:"grade"`
}

// A RecurRule struct contains information about how an event recurs. Inspired by the iCal RRULE system.
type RecurRule struct {
	ID          int            `json:"id"`
	EventID     int            `json:"eventId"`
	Frequency   RecurFrequency `json:"frequency"`
	Interval    int            `json:"interval"`
	ByDayString string         `json:"-"`
	ByDay       []time.Weekday `json:"byDay"`
	ByMonthDay  int            `json:"byMonthDay"`
	ByMonth     time.Month     `json:"byMonth"`
	UntilString string         `json:"until"`
	Until       time.Time      `json:"-"`
}

// CalculateTimes returns a list of all times the given event will take place, using its RecurRule information.
func (e *Event) CalculateTimes(until time.Time) []time.Time {
	eventStartTime := time.Unix(int64(e.Start), 0)

	eventTimes := []time.Time{}

	// obviously it has to happen at least once
	eventTimes = append(eventTimes, eventStartTime)

	if e.RecurRule != nil {
		currentTime := eventStartTime

		for currentTime.Before(until) {
			years := 0
			months := 0
			days := 0

			if e.RecurRule.Frequency == RecurFrequencyDaily {
				days = 1
			} else if e.RecurRule.Frequency == RecurFrequencyWeekly {
				days = 7
			} else if e.RecurRule.Frequency == RecurFrequencyMonthly {
				months = 1
			} else if e.RecurRule.Frequency == RecurFrequencyYearly {
				years = 0
			}

			years *= e.RecurRule.Interval
			months *= e.RecurRule.Interval
			days *= e.RecurRule.Interval

			currentTime = currentTime.AddDate(years, months, days)

			if e.RecurRule.UntilString != "" {
				ruleUntil := e.RecurRule.Until

				if ruleUntil.Sub(currentTime) < -24*time.Hour {
					break
				}
			}

			eventTimes = append(eventTimes, currentTime)
		}
	}

	return eventTimes
}
