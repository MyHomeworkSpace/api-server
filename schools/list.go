package schools

import (
	"github.com/MyHomeworkSpace/api-server/data"
)

// Registry is a struct that tracks available schools
type Registry struct {
	schoolList []data.School
}

// MainRegistry is the main registry of schools
var MainRegistry = &Registry{
	[]data.School{},
}

// GetSchoolByEmailDomain returns the school associated with the given email domain, or nil if it doesn't exist.
func (r *Registry) GetSchoolByEmailDomain(domain string) (data.School, error) {
	for _, school := range r.schoolList {
		if school.EmailDomain() == domain {
			return school, nil
		}
	}
	return nil, data.ErrNotFound
}

// GetSchoolByID returns the school associated with the given ID, or nil if it doesn't exist.
func (r *Registry) GetSchoolByID(id string) (data.School, error) {
	for _, school := range r.schoolList {
		if school.ID() == id {
			return school, nil
		}
	}
	return nil, data.ErrNotFound
}

// Register registers the given school
func (r *Registry) Register(school data.School) {
	r.schoolList = append(r.schoolList, school)
}
