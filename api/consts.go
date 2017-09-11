package api

import "time"

// these change every year
var (
	Day_SchoolStart, _ = time.Parse("2006-01-02", "2017-09-11")
	Day_ExamRelief, _  = time.Parse("2006-01-02", "2018-01-24")
	Day_SchoolEnd, _   = time.Parse("2006-01-02", "2018-06-07")

	Term1_Import_Start = time.Date(2017, time.September, 11, 0, 0, 0, 0, time.UTC)
	Term1_Import_End   = time.Date(2017, time.October, 7, 0, 0, 0, 0, time.UTC)

	Term2_Import_Start = time.Date(2018, time.January, 29, 0, 0, 0, 0, time.UTC)
	Term2_Import_End   = time.Date(2018, time.February, 24, 0, 0, 0, 0, time.UTC)

	// the grade that someone in the class of 2019 is in for this year
	// used to calculate other people's grade
	Grade_ClassOf2019 = 11
)

// these are constants used to keep track of things
const (
	AnnouncementGrade_All              = 0  // everyone
	AnnouncementGrade_MiddleSchool     = 14 // 4th grade through 8th grade
	AnnouncementGrade_HighSchool       = 15 // 9th grade through 12th grade
	AnnouncementGrade_MiddleSchool_456 = 16 // 4th, 5th, and 6th grade
	AnnouncementGrade_MiddleSchool_78  = 17 // 7th, and 8th grade
	AnnouncementGrade_Faculty          = 18 // 7th, and 8th grade
)

const (
	AnnouncementType_Text       = 0 // just informative
	AnnouncementType_FullOff    = 1 // no classes at all
	AnnouncementType_BreakStart = 2 // start of a break (inclusive of that day!)
	AnnouncementType_BreakEnd   = 3 // end of a break (exclusive of that day!)
)
