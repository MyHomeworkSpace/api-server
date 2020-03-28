package schools

import "github.com/MyHomeworkSpace/api-server/data"

// A Provider is a generic type that implements all methods of a calendar.Provider except for GetView()
type Provider struct {
	School data.School
}

func (p *Provider) ID() string {
	return p.School.ID()
}

func (p *Provider) Name() string {
	return p.School.Name() + " Schedule"
}
