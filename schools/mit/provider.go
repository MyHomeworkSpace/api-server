package mit

import (
	"database/sql"
	"time"

	"github.com/MyHomeworkSpace/api-server/util"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

type provider struct {
	schools.Provider
}

func (p *provider) GetData(db *sql.DB, user *data.User, location *time.Location, startTime time.Time, endTime time.Time, dataType data.ProviderDataType) (data.ProviderData, error) {
	result := data.ProviderData{
		Announcements: nil,
		Events:        []data.Event{},
	}

	school := (p.Provider.School).(*school)

	dayCount := int((endTime.Sub(startTime).Hours() / 24) + 0.5)

	if school.peInfo != nil {
		// we have a PE class
		peInfo := *school.peInfo

		peFirstDay, err := time.ParseInLocation("2006-01-02", peInfo.ParsedFirstDay, location)
		if err != nil {
			return data.ProviderData{}, err
		}

		peLastDay, err := time.ParseInLocation("2006-01-02", peInfo.ParsedLastDay, location)
		if err != nil {
			return data.ProviderData{}, err
		}

		currentDay := startTime
		for i := 0; i < dayCount; i++ {
			if i != 0 {
				currentDay = currentDay.Add(24 * time.Hour)
			}

			if currentDay.Before(peFirstDay) {
				continue
			}

			if currentDay.After(peLastDay) {
				continue
			}

			foundWeekday := false
			for _, weekday := range peInfo.ParsedDaysOfWeek {
				if currentDay.Weekday() == weekday {
					foundWeekday = true
					break
				}
			}
			if !foundWeekday {
				continue
			}

			dayString := currentDay.Format("2006-01-02")
			dayTime, _ := time.ParseInLocation("2006-01-02", dayString, location)
			dayOffset := int(dayTime.Unix())

			if peInfo.ParsedSkipDays != nil {
				if util.StringSliceContains(peInfo.ParsedSkipDays, dayString) {
					continue
				}
			}

			event := data.Event{
				Tags: map[data.EventTagType]interface{}{},
			}

			event.ID = -1
			event.Name = peInfo.SectionID + " - " + peInfo.Activity + " - " + peInfo.CourseTitle

			event.Tags[data.EventTagReadOnly] = true
			event.Tags[data.EventTagLocation] = peInfo.ParsedLocation

			event.Start = peInfo.ParsedStartTime
			event.End = peInfo.ParsedEndTime

			event.Start += dayOffset
			event.End += dayOffset

			result.Events = append(result.Events, event)
		}
	}

	return result, nil
}
