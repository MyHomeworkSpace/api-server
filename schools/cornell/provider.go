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

type holiday struct {
	ID         int
	StartDate  string
	EndDate    string
	Name       string
	HasClasses bool
}

func (p *provider) GetData(db *sql.DB, user *data.User, location *time.Location, startTime time.Time, endTime time.Time, dataType data.ProviderDataType) (data.ProviderData, error) {
	results := data.ProviderData{}
	// startTimeISO8601 := startTime.Format("2006-01-02")
	// endTimeISO8601 := endTime.Format("2006-01-02")

	rows, err := db.Query("SELECT id, startDate, endDate, name, hasClasses FROM cornell_holidays")
	if err != nil {
		return data.ProviderData{}, err
	}

	holidays := []holiday{}

	defer rows.Close()

	for rows.Next() {
		hasClassesInt := -1
		h := holiday{}
		err = rows.Scan(&h.ID, &h.StartDate, &h.EndDate, &h.Name, &hasClassesInt)
		if err != nil {
			return data.ProviderData{}, err
		}
		h.HasClasses = hasClassesInt == 1
		holidays = append(holidays, h)
	}

	if dataType&data.ProviderDataEvents != 0 {
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
				hasHoliday, currentHoliday, err := getHolidayForDate(currentDate, holidays)
				if err != nil {
					return data.ProviderData{}, err
				}
				if (hasHoliday && !currentHoliday.HasClasses) || startDateTime.After(currentDate) {
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
									URL:  "https://classes.cornell.edu/browse/roster/" + GetCurrentTerm().Code + "/class/" + subject + "/" + catalogNum,
								},
							},
							data.EventTagSection: section,
						},
					}

					if subject == "CS" {
						// all CS course websites follow the same structure
						actions := event.Tags[data.EventTagActions]

						actions = append(actions.([]data.EventAction), data.EventAction{
							Icon: "external-link",
							Name: "Course Website",
							URL:  "https://courses.cs.cornell.edu/cs" + catalogNum + "/" + GetCurrentTerm().CSCode + "/",
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

	if dataType&data.ProviderDataAnnouncements != 0 {
		announcements := []data.PlannerAnnouncement{}

		currentDate := startTime
		for currentDate.Before(endTime) {
			hasHoliday, currentHoliday, err := getHolidayForDate(currentDate, holidays)
			if err != nil {
				return data.ProviderData{}, err
			}

			if hasHoliday {
				announcement := data.PlannerAnnouncement{
					ID:   currentHoliday.ID,
					Date: currentDate.Format("2006-01-02"),
					Text: currentHoliday.Name,
				}
				announcements = append(announcements, announcement)
			}

			currentDate = currentDate.Add(time.Hour * 24)
		}

		results.Announcements = announcements
	}
	return results, nil
}

func getHolidayForDate(date time.Time, holidays []holiday) (bool, holiday, error) {
	for _, h := range holidays {
		hStartDate, err := time.Parse("2006-01-02", h.StartDate)
		hEndDate, err := time.Parse("2006-01-02", h.EndDate)
		if err != nil {
			return false, holiday{}, err
		}

		hEndDate = hEndDate.Add(24 * time.Hour) // we need to do that because both sides are inclusive

		if hStartDate.Before(date) && hEndDate.After(date) {
			return true, h, nil
		}
	}
	return false, holiday{}, nil
}
