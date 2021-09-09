package columbia

import (
	"database/sql"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

type provider struct {
	schools.Provider
}

type columbiaMeeting struct {
	ID         int
	Department string
	Number     string
	Section    string
	Name       string
	Building   string
	Room       string
	DayOfWeek  time.Weekday
	Start      int
	End        int
	BeginDate  time.Time
	EndDate    time.Time
	UserID     int
}

func (p *provider) GetData(db *sql.DB, user *data.User, location *time.Location, startTime time.Time, endTime time.Time, dataType data.ProviderDataType) (data.ProviderData, error) {
	result := data.ProviderData{
		Announcements: []data.PlannerAnnouncement{},
		Events:        []data.Event{},
	}

	dayCount := int((endTime.Sub(startTime).Hours() / 24) + 0.5)

	// check for any holidays during the time period
	announcementRows, err := db.Query("SELECT id, date, text FROM columbia_holidays WHERE date >= ? AND date <= ?", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
	if err != nil {
		return data.ProviderData{}, err
	}
	defer announcementRows.Close()
	for announcementRows.Next() {
		resp := data.PlannerAnnouncement{
			Type: data.AnnouncementTypeFullOff,
		}

		err := announcementRows.Scan(&resp.ID, &resp.Date, &resp.Text)
		if err != nil {
			return data.ProviderData{}, err
		}

		result.Announcements = append(result.Announcements, resp)
	}

	// generate list of all off days in time period
	offDays := map[string]bool{}
	for _, announcement := range result.Announcements {
		if announcement.Type == data.AnnouncementTypeFullOff {
			offDays[announcement.Date] = true
		}
	}

	if dataType&data.ProviderDataEvents != 0 {
		meetingRows, err := db.Query("SELECT id, department, number, section, name, building, room, dow, start, end, beginDate, endDate, userID FROM columbia_meetings WHERE userID = ?", user.ID)
		if err != nil {
			return data.ProviderData{}, err
		}
		defer meetingRows.Close()

		meetingMap := map[time.Weekday][]columbiaMeeting{}
		for i := time.Sunday; i <= time.Saturday; i++ {
			meetingMap[i] = []columbiaMeeting{}
		}

		for meetingRows.Next() {
			meeting := columbiaMeeting{}
			beginDateString := ""
			endDateString := ""

			err = meetingRows.Scan(
				&meeting.ID,
				&meeting.Department,
				&meeting.Number,
				&meeting.Section,
				&meeting.Name,
				&meeting.Building,
				&meeting.Room,
				&meeting.DayOfWeek,
				&meeting.Start,
				&meeting.End,
				&beginDateString,
				&endDateString,
				&meeting.UserID,
			)
			if err != nil {
				return data.ProviderData{}, err
			}

			meeting.BeginDate, err = time.Parse("2006-01-02", beginDateString)
			if err != nil {
				return data.ProviderData{}, err
			}

			meeting.EndDate, err = time.Parse("2006-01-02", endDateString)
			if err != nil {
				return data.ProviderData{}, err
			}

			meetingMap[meeting.DayOfWeek] = append(meetingMap[meeting.DayOfWeek], meeting)
		}

		currentDay, err := time.Parse("2006-01-02", startTime.Format("2006-01-02"))
		if err != nil {
			return data.ProviderData{}, err
		}

		for i := 0; i < dayCount; i++ {
			if i != 0 {
				currentDay = currentDay.Add(24 * time.Hour)
			}

			dayString := currentDay.Format("2006-01-02")

			_, isOffDay := offDays[dayString]
			if isOffDay {
				continue
			}

			meetingsForDay := meetingMap[currentDay.Weekday()]
			dayTime, err := time.ParseInLocation("2006-01-02", dayString, location)
			if err != nil {
				return data.ProviderData{}, err
			}
			dayOffset := int(dayTime.Unix())

			for _, meetingForDay := range meetingsForDay {
				if currentDay.Before(meetingForDay.BeginDate) {
					continue
				}

				if currentDay.After(meetingForDay.EndDate) {
					continue
				}

				event := data.Event{
					Tags: map[data.EventTagType]interface{}{},
				}

				event.ID = -1
				event.UniqueID = meetingForDay.Department + "-" + meetingForDay.Number + "-" + meetingForDay.Section + "-" + dayString
				event.SeriesID = meetingForDay.Department + "-" + meetingForDay.Number
				event.SeriesName = meetingForDay.Department + " " + meetingForDay.Number
				event.Name = meetingForDay.Department + " " + meetingForDay.Number + ": " + meetingForDay.Name

				event.Tags[data.EventTagSection] = meetingForDay.Section
				event.Tags[data.EventTagReadOnly] = true
				event.Tags[data.EventTagCancelable] = true
				event.Tags[data.EventTagBuildingName] = meetingForDay.Building
				event.Tags[data.EventTagRoomNumber] = meetingForDay.Room

				event.Start = meetingForDay.Start
				event.End = meetingForDay.End

				event.Start += dayOffset
				event.End += dayOffset

				result.Events = append(result.Events, event)
			}
		}
	}

	return result, nil
}
