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

// these change every year
var (
	// the grade that someone in the class of 2019 is in for this year
	// used to calculate other people's grade
	Grade_ClassOf2019 AnnouncementGrade = 13

	Day_SchoolStart, _ = time.Parse("2006-01-02", "2019-09-09")
	Day_ExamRelief, _  = time.Parse("2006-01-02", "2020-01-24")
	Day_SeniorEnd, _   = time.Parse("2006-01-02", "2020-04-23")
	Day_SchoolEnd, _   = time.Parse("2006-01-02", "2020-06-11")

	AssemblyTypeList = map[string]AssemblyType{
		"2019-09-12": AssemblyTypeAssembly,
		"2019-09-19": AssemblyTypeAssembly,
		"2019-09-26": AssemblyTypeLab,
		"2019-10-03": AssemblyTypeAssembly,
		"2019-10-10": AssemblyTypeLab,
		"2019-10-17": AssemblyTypeAssembly,
		"2019-10-24": AssemblyTypeLongHouse,
		"2019-10-31": AssemblyTypeLab,
		"2019-11-07": AssemblyTypeLab,
		"2019-11-14": AssemblyTypeAssembly,
		"2019-11-21": AssemblyTypeAssembly,
		"2019-12-05": AssemblyTypeAssembly,
		"2019-12-12": AssemblyTypeLongHouse,
		"2019-12-19": AssemblyTypeAssembly,
		"2020-01-09": AssemblyTypeAssembly,
		"2020-01-30": AssemblyTypeAssembly,
		"2020-02-06": AssemblyTypeLab,
		"2020-02-13": AssemblyTypeAssembly,
		"2020-02-20": AssemblyTypeLab,
		"2020-02-27": AssemblyTypeLab,
		"2020-03-05": AssemblyTypeLab,
		"2020-03-12": AssemblyTypeAssembly,
		"2020-04-02": AssemblyTypeLongHouse,
		"2020-04-16": AssemblyTypeAssembly,
		"2020-04-23": AssemblyTypeLab,
		"2020-04-30": AssemblyTypeAssembly,
		"2020-05-07": AssemblyTypeAssembly,
		"2020-05-14": AssemblyTypeAssembly,
		"2020-05-21": AssemblyTypeAssembly,
		"2020-05-28": AssemblyTypeLab,
	}
)
