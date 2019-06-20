package dalton

import (
	"fmt"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

type school struct {
	importStatus ImportStatus
	name         string
	username     string
}

func (s *school) ID() string {
	return "dalton"
}

func (s *school) Name() string {
	return "The Dalton School"
}

func (s *school) UserDetails() string {
	return fmt.Sprintf("Signed in as %s (%s)", s.name, s.username)
}

func (s *school) EmailDomain() string {
	return "dalton.org"
}

func (s *school) Hydrate(data map[string]interface{}) error {
	s.importStatus = ImportStatus(data["status"].(float64))
	s.name = data["name"].(string)
	s.username = data["username"].(string)
	return nil
}

func (s *school) CalendarProvider() data.Provider {
	return &provider{
		schools.Provider{
			School: s,
		},
	}
}

func CreateSchool() *school {
	return &school{}
}
