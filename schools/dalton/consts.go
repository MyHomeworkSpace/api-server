package dalton

import "time"

// An AnnouncementGrade describes the target audience for a planner announcement.
type AnnouncementGrade int

const (
	AnnouncementGradeAll             AnnouncementGrade = 0  // everyone
	AnnouncementGradeMiddleSchool    AnnouncementGrade = 14 // 4th grade through 8th grade
	AnnouncementGradeHighSchool      AnnouncementGrade = 15 // 9th grade through 12th grade
	AnnouncementGradeMiddleSchool456 AnnouncementGrade = 16 // 4th, 5th, and 6th grade
	AnnouncementGradeMiddleSchool78  AnnouncementGrade = 17 // 7th, and 8th grade
	AnnouncementGradeFaculty         AnnouncementGrade = 18 // faculty member
)

// An AssemblyType describes what happens for assembly on a given week.
type AssemblyType int

const (
	AssemblyTypeAssembly AssemblyType = iota
	AssemblyTypeLongHouse
	AssemblyTypeLab
)

// A SchoolMode describes the overall mode that scheduling takes place in.
type SchoolMode int

// The available SchoolModes.
const (
	// SchoolModeNormal refers to the normal Dalton schedule: two semesters with 45-minute classes.
	SchoolModeNormal SchoolMode = iota

	// SchoolModeVirtual refers to the Dalton schedule as modified to respond to the COVID-19 pandemic: three semesters with classes varying in duration.
	SchoolModeVirtual
)

type importTerm struct {
	Start      time.Time
	End        time.Time
	DayOffsets []int
}

func mustParse(t time.Time, err error) time.Time {
	if err != nil {
		panic(err)
	}

	return t
}

// these change every year
var (
	// the grade that someone in the class of 2019 is in for this year
	// used to calculate other people's grade
	Grade_ClassOf2019 AnnouncementGrade = 13

	// the current school mode
	CurrentMode = SchoolModeVirtual

	// only relevant in SchoolModeNormal
	Day_SeniorEnd, _ = time.Parse("2006-01-02", "2020-04-23")

	TermMap = map[string][]time.Time{
		"1st Term": {
			mustParse(time.Parse("2006-01-02", "2020-09-21")),
			mustParse(time.Parse("2006-01-02", "2020-12-19")),
		},
		"2nd Term": {
			mustParse(time.Parse("2006-01-02", "2021-01-04")),
			mustParse(time.Parse("2006-01-02", "2021-03-20")),
		},
		"3rd Term": {
			mustParse(time.Parse("2006-01-02", "2021-04-05")),
			mustParse(time.Parse("2006-01-02", "2021-06-17")),
		},
	}

	// days that are overridden with another weekday's schedule
	ExceptionDays = map[string]time.Weekday{
		"2020-12-16": time.Thursday,
		"2020-12-17": time.Friday,
		"2021-01-20": time.Monday,
		"2021-02-10": time.Friday,
	}

	// import ranges
	// these should be ranges with 4 fridays/2 wednesdays in a row and the first week having no off days
	ImportTerms = []importTerm{
		{
			Start: time.Date(2020, time.September, 21, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2021, time.January, 3, 0, 0, 0, 0, time.UTC),
			DayOffsets: []int{
				2,
				((7 * 1) + 2),
			},
		},
		{
			Start: time.Date(2021, time.January, 4, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2021, time.March, 19, 0, 0, 0, 0, time.UTC),
			DayOffsets: []int{
				2,
				((7 * 1) + 2),
			},
		},
		{
			Start: time.Date(2021, time.April, 5, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2021, time.June, 17, 0, 0, 0, 0, time.UTC),
			DayOffsets: []int{
				2,
				((7 * 1) + 2),
			},
		},
	}

	AssemblyTypeList = map[string]AssemblyType{}
)
