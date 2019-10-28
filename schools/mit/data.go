package mit

import "time"

type peInfo struct {
	SectionID   string `json:"sectionID"`
	Activity    string `json:"activity"`
	CourseTitle string `json:"courseTitle"`

	RawSchedule string `json:"rawSchedule"`

	RawFirstDay      string `json:"rawFirstDay"`
	RawLastDay       string `json:"rawLastDay"`
	RawCalendarNotes string `json:"rawCalendarNotes"`

	ParsedDaysOfWeek []time.Weekday `json:"parsedDaysOfWeek"`
	ParsedStartTime  int            `json:"parsedStartTime"`
	ParsedEndTime    int            `json:"parsedEndTime"`
	ParsedLocation   string         `json:"parsedLocation"`

	ParsedFirstDay string   `json:"parsedFirstDay"`
	ParsedLastDay  string   `json:"parsedLastDay"`
	ParsedSkipDays []string `json:"parsedSkipDays"`
}
