package calendar

import (
	"time"
)

// these change every year
var (
	Day_SchoolStart, _    = time.Parse("2006-01-02", "2018-09-12")
	Day_Candlelighting, _ = time.Parse("2006-01-02", "2018-12-21")
	Day_ExamRelief, _     = time.Parse("2006-01-02", "2019-01-25")
	Day_SchoolEnd, _      = time.Parse("2006-01-02", "2019-06-06")

	// SpecialSchedule_HS_Candlelighting = []CalendarSpecialScheduleItem{
	// 	CalendarSpecialScheduleItem{"C", "", 29400, 31500},
	// 	CalendarSpecialScheduleItem{"D", "", 31800, 33900},
	// 	CalendarSpecialScheduleItem{"H", "", 34200, 36300},
	// 	CalendarSpecialScheduleItem{"G", "", 36600, 38700},
	// 	CalendarSpecialScheduleItem{"", "Long House", 39000, 41100},
	// 	CalendarSpecialScheduleItem{"", "Candlelighting ceremony", 41400, 43200},
	// }

	// import ranges
	// these should be ranges with 4 fridays in a row and the first week having no off days
	Term1_Import_Start = time.Date(2018, time.September, 24, 0, 0, 0, 0, time.UTC)
	Term1_Import_End   = time.Date(2018, time.October, 20, 0, 0, 0, 0, time.UTC)

	Term1_Import_DayOffset_Friday1 = ((7 * 3) + 4)
	Term1_Import_DayOffset_Friday2 = 4
	Term1_Import_DayOffset_Friday3 = ((7 * 1) + 4)
	Term1_Import_DayOffset_Friday4 = ((7 * 2) + 4)

	Term2_Import_Start = time.Date(2019, time.January, 28, 0, 0, 0, 0, time.UTC)
	Term2_Import_End   = time.Date(2019, time.February, 23, 0, 0, 0, 0, time.UTC)

	Term2_Import_DayOffset_Friday1 = ((7 * 3) + 4)
	Term2_Import_DayOffset_Friday2 = 4
	Term2_Import_DayOffset_Friday3 = ((7 * 1) + 4)
	Term2_Import_DayOffset_Friday4 = ((7 * 2) + 4)

	// HACK: hard-coded friday list because we can't get the fridays from the schedule because some MS teacher schedules don't have the numbers for some reason
	ScheduleFridayList = map[string]int{
		"2018-09-14": 1,
		"2018-09-21": 2,
		"2018-09-28": 3,
		"2018-10-05": 4,
		"2018-10-12": 1,
		"2018-10-19": 2,
		"2018-10-26": 3,
		"2018-11-02": 4,
		"2018-11-09": 1,
		"2018-11-16": 2,
		"2018-11-30": 3,
		"2018-12-07": 4,
		"2018-12-14": 1,
		"2019-01-11": 3,
		"2019-01-18": 4,
		"2019-01-25": 1,
		"2019-02-01": 2,
		"2019-02-08": 3,
		"2019-02-15": 4,
		"2019-02-22": 1,
		"2019-03-01": 2,
		"2019-03-08": 3,
		"2019-03-15": 4,
		"2019-04-05": 1,
		"2019-04-12": 2,
		"2019-04-26": 3,
		"2019-05-03": 4,
		"2019-05-10": 1,
		"2019-05-17": 2,
		"2019-05-24": 3,
		"2019-05-31": 4,
	}

	SpecialAssessmentList = map[int]*SpecialAssessmentInfo{}
	SpecialAssessmentDays = map[string]SpecialAssessmentType{}
)

const (
	AnnouncementType_Text       = 0 // just informative
	AnnouncementType_FullOff    = 1 // no classes at all
	AnnouncementType_BreakStart = 2 // start of a break (inclusive of that day!)
	AnnouncementType_BreakEnd   = 3 // end of a break (exclusive of that day!)
)

const (
	SpecialAssessmentType_Unknown  SpecialAssessmentType = 0
	SpecialAssessmentType_English                        = 1
	SpecialAssessmentType_History                        = 2
	SpecialAssessmentType_Math                           = 3
	SpecialAssessmentType_Science                        = 4
	SpecialAssessmentType_Language                       = 5
)

func InitCalendar() {
	// special assessments
	// no special assessments at this time
}
