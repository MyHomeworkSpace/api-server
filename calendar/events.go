package calendar

import (
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
)

// An EventType represents the type of a calendar event: plain, homework, or schedule.
type EventType int

// An SpecialAssessmentType represents the subject of a special event.
type SpecialAssessmentType int

// A SpecialScheduleItem represents an item on a day with a special schedule, such as as Candlelighting.
type SpecialScheduleItem struct {
	Block string
	Name  string
	Start int
	End   int
}

// The available event types.
const (
	PlainEvent EventType = iota
	HomeworkEvent
	ScheduleEvent
)

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

// An Event is an event on a user's calendar. It could be from their schedule, homework, or manually added.
type Event struct {
	ID     int         `json:"id"`
	Name   string      `json:"name"`
	Start  int         `json:"start"`
	End    int         `json:"end"`
	Type   EventType   `json:"type"`
	Data   interface{} `json:"data"`
	UserID int         `json:"userId"`
}

// PlainEventData stores additional data associated with a plain event.
type PlainEventData struct {
	Desc string `json:"desc"`
}

// HomeworkEventData stores additional data associated with a homework event.
type HomeworkEventData struct {
	Homework data.Homework `json:"homework"`
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

// SpecialAssessmentInfo stores information related to a special assessment (midterm or final). Used for the internal, server-side list.
type SpecialAssessmentInfo struct {
	Subject     SpecialAssessmentType `json:"subject"`
	Start       int                   `json:"start"`
	End         int                   `json:"end"`
	ClassName   string                `json:"className"`
	TeacherName string                `json:"teacherName"`
	RoomNumber  string                `json:"roomNumber"`
}
