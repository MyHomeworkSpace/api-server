package mit

import (
	"errors"
	"time"
)

// ErrBadTermCode is reported when the given term code doesn't exist.
var ErrBadTermCode = errors.New("mit: unknown term code")

// A TermInfo struct contains information about an academic term.
type TermInfo struct {
	Code              string
	FirstDayOfClasses time.Time
	LastDayOfClasses  time.Time
	ExceptionDays     map[string]time.Weekday
}

// GetTermByCode returns the TermInfo struct for the term with the given code, or ErrBadTermCode if the term couldn't be found.
func GetTermByCode(code string) (TermInfo, error) {
	if code == "2020FA" {
		return TermInfo{
			Code:              "2020FA",
			FirstDayOfClasses: time.Date(2019, 9, 4, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2019, 12, 11, 0, 0, 0, 0, time.UTC),
			ExceptionDays:     map[string]time.Weekday{},
		}, nil
	} else if code == "2020JA" {
		return TermInfo{
			Code:              "2020JA",
			FirstDayOfClasses: time.Date(2020, 1, 6, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2020, 1, 31, 0, 0, 0, 0, time.UTC),
			ExceptionDays:     map[string]time.Weekday{},
		}, nil
	} else if code == "2020SP" {
		return TermInfo{
			Code:              "2020SP",
			FirstDayOfClasses: time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2020, 5, 12, 0, 0, 0, 0, time.UTC),
			ExceptionDays: map[string]time.Weekday{
				// Feb 18: Monday schedule of classes to be held.
				"2020-02-18": time.Monday,
			},
		}, nil
	} else if code == "2021FA" {
		return TermInfo{
			Code:              "2021FA",
			FirstDayOfClasses: time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2020, 12, 9, 0, 0, 0, 0, time.UTC),
			ExceptionDays: map[string]time.Weekday{
				// Oct 13: Monday schedule of classes to be held.
				"2020-10-13": time.Monday,
			},
		}, nil
	} else if code == "2021JA" {
		return TermInfo{
			Code:              "2021JA",
			FirstDayOfClasses: time.Date(2021, 1, 4, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2021, 1, 29, 0, 0, 0, 0, time.UTC),
			ExceptionDays:     map[string]time.Weekday{},
		}, nil
	} else if code == "2021SP" {
		return TermInfo{
			Code:              "2021SP",
			FirstDayOfClasses: time.Date(2021, 2, 16, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2021, 5, 20, 0, 0, 0, 0, time.UTC),
			ExceptionDays:     map[string]time.Weekday{},
		}, nil
	}

	return TermInfo{}, ErrBadTermCode
}

// GetCurrentTerm returns a TermInfo struct for the current academic term.
func GetCurrentTerm() TermInfo {
	term, _ := GetTermByCode("2021SP")
	return term
}
