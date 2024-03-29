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

// An EventTagType describes the type of an event tag.
type EventTagType int

// The available event tags.
const (
	EventTagReserved EventTagType = iota
	EventTagDescription
	EventTagHomework
	EventTagTermID
	EventTagClassID
	EventTagOwnerID
	EventTagOwnerName
	EventTagDayNumber
	EventTagBlock
	EventTagBuildingName
	EventTagRoomNumber
	EventTagLocation
	EventTagReadOnly
	EventTagShortName
	EventTagActions
	EventTagCancelled
	EventTagCancelable
	EventTagSection
	EventTagOriginalStart
	EventTagOriginalEnd
	EventTagHideBuildingName
	EventTagHomeworkClass
	EventTagInstanceStart
	EventTagInstanceEnd
	EventTagIsContinuation
	EventTagContinues
)

// An Event is an event on a user's calendar. It could be from their schedule, homework, or manually added.
type Event struct {
	ID            int                          `json:"id"`
	UniqueID      string                       `json:"uniqueId"`
	SeriesID      string                       `json:"seriesId"`
	SeriesName    string                       `json:"seriesName"`
	Name          string                       `json:"name"`
	Start         int                          `json:"start"`
	End           int                          `json:"end"`
	StartTimezone string                       `json:"startTimezone"`
	EndTimezone   string                       `json:"endTimezone"`
	RecurRule     *RecurRule                   `json:"recurRule"`
	Tags          map[EventTagType]interface{} `json:"tags"`
	Source        int                          `json:"source"`
	UserID        int                          `json:"userId"`
}

// An EventAction is an action that can be performed on an Event; for example, a link to open a class's website.
type EventAction struct {
	Icon string `json:"icon"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// An EventChange is a modification a user makes to an Event that came from a provider.
type EventChange struct {
	EventID string `json:"eventID"`
	Cancel  bool   `json:"cancel"`
	UserID  int    `json:"userID"`
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
	Until       string         `json:"until"`
}

// CalculateTimes returns a list of all times the given event will take place, using its RecurRule information.
func (e *Event) CalculateTimes(until time.Time) ([]time.Time, error) {
	eventStartTime := time.Unix(int64(e.Start), 0).UTC()

	if e.StartTimezone != "" {
		location, err := time.LoadLocation(e.StartTimezone)
		if err != nil {
			return nil, err
		}
		eventStartTime = eventStartTime.In(location)
	}

	eventTimes := []time.Time{}

	// obviously it has to happen at least once
	eventTimes = append(eventTimes, eventStartTime)

	if e.RecurRule != nil {
		currentTime := eventStartTime

		var ruleUntilTime time.Time
		haveRuleUntilTime := false
		if e.RecurRule.Until != "" {
			haveRuleUntilTime = true

			location, err := time.LoadLocation(e.EndTimezone)
			if err != nil {
				return nil, err
			}

			ruleUntilTime, err = time.ParseInLocation("2006-01-02", e.RecurRule.Until, location)
			if err != nil {
				return nil, err
			}
		}

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
			} else { // if e.RecurRule.Frequency == RecurFrequencyYearly {
				years = 1
			}

			years *= e.RecurRule.Interval
			months *= e.RecurRule.Interval
			days *= e.RecurRule.Interval

			previousTime := currentTime
			currentTime = currentTime.AddDate(years, months, days)

			if previousTime == currentTime {
				// we're not making progress, escape
				break
			}

			if haveRuleUntilTime {
				if ruleUntilTime.Sub(currentTime) < -24*time.Hour {
					break
				}
			}

			eventTimes = append(eventTimes, currentTime)
		}
	}

	return eventTimes, nil
}
