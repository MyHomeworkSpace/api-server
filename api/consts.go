package api

// this changes every year
const (
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
	AnnouncementGrade_Faculty          = 18 // faculty member
)

const (
	AnnouncementType_Text       = 0 // just informative
	AnnouncementType_FullOff    = 1 // no classes at all
	AnnouncementType_BreakStart = 2 // start of a break (inclusive of that day!)
	AnnouncementType_BreakEnd   = 3 // end of a break (exclusive of that day!)
)

const (
	BlackbaudLevel_MiddleSchool = 167
	BlackbaudLevel_HighSchool   = 166
)

const (
	BlackbaudPersona_Student = 2
	BlackbaudPersona_Faculty = 3
)
