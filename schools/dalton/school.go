package dalton

import (
	"fmt"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

type school struct {
	importStatus schools.ImportStatus
	name         string
	username     string
}

func (s *school) ID() string {
	return "dalton"
}

func (s *school) Name() string {
	return "The Dalton School"
}

func (s *school) ShortName() string {
	return "Dalton"
}

func (s *school) UserDetails() string {
	return fmt.Sprintf("Signed in as %s (%s)", s.name, s.username)
}

func (s *school) EmailAddress() string {
	return s.username + "@dalton.org"
}

func (s *school) EmailDomain() string {
	return "dalton.org"
}

func (s *school) Prefixes() []data.Prefix {
	return []data.Prefix{
		data.Prefix{
			ID:         -1,
			Background: "2AF15E",
			Color:      "FFFFFF",
			Words:      []string{"Lab", "BookALab", "BookLab"},
			TimedEvent: true,
			Default:    true,
		},
	}
}

func (s *school) Hydrate(data map[string]interface{}) error {
	s.importStatus = schools.ImportStatus(data["status"].(float64))
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

// CreateSchool returns a new instance of the school.
func CreateSchool() *school {
	return &school{}
}
