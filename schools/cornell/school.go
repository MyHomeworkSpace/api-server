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
	return "cu"
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
	}
}
