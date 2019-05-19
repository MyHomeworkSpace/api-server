package schools

import "github.com/MyHomeworkSpace/api-server/data"

type School interface {
	ID() string
	Name() string
	EmailDomain() string
	CalendarProvider() data.Provider
}
