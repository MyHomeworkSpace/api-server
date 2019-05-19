package calendar

import (
	"time"
)

// An SpecialAssessmentType represents the subject of a special event.
type SpecialAssessmentType int

// A SpecialScheduleItem represents an item on a day with a special schedule, such as as Candlelighting.
type SpecialScheduleItem struct {
	Block string
	Name  string
	Start int
	End   int
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

type Term struct {
	ID     int    `json:"id"`
	TermID int    `json:"termId"`
	Name   string `json:"name"`
	UserID int    `json:"userId"`
}

// SpecialAssessmentInfo stores information related to a special assessment (midterm or final). Used for the internal, server-side list.
type SpecialAssessmentInfo struct {
	Subject     SpecialAssessmentType `json:"subject"`
	Start       int                   `json:"start"`
	End         int                   `json:"end"`
	ClassName   string                `json:"className"`
	TeacherName string                `json:"teacherName"`
	RoomNumber  string                `json:"roomNumber"`
}
