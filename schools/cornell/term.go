package cornell

import (
	"errors"
)

// ErrBadTermCode is reported when the given term code doesn't exist.
var ErrBadTermCode = errors.New("cornell: unknown term code")

// A TermInfo struct contains information about an academic term.
type TermInfo struct {
	Code   string //contains the current term code
	CSCode string //contains the term code used by the CS department
}

// GetTermByCode returns the TermInfo struct for the term with the given code, or ErrBadTermCode if the term couldn't be found.
func GetTermByCode(code string) (TermInfo, error) {
	if code == "FA20" {
		return TermInfo{
			Code:   "FA20",
			CSCode: "2020fa",
		}, nil
	} else if code == "WI21" {
		return TermInfo{
			Code:   "WI21",
			CSCode: "2021wi", // the CS department doesn't offer winter courses, but if they did, this is probably what the code would be
		}, nil
	} else if code == "SP21" {
		return TermInfo{
			Code:   "SP21",
			CSCode: "2021sp",
		}, nil
	} else if code == "FA21" {
		return TermInfo{
			Code:   "FA21",
			CSCode: "2021fa",
		}, nil
	} else if code == "SP22" {
		return TermInfo{
			Code:   "SP22",
			CSCode: "2022sp",
		}, nil
	} else if code == "FA22" {
		return TermInfo{
			Code:   "FA22",
			CSCode: "2022sp",
		}, nil
	}
	return TermInfo{}, ErrBadTermCode
}

// GetCurrentTerm returns a TermInfo struct for the current academic term.
func GetCurrentTerm() TermInfo {
	term, _ := GetTermByCode("FA22")
	return term
}
