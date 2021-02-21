package cornell

import (
	"fmt"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

type school struct {
	importStatus schools.ImportStatus

	name  string
	netID string
}

func (s *school) ID() string {
	return "cornell"
}

func (s *school) Name() string {
	return "Cornell University"
}

func (s *school) ShortName() string {
	return "Cornell"
}

func (s *school) UserDetails() string {
	return fmt.Sprintf("Signed in as %s (%s)", s.name, s.netID)
}

func (s *school) EmailAddress() string {
	return s.netID + "@cornell.edu"
}

func (s *school) EmailDomain() string {
	return "cornell.edu"
}

func (s *school) Prefixes() []data.Prefix {
	return []data.Prefix{
		{
			ID:         -1,
			Background: "DC143C",
			Color:      "FFFFFF",
			Words:      []string{"Prelim", "Preliminary", "SemiFinal"},
			Default:    true,
		},
		{
			ID:         -1,
			Background: "2AC0F1",
			Color:      "FFFFFF",
			Words:      []string{"Workshop"},
			Default:    true,
		},
		{
			ID:         -1,
			Background: "C3A528",
			Color:      "FFFFFF",
			Words:      []string{"PreLab", "PostLab"},
			Default:    true,
		},
	}
}

func (s *school) Hydrate(data map[string]interface{}) error {
	s.name = data["name"].(string)
	s.netID = data["netid"].(string)
	s.importStatus = schools.ImportStatus(data["status"].(float64))
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
