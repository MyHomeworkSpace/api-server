package dalton

import "time"

const (
	AnnouncementGrade_All              = 0  // everyone
	AnnouncementGrade_MiddleSchool     = 14 // 4th grade through 8th grade
	AnnouncementGrade_HighSchool       = 15 // 9th grade through 12th grade
	AnnouncementGrade_MiddleSchool_456 = 16 // 4th, 5th, and 6th grade
	AnnouncementGrade_MiddleSchool_78  = 17 // 7th, and 8th grade
	AnnouncementGrade_Faculty          = 18 // faculty member
)

const (
	BlackbaudLevel_MiddleSchool = 167
	BlackbaudLevel_HighSchool   = 166
)

const (
	BlackbaudPersona_Student = 2
	BlackbaudPersona_Faculty = 3
)

// An AssemblyType describes what happens for assembly on a given week.
type AssemblyType int

const (
	AssemblyType_Assembly AssemblyType = iota
	AssemblyType_LongHouse
	AssemblyType_Lab
)

// these change every year
var (
	// the grade that someone in the class of 2019 is in for this year
	// used to calculate other people's grade
	Grade_ClassOf2019 = 13

	Day_SchoolStart, _   = time.Parse("2006-01-02", "2019-09-09")
	Day_ExamRelief, _    = time.Parse("2006-01-02", "2020-01-24")
	Day_SeniorLastDay, _ = time.Parse("2006-01-02", "2020-04-23")
	Day_SchoolEnd, _     = time.Parse("2006-01-02", "2020-06-11")

	AssemblyTypeList = map[string]AssemblyType{
		"2019-09-12": AssemblyType_Assembly,
		"2019-09-19": AssemblyType_Assembly,
		"2019-09-26": AssemblyType_Lab,
		"2019-10-03": AssemblyType_Assembly,
		"2019-10-10": AssemblyType_Lab,
		"2019-10-17": AssemblyType_Assembly,
		"2019-10-24": AssemblyType_LongHouse,
		"2019-10-31": AssemblyType_Lab,
		"2019-11-07": AssemblyType_Lab,
		"2019-11-14": AssemblyType_Assembly,
		"2019-11-21": AssemblyType_Assembly,
		"2019-12-05": AssemblyType_Assembly,
		"2019-12-12": AssemblyType_LongHouse,
		"2019-12-19": AssemblyType_Assembly,
		"2020-01-09": AssemblyType_Assembly,
		"2020-01-30": AssemblyType_Assembly,
		"2020-02-06": AssemblyType_Lab,
		"2020-02-13": AssemblyType_Assembly,
		"2020-02-20": AssemblyType_Lab,
		"2020-02-27": AssemblyType_Lab,
		"2020-03-05": AssemblyType_Lab,
		"2020-03-12": AssemblyType_Assembly,
		"2020-04-02": AssemblyType_LongHouse,
		"2020-04-16": AssemblyType_Assembly,
		"2020-04-23": AssemblyType_Lab,
		"2020-04-30": AssemblyType_Assembly,
		"2020-05-07": AssemblyType_Assembly,
		"2020-05-14": AssemblyType_Assembly,
		"2020-05-21": AssemblyType_Assembly,
		"2020-05-28": AssemblyType_Lab,
	}
)
