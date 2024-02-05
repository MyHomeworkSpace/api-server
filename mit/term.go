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
			ExceptionDays: map[string]time.Weekday{
				// Mar 9: Monday schedule of classes to be held.
				"2021-03-09": time.Monday,
			},
		}, nil
	} else if code == "2022FA" {
		return TermInfo{
			Code:              "2022FA",
			FirstDayOfClasses: time.Date(2021, 9, 8, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2021, 12, 9, 0, 0, 0, 0, time.UTC),
			ExceptionDays:     map[string]time.Weekday{},
		}, nil
	} else if code == "2022JA" {
		return TermInfo{
			Code:              "2022JA",
			FirstDayOfClasses: time.Date(2022, 1, 3, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2022, 1, 28, 0, 0, 0, 0, time.UTC),
			ExceptionDays:     map[string]time.Weekday{},
		}, nil
	} else if code == "2022SP" {
		return TermInfo{
			Code:              "2022SP",
			FirstDayOfClasses: time.Date(2022, 1, 31, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2022, 5, 10, 0, 0, 0, 0, time.UTC),
			ExceptionDays: map[string]time.Weekday{
				// Feb 22: Monday schedule of classes to be held.
				"2022-02-22": time.Monday,
			},
		}, nil
	} else if code == "2023FA" {
		return TermInfo{
			Code:              "2023FA",
			FirstDayOfClasses: time.Date(2022, 9, 7, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2022, 12, 14, 0, 0, 0, 0, time.UTC),
			ExceptionDays:     map[string]time.Weekday{},
		}, nil
	} else if code == "2023JA" {
		return TermInfo{
			Code:              "2023JA",
			FirstDayOfClasses: time.Date(2023, 1, 9, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2023, 2, 3, 0, 0, 0, 0, time.UTC),
			ExceptionDays:     map[string]time.Weekday{},
		}, nil
	} else if code == "2023SP" {
		return TermInfo{
			Code:              "2023SP",
			FirstDayOfClasses: time.Date(2023, 2, 6, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2023, 5, 16, 0, 0, 0, 0, time.UTC),
			ExceptionDays: map[string]time.Weekday{
				// Feb 21: Monday schedule of classes to be held.
				"2023-02-21": time.Monday,
			},
		}, nil
	} else if code == "2024FA" {
		return TermInfo{
			Code:              "2024FA",
			FirstDayOfClasses: time.Date(2023, 9, 6, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2023, 12, 13, 0, 0, 0, 0, time.UTC),
			ExceptionDays:     map[string]time.Weekday{},
		}, nil
	} else if code == "2024SP" {
		return TermInfo{
			Code:              "2024SP",
			FirstDayOfClasses: time.Date(2024, 2, 5, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2024, 5, 14, 0, 0, 0, 0, time.UTC),
			ExceptionDays: map[string]time.Weekday{
				// Feb 20: Monday schedule of classes to be held.
				"2024-02-20": time.Monday,
			},
		}, nil
	}

	return TermInfo{}, ErrBadTermCode
}

// GetCurrentTerm returns a TermInfo struct for the current academic term.
func GetCurrentTerm() TermInfo {
	term, _ := GetTermByCode("2024SP")
	return term
}
