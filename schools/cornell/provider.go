package cornell

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MyHomeworkSpace/api-server/config"
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

			startDateTime, err := time.Parse("2006-01-02", startDate)
			endDateTime, err := time.Parse("2006-01-02", endDate)
			if err != nil {
				return data.ProviderData{}, err
			}
			for currentDate.Before(endTime) {
				iso8601CurrentDate := currentDate.Format("2006-01-02")
				hasClasses := 1
				db.QueryRow("SELECT hasClasses FROM cornell_holidays WHERE startDate <= DATE(?) AND endDate >= DATE(?)", iso8601CurrentDate, iso8601CurrentDate).Scan(&hasClasses)
				if startDateTime.After(currentDate) || hasClasses == 0 {
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
							data.EventTagActions: []data.EventAction{
								{
									Icon: "external-link",
									Name: "View Roster",
									URL:  "https://classes.cornell.edu/browse/roster/" + config.GetCurrent().Cornell.CurrentTerm + "/class/" + subject + "/" + catalogNum,
								},
							},
						},
					}

					if subject == "CS" {
						// all CS course websites follow the same structure
						actions := event.Tags[data.EventTagActions]

						actions = append(actions.([]data.EventAction), data.EventAction{
							Icon: "external-link",
							Name: "Course Website",
							URL:  "https://courses.cs.cornell.edu/cs" + catalogNum + "/" + config.GetCurrent().Cornell.CurrentCSTerm + "/",
						})

						event.Tags[data.EventTagActions] = actions
					}

					events = append(events, event)
				}
				currentDate = currentDate.Add(time.Hour * 24)
			}
		}

		results.Events = events
	}

	if dataType == data.ProviderDataAnnouncements || dataType == data.ProviderDataAll {
		announcements := []data.PlannerAnnouncement{}

		currentDate := startTime
		for currentDate.Before(endTime) {
			iso8601CurrentDate := currentDate.Format("2006-01-02")
			name := ""
			ID := -1
			errNoRows := db.QueryRow("SELECT id, name FROM cornell_holidays WHERE startDate <= DATE(?) AND endDate >= DATE(?)", iso8601CurrentDate, iso8601CurrentDate).Scan(&ID, &name)

			if errNoRows == nil {
				announcement := data.PlannerAnnouncement{
					ID:   ID,
					Date: iso8601CurrentDate,
					Text: name,
				}
				announcements = append(announcements, announcement)
			} else if errNoRows != sql.ErrNoRows {
				return data.ProviderData{}, errNoRows
			}

			currentDate = currentDate.Add(time.Hour * 24)
		}

		results.Announcements = announcements
	}
	return results, nil
}
