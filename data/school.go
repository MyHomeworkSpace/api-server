package data

import "database/sql"

var MainRegistry SchoolRegistry

// School is an interface implemented by all schools that a user can connect to their account
type School interface {
	ID() string
	Name() string
	ShortName() string
	UserDetails() string
	EmailAddress() string
	EmailDomain() string
	Prefixes() []Prefix

	CalendarProvider() Provider

	Hydrate(data map[string]interface{}) error

	Enroll(tx *sql.Tx, user *User, params map[string]interface{}) (map[string]interface{}, error)
	Unenroll(tx *sql.Tx, user *User) error
	NeedsUpdate(db *sql.DB) (bool, error)
}

// SchoolError is a type of Error that occurs due to invalid data or a similar condition. It's used to distinguish from internal server errors.
type SchoolError struct {
	Code string
}

func (e SchoolError) Error() string {
	return "school: " + e.Code
}

// DetailedSchoolError is a type of Error that occurs when a school provider wants to communicate data back to the client along with an error. This is useful for multi-step enrollments, where more information is required from the user.
type DetailedSchoolError struct {
	Code    string
	Details map[string]interface{}
}

func (e DetailedSchoolError) Error() string {
	return "school: " + e.Code
}

// SchoolInfo is a struct that holds information about a school. It's used to hold data in a format that the JSON package can then marshal out to the client.
type SchoolInfo struct {
	EnrollmentID int    `json:"enrollmentID"`
	SchoolID     string `json:"schoolID"`
	DisplayName  string `json:"displayName"`
	ShortName    string `json:"shortName"`
	UserDetails  string `json:"userDetails"`
	EmailAddress string `json:"emailAddress"`
	School       School `json:"-"`
	UserID       int    `json:"userID"`
}

// SchoolResult is a struct that holds information about a school that was searched for. (e.g. by email domain) It's used to hold data in a format that the JSON package can then marshal out to the client.
type SchoolResult struct {
	SchoolID    string `json:"schoolID"`
	DisplayName string `json:"displayName"`
	ShortName   string `json:"shortName"`
}

// SchoolRegistry is an interface implemented by the central registry in the schools package
type SchoolRegistry interface {
	GetAllSchools() []School
	GetSchoolByEmailDomain(domain string) (School, error)
	GetSchoolByID(id string) (School, error)
	Register(school School)
}

// GetSchoolsForUser returns a list of schools that the given user is enrolled in
func GetSchoolsForUser(user *User) ([]School, error) {
	schools := []School{}

	for _, info := range user.Schools {
		schools = append(schools, info.School)
	}

	return schools, nil
}
