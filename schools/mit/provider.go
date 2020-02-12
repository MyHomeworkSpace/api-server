package mit

import (
	"database/sql"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/mit"
	"github.com/MyHomeworkSpace/api-server/schools"
	"github.com/MyHomeworkSpace/api-server/util"
)

var sectionCharToDisplayName = map[byte]string{
	'B': "lab",
	'D': "design",
	'L': "lecture",
	'R': "recitation",
}

type provider struct {
	schools.Provider
}

type offeringInfo struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Section string `json:"section"`
	Term    string `json:"term"`

	Time       string       `json:"time"`
	ParsedTime mit.TimeInfo `json:"parsedTime"`

	Place string `json:"place"`

	FacultyID   string `json:"facultyID"`
	FacultyName string `json:"facultyName"`
}

func (p *provider) GetData(db *sql.DB, user *data.User, location *time.Location, startTime time.Time, endTime time.Time, dataType data.ProviderDataType) (data.ProviderData, error) {
	result := data.ProviderData{
		Announcements: []data.PlannerAnnouncement{},
		Events:        []data.Event{},
	}

	school := (p.Provider.School).(*school)

	dayCount := int((endTime.Sub(startTime).Hours() / 24) + 0.5)

	// check for any holidays during the time period
	announcementRows, err := db.Query("SELECT id, date, text FROM mit_holidays WHERE date >= ? AND date <= ?", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
	if err != nil {
		return data.ProviderData{}, err
	}
	defer announcementRows.Close()
	for announcementRows.Next() {
		resp := data.PlannerAnnouncement{
			Type: data.AnnouncementTypeFullOff,
		}
		announcementRows.Scan(&resp.ID, &resp.Date, &resp.Text)
		result.Announcements = append(result.Announcements, resp)
	}

	// generate list of all off days in time period
	offDays := []string{}

	for _, announcement := range result.Announcements {
		if announcement.Type == data.AnnouncementTypeFullOff {
			offDays = append(offDays, announcement.Date)
		}
	}

	// using the user's registration, find when their classes are offered
	offerings := []offeringInfo{}
	offeringRows, err := db.Query(
		`SELECT
			mit_offerings.id, mit_listings.title, mit_offerings.section, mit_offerings.term, mit_offerings.time, mit_offerings.place, mit_offerings.facultyID, mit_offerings.facultyName, mit_classes.sections
		FROM mit_offerings
		INNER JOIN mit_classes ON mit_offerings.id = mit_classes.subjectID
		INNER JOIN mit_listings ON mit_offerings.id = mit_listings.id
		WHERE mit_classes.userID = ?`,
		user.ID)
	if err != nil {
		return data.ProviderData{}, err
	}
	defer offeringRows.Close()
	for offeringRows.Next() {
		selectedSections := ""

		info := offeringInfo{}
		err = offeringRows.Scan(
			&info.ID,
			&info.Title,
			&info.Section,
			&info.Term,

			&info.Time,

			&info.Place,

			&info.FacultyID,
			&info.FacultyName,

			&selectedSections,
		)

		if !strings.Contains(selectedSections, info.Section) {
			continue
		}

		termInfo, err := mit.GetTermByCode(info.Term)
		if err != nil {
			return data.ProviderData{}, err
		}

		info.ParsedTime, err = mit.ParseTimeInfo(info.Time, termInfo)
		if err != nil {
			return data.ProviderData{}, err
		}

		offerings = append(offerings, info)
	}

	for _, offering := range offerings {
		currentDay := startTime
		for i := 0; i < dayCount; i++ {
			if i != 0 {
				currentDay = currentDay.Add(24 * time.Hour)
			}

			if currentDay.Before(offering.ParsedTime.BeginsOn) {
				continue
			}

			if currentDay.After(offering.ParsedTime.EndsOn.AddDate(0, 0, 1)) {
				continue
			}

			termInfo, err := mit.GetTermByCode(offering.Term)
			if err != nil {
				return data.ProviderData{}, err
			}

			currentEffectiveWeekday := currentDay.Weekday()
			newWeekday, isException := termInfo.ExceptionDays[currentDay.Format("2006-01-02")]
			if isException {
				currentEffectiveWeekday = newWeekday
			}

			foundWeekday := false
			relevantInfo := mit.ScheduledMeeting{}
			for _, info := range offering.ParsedTime.Meetings {
				for _, weekday := range info.Weekdays {
					if currentEffectiveWeekday == weekday {
						foundWeekday = true
						relevantInfo = info
						break
					}
				}
			}
			if !foundWeekday {
				continue
			}

			dayString := currentDay.Format("2006-01-02")
			dayTime, _ := time.ParseInLocation("2006-01-02", dayString, location)
			dayOffset := int(dayTime.Unix())

			// check if it's an off day
			if util.StringSliceContains(offDays, dayString) {
				continue
			}

			event := data.Event{
				Tags: map[data.EventTagType]interface{}{},
			}

			event.ID = -1
			event.Name = offering.ID + " - " + offering.Title + " - " + offering.Section

			typeDisplay, _ := sectionCharToDisplayName[offering.Section[0]]

			event.Tags[data.EventTagShortName] = offering.ID + " " + typeDisplay
			event.Tags[data.EventTagReadOnly] = true
			event.Tags[data.EventTagLocation] = offering.Place

			event.Start = relevantInfo.StartSeconds
			event.End = relevantInfo.EndSeconds

			event.Start += dayOffset
			event.End += dayOffset

			result.Events = append(result.Events, event)
		}
	}

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

			// TODO: make this not depend on the global current term
			termInfo := mit.GetCurrentTerm()

			currentEffectiveWeekday := currentDay.Weekday()
			newWeekday, isException := termInfo.ExceptionDays[currentDay.Format("2006-01-02")]
			if isException {
				currentEffectiveWeekday = newWeekday
			}

			foundWeekday := false
			for _, weekday := range peInfo.ParsedDaysOfWeek {
				if currentEffectiveWeekday == weekday {
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

			// check if it's an off day
			if util.StringSliceContains(offDays, dayString) {
				continue
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
