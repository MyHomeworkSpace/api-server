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
	ID() string
	Name() string
	GetData(db *sql.DB, user *User, location *time.Location, startTime time.Time, endTime time.Time, dataType ProviderDataType) (ProviderData, error)
}

// A ProviderData struct contains all data returned by a Provider for a given time
type ProviderData struct {
	Announcements []PlannerAnnouncement `json:"announcements"`
	Events        []Event               `json:"events"`
}

// GetProvidersForUser returns a list of calendar providers associated with the given user
func GetProvidersForUser(db *sql.DB, user *User) ([]Provider, error) {
	schools, err := GetSchoolsForUser(user)
	if err != nil {
		return nil, err
	}

	providers := []Provider{}
	for _, school := range schools {
		needsUpdate, err := school.NeedsUpdate(db)
		if err != nil {
			return nil, err
		}

		if needsUpdate {
			continue
		}

		providers = append(providers, school.CalendarProvider())
	}

	return providers, nil
}
