package dalton

import (
	"database/sql"
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

	if dataType&data.ProviderDataAnnouncements != 0 {
		result.Announcements = []data.PlannerAnnouncement{}
	}

	if dataType&data.ProviderDataEvents != 0 {
		result.Events = []data.Event{}
	}

	return result, nil
}
