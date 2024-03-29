package columbia

import (
	"fmt"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

type school struct {
	barnard bool

	importStatus schools.ImportStatus

	name     string
	username string
}

func (s *school) ID() string {
	if s.barnard {
		return "barnard"
	}

	return "columbia"
}

func (s *school) Name() string {
	if s.barnard {
		return "Barnard College"
	}

	return "Columbia University"
}

func (s *school) ShortName() string {
	if s.barnard {
		return "Barnard"
	}

	return "Columbia"
}

func (s *school) UserDetails() string {
	return fmt.Sprintf("Signed in as %s (%s)", s.name, s.username)
}

func (s *school) EmailAddress() string {
	if s.barnard {
		return s.username + "@barnard.edu"
	}

	return s.username + "@columbia.edu"
}

func (s *school) EmailDomain() string {
	if s.barnard {
		return "barnard.edu"
	}

	return "columbia.edu"
}

func (s *school) Prefixes() []data.Prefix {
	return []data.Prefix{}
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
func CreateSchool(barnard bool) *school {
	return &school{
		barnard: barnard,
	}
}
