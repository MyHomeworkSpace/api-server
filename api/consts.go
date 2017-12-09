package api

import "time"

// these change every year
var (
	Day_SchoolStart, _    = time.Parse("2006-01-02", "2017-09-11")
	Day_Candlelighting, _ = time.Parse("2006-01-02", "2017-12-22")
	Day_ExamRelief, _     = time.Parse("2006-01-02", "2018-01-24")
	Day_SchoolEnd, _      = time.Parse("2006-01-02", "2018-06-07")

	SpecialSchedule_HS_Candlelighting = []CalendarSpecialScheduleItem{
		CalendarSpecialScheduleItem{"C", "", 29400, 31500},
		CalendarSpecialScheduleItem{"D", "", 31800, 33900},
		CalendarSpecialScheduleItem{"H", "", 34200, 36300},
		CalendarSpecialScheduleItem{"G", "", 36600, 38700},
		CalendarSpecialScheduleItem{"", "Long House", 39000, 41100},
		CalendarSpecialScheduleItem{"", "Candlelighting ceremony", 41400, 43200},
	}

	// import ranges
	// these should be ranges with 4 fridays in a row and the first week having no off days
	Term1_Import_Start = time.Date(2017, time.September, 11, 0, 0, 0, 0, time.UTC)
	Term1_Import_End   = time.Date(2017, time.October, 7, 0, 0, 0, 0, time.UTC)

	Term2_Import_Start = time.Date(2018, time.January, 29, 0, 0, 0, 0, time.UTC)
	Term2_Import_End   = time.Date(2018, time.February, 24, 0, 0, 0, 0, time.UTC)

	// the grade that someone in the class of 2019 is in for this year
	// used to calculate other people's grade
	Grade_ClassOf2019 = 11

	// HACK: hard-coded friday list because we can't get the fridays from the schedule because some MS teacher schedules don't have the numbers for some reason
	ScheduleFridayList = map[string]int{
		"2017-09-15": 1,
		"2017-09-22": 2,
		"2017-09-29": 3,
		"2017-10-06": 4,
		"2017-10-13": 1,
		"2017-10-20": 2,
		"2017-10-27": 3,
		"2017-11-03": 4,
		"2017-11-10": 1,
		"2017-12-01": 3,
		"2017-12-08": 4,
		"2017-12-15": 1,
		"2018-01-12": 3,
		"2018-01-26": 1,
		"2018-02-02": 2,
		"2018-02-09": 3,
		"2018-02-16": 4,
		"2018-02-23": 1,
		"2018-03-02": 2,
		"2018-03-09": 3,
		"2018-03-16": 4,
		"2018-04-13": 2,
		"2018-04-20": 3,
		"2018-05-04": 1,
		"2018-05-11": 2,
		"2018-05-18": 3,
		"2018-05-25": 4,
		"2018-01-19": 4,
		"2018-06-01": 1,
		"2018-04-06": 1,
		"2018-04-27": 4,
	}
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
