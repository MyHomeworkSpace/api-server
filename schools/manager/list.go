package manager

import (
	"github.com/MyHomeworkSpace/api-server/schools"
	"github.com/MyHomeworkSpace/api-server/schools/dalton"
)

var schoolList = []schools.School{
	dalton.CreateSchool(),
}

// GetSchoolByID returns the school associated with the given ID, or nil if it doesn't exist.
func GetSchoolByID(id string) schools.School {
	for _, school := range schoolList {
		if school.ID() == id {
			return school
		}
	}
	return nil
}
