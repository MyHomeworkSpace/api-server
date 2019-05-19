package dalton

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

type provider struct {
	schools.Provider
}

func (p *provider) GetData(db *sql.DB, user *data.User, startTime time.Time, endTime time.Time, dataType data.ProviderDataType) (data.ProviderData, error) {
	result := data.ProviderData{
		Announcements: nil,
		Events:        nil,
	}

	// get all friday information for time period
	fridayRows, err := db.Query("SELECT * FROM fridays WHERE date >= ? AND date <= ?", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
	if err != nil {
		return data.ProviderData{}, err
	}
	fridays := []data.PlannerFriday{}
	for fridayRows.Next() {
		friday := data.PlannerFriday{}
		fridayRows.Scan(&friday.ID, &friday.Date, &friday.Index)
		fridays = append(fridays, friday)
	}
	fridayRows.Close()

	if dataType&data.ProviderDataAnnouncements != 0 {
		result.Announcements = []data.PlannerAnnouncement{}

		// add fridays as announcements
		for _, friday := range fridays {
			fridayAnnouncement := data.PlannerAnnouncement{
				ID:    -1,
				Date:  friday.Date,
				Text:  "Friday " + strconv.Itoa(friday.Index),
				Grade: -1,
				Type:  0,
			}
			result.Announcements = append(result.Announcements, fridayAnnouncement)
		}
	}

	if dataType&data.ProviderDataEvents != 0 {
		result.Events = []data.Event{}
	}

	return result, nil
}
