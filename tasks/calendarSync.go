package tasks

import (
	"database/sql"
	"time"

	"github.com/MyHomeworkSpace/api-server/calendar/external"
)

// StartCalendarSync begins syncing all external calendar.
func StartCalendarSync(db *sql.DB) error {
	go taskWatcher("calendar_sync", "Calendar sync", calendarSync, "", db)
	return nil
}

func calendarSync(lastCompletion *time.Time, source string, db *sql.DB) (taskResponse, error) {
	// load all external calendars
	rows, err := db.Query("SELECT id, name, url FROM calendar_external WHERE enabled = 1")
	if err != nil {
		return taskResponse{}, err
	}
	externalCalendars := []external.Provider{}
	for rows.Next() {
		externalCalendar := external.Provider{}
		err = rows.Scan(&externalCalendar.ExternalCalendarID, &externalCalendar.ExternalCalendarName, &externalCalendar.ExternalCalendarURL)
		if err != nil {
			return taskResponse{}, err
		}
		externalCalendars = append(externalCalendars, externalCalendar)
	}
	rows.Close()

	for _, externalCalendar := range externalCalendars {
		tx, err := db.Begin()
		if err != nil {
			return taskResponse{}, err
		}

		err = externalCalendar.Update(tx)
		if err != nil {
			return taskResponse{}, err
		}

		err = tx.Commit()
		if err != nil {
			return taskResponse{}, err
		}
	}

	return taskResponse{
		RowsAffected: 0,
	}, nil
}
