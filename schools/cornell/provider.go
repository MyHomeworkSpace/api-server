package cornell

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

type provider struct {
	schools.Provider
}

func (p *provider) GetData(db *sql.DB, user *data.User, location *time.Location, startTime time.Time, endTime time.Time, dataType data.ProviderDataType) (data.ProviderData, error) {
	results := data.ProviderData{}
	if dataType == data.ProviderDataEvents || dataType == data.ProviderDataAll {
		events := []data.Event{}
		rows, err := db.Query("SELECT title, subject, catalogNum, component, componentLong, section, startDate, endDate, startTime, endTime, monday, tuesday, wednesday, thursday, friday, saturday, sunday, facilityLong FROM cornell_events WHERE userId = ?", user.ID)
		if err != nil {
			return data.ProviderData{}, err
		}
		defer rows.Close()

		for rows.Next() {
			currentDate := startTime

			var subject, catalogNum, component, componentLong, section, startDate, endDate, facility, title string
			var eventStartTime, eventEndTime, monday, tuesday, wednesday, thursday, friday, saturday, sunday int
			err := rows.Scan(
				&title,
				&subject,
				&catalogNum,
				&component,
				&componentLong,
				&section,
				&startDate,
				&endDate,
				&eventStartTime,
				&eventEndTime,
				&monday,
				&tuesday,
				&wednesday,
				&thursday,
				&friday,
				&saturday,
				&sunday,
				&facility,
			)
			if err != nil {
				return data.ProviderData{}, err
			}

			fmt.Printf("%s %s %s\n", subject, catalogNum, component)

			startDateTime, err := time.Parse("2006-01-02", startDate)
			endDateTime, err := time.Parse("2006-01-02", endDate)
			if err != nil {
				return data.ProviderData{}, err
			}
			for currentDate.Before(endTime) {
				if startDateTime.After(currentDate) {
					currentDate = currentDate.Add(time.Hour * 24)
					continue
				} else if endDateTime.Before(currentDate) {
					break
				}
				weekday := currentDate.Weekday()
				if (weekday == time.Monday && monday == 1) || (weekday == time.Tuesday && tuesday == 1) || (weekday == time.Wednesday && wednesday == 1) || (weekday == time.Thursday && thursday == 1) || (weekday == time.Friday && friday == 1) || (weekday == time.Saturday && saturday == 1) || (weekday == time.Sunday && sunday == 1) {
					event := data.Event{
						UniqueID: fmt.Sprintf("%s%s-%s-%d", subject, catalogNum, component, eventStartTime+int(time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, time.UTC).Unix())),
						Name:     fmt.Sprintf("%s (%s)", title, componentLong),
						Start:    eventStartTime + int(time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, location).Unix()),
						End:      eventEndTime + int(time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 0, 0, 0, 0, location).Unix()),
						Tags: map[data.EventTagType]interface{}{
							data.EventTagLocation:   facility,
							data.EventTagCancelable: true,
							data.EventTagShortName:  fmt.Sprintf("%s %s %s", subject, catalogNum, component),
							data.EventTagReadOnly:   true,
						},
					}
					events = append(events, event)
				}
				currentDate = currentDate.Add(time.Hour * 24)
			}
		}

		results.Events = events
	}
	return results, nil
}
