package calendar

import (
	"time"
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
