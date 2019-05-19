package data

import (
	"database/sql"
	"time"
)

// ProviderDataType describes the different types of data that can be requested from a Provider.
type ProviderDataType uint8

const (
	ProviderDataAnnouncements = (1 << 0)
	ProviderDataEvents        = (1 << 1)

	ProviderDataAll = ProviderDataAnnouncements | ProviderDataEvents
)

// A Provider is a source of calendar data (events, announcements, etc)
type Provider interface {
	Name() string
	GetData(db *sql.DB, user *User, startTime time.Time, endTime time.Time, dataType ProviderDataType) (ProviderData, error)
}

// A ProviderData struct contains all data returned by a Provider for a given time
type ProviderData struct {
	Announcements []PlannerAnnouncement `json:"announcements"`
	Events        []Event               `json:"events"`
}
