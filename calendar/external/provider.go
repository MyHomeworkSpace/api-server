package external

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
)

// A Provider that implements the data.Provider interface for an external calendar.
type Provider struct {
	ExternalCalendarID   int
	ExternalCalendarName string
	ExternalCalendarURL  string
}

type externalEvent struct {
	UID        string
	Name       string
	Start      int64
	End        int64
	CalendarID int
}

// ID returns the ID of the Provider.
func (p *Provider) ID() string {
	return "calendar-external-" + strconv.Itoa(p.ExternalCalendarID)
}

// Name returns the name of the Provider.
func (p *Provider) Name() string {
	return p.ExternalCalendarName
}

// GetData gets the requested calendar data from the provider.
func (p *Provider) GetData(db *sql.DB, user *data.User, location *time.Location, startTime time.Time, endTime time.Time, dataType data.ProviderDataType) (data.ProviderData, error) {
	startUnix := startTime.Unix()
	endUnix := endTime.Unix()

	rows, err := db.Query(
		"SELECT uid, name, start, end, calendarID FROM calendar_external_events WHERE calendarID = ? AND start >= ? AND end <= ?",
		p.ExternalCalendarID,
		startUnix,
		endUnix,
	)
	if err != nil {
		return data.ProviderData{}, err
	}

	result := data.ProviderData{
		Announcements: []data.PlannerAnnouncement{},
		Events:        []data.Event{},
	}

	for rows.Next() {
		externalEvent := externalEvent{}
		err = rows.Scan(&externalEvent.UID, &externalEvent.Name, &externalEvent.Start, &externalEvent.End, &externalEvent.CalendarID)
		if err != nil {
			return data.ProviderData{}, err
		}

		event := data.Event{
			ID:        -1,
			UniqueID:  externalEvent.UID,
			Name:      externalEvent.Name,
			Start:     int(externalEvent.Start),
			End:       int(externalEvent.End),
			RecurRule: nil,
			Tags: map[data.EventTagType]interface{}{
				data.EventTagReadOnly: true,
			},
		}

		result.Events = append(result.Events, event)
	}

	return result, nil
}
