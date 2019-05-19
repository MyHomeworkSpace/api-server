package dalton

import (
	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

type school struct {
}

func (s *school) ID() string {
	return "dalton"
}

func (s *school) Name() string {
	return "The Dalton School"
}

func (s *school) EmailDomain() string {
	return "dalton.org"
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
