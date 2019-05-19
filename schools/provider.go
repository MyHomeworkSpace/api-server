package schools

// A Provider is a generic type that implements all methods of a calendar.Provider except for GetView()
type Provider struct {
	School School
}

func (p *Provider) Name() string {
	return p.School.Name() + " Calendar"
}
