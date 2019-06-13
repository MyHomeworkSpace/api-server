package data

// School is an interface implemented by all schools that a user can connect to their account
type School interface {
	ID() string
	Name() string
	EmailDomain() string
	CalendarProvider() Provider
}

// SchoolRegistry is an interface implemented by the central registry in the schools package
type SchoolRegistry interface {
	GetSchoolByID(id string) (School, error)
	Register(school School)
}

// GetSchoolsForUser returns a list of schools that the given user ID is enrolled in
func GetSchoolsForUser(r SchoolRegistry, userID int) ([]School, error) {
	// TODO: not hardcode this for dalton
	dalton, err := r.GetSchoolByID("dalton")
	if err != nil {
		return nil, err
	}

	return []School{
		dalton,
	}, nil
}
