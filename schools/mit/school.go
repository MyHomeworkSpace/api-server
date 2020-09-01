package mit

import (
	"encoding/json"
	"fmt"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

type school struct {
	importStatus schools.ImportStatus

	name     string
	username string

	peInfo *peInfo
	showPE bool
}

func (s *school) ID() string {
	return "mit"
}

func (s *school) Name() string {
	return "Massachusetts Institute of Technology"
}

func (s *school) ShortName() string {
	return "MIT"
}

func (s *school) UserDetails() string {
	return fmt.Sprintf("Signed in as %s (%s)", s.name, s.username)
}

func (s *school) EmailAddress() string {
	return s.username + "@mit.edu"
}

func (s *school) EmailDomain() string {
	return "mit.edu"
}

func (s *school) Prefixes() []data.Prefix {
	return []data.Prefix{
		{
			ID:         -1,
			Background: "4C6C9B",
			Color:      "FFFFFF",
			Words:      []string{"ProblemSet", "Pset"},
			Default:    true,
		},
	}
}

func (s *school) Hydrate(data map[string]interface{}) error {
	s.importStatus = schools.ImportStatus(data["status"].(float64))

	s.name = data["name"].(string)
	s.username = data["username"].(string)

	peInfoString, havePEInfo := data["peInfo"].(string)
	if havePEInfo {
		s.peInfo = &peInfo{}
		err := json.Unmarshal([]byte(peInfoString), s.peInfo)
		if err != nil {
			return err
		}
	} else {
		s.peInfo = nil
	}

	s.showPE = true
	showPEInterface, ok := data["showPE"]
	if ok {
		showPEBool, ok := showPEInterface.(bool)
		if ok {
			s.showPE = showPEBool
		}
	}

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
