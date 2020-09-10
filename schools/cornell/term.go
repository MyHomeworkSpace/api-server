package cornell

import "errors"

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
	}
	return TermInfo{}, ErrBadTermCode
}

// GetCurrentTerm returns a TermInfo struct for the current academic term.
func GetCurrentTerm() TermInfo {
	term, _ := GetTermByCode("FA20")
	return term
}
