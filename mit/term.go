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
}

// GetTermByCode returns the TermInfo struct for the term with the given code, or ErrBadTermCode if the term couldn't be found.
func GetTermByCode(code string) (TermInfo, error) {
	if code == "2020FA" {
		return TermInfo{
			Code:              "2020FA",
			FirstDayOfClasses: time.Date(2019, 9, 4, 0, 0, 0, 0, time.UTC),
			LastDayOfClasses:  time.Date(2019, 12, 11, 0, 0, 0, 0, time.UTC),
		}, nil
	}

	return TermInfo{}, ErrBadTermCode
}

// GetCurrentTerm returns a TermInfo struct for the current academic term.
func GetCurrentTerm() TermInfo {
	term, _ := GetTermByCode("2020FA")
	return term
}
