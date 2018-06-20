package calendar

import (
	"time"
)

// these change every year
var (
	Day_SchoolStart, _    = time.Parse("2006-01-02", "2017-09-11")
	Day_Candlelighting, _ = time.Parse("2006-01-02", "2017-12-22")
	Day_ExamRelief, _     = time.Parse("2006-01-02", "2018-01-24")
	Day_SchoolEnd, _      = time.Parse("2006-01-02", "2018-06-07")

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
	Term1_Import_Start = time.Date(2017, time.September, 11, 0, 0, 0, 0, time.UTC)
	Term1_Import_End   = time.Date(2017, time.October, 7, 0, 0, 0, 0, time.UTC)

	Term2_Import_Start = time.Date(2018, time.January, 29, 0, 0, 0, 0, time.UTC)
	Term2_Import_End   = time.Date(2018, time.February, 24, 0, 0, 0, 0, time.UTC)

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
	SpecialAssessmentType_English  = 1
	SpecialAssessmentType_History  = 2
	SpecialAssessmentType_Math     = 3
	SpecialAssessmentType_Science  = 4
	SpecialAssessmentType_Language = 5
)

func InitCalendar() {
	// special assessments
	// final schedule for 2017-2018 school year

	SpecialAssessmentDays["2018-05-30"] = SpecialAssessmentType_English
	SpecialAssessmentDays["2018-05-31"] = SpecialAssessmentType_History
	SpecialAssessmentDays["2018-06-01"] = SpecialAssessmentType_Math
	SpecialAssessmentDays["2018-06-04"] = SpecialAssessmentType_Science
	SpecialAssessmentDays["2018-06-05"] = SpecialAssessmentType_Language

	// ** english **

	// introduction to drama
	SpecialAssessmentList[30434369] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Introduction to Drama", "Rothwell", "601"}
	SpecialAssessmentList[30434366] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Introduction to Drama", "Stifler", "501"}
	SpecialAssessmentList[3464253] = SpecialAssessmentList[30434366]
	SpecialAssessmentList[30434367] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Introduction to Drama", "Glassman", "Cafeteria"}
	SpecialAssessmentList[30434372] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Introduction to Drama", "Bender", "608"}
	SpecialAssessmentList[30434364] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Introduction to Drama", "Leja", "Cafeteria"}
	SpecialAssessmentList[30434365] = SpecialAssessmentList[30434364]
	SpecialAssessmentList[30434370] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Introduction to Drama", "Kerman", "Library"}
	SpecialAssessmentList[30434371] = SpecialAssessmentList[30434370]
	SpecialAssessmentList[30434368] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Introduction to Drama", "Gault", "Library"}

	// american literature
	SpecialAssessmentList[30434880] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "American Literature", "Glassman", "Cafeteria"}
	SpecialAssessmentList[30434882] = SpecialAssessmentList[30434880]
	SpecialAssessmentList[30434872] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "American Literature", "Fisher", "503/505"}
	SpecialAssessmentList[30434874] = SpecialAssessmentList[30434872]
	SpecialAssessmentList[30434875] = SpecialAssessmentList[30434872]
	SpecialAssessmentList[30434876] = SpecialAssessmentList[30434872]
	SpecialAssessmentList[30434894] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "American Literature", "Luten", "610/612"}
	SpecialAssessmentList[30434896] = SpecialAssessmentList[30434894]
	SpecialAssessmentList[30434897] = SpecialAssessmentList[30434894]
	SpecialAssessmentList[30434898] = SpecialAssessmentList[30434894]
	SpecialAssessmentList[30434866] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "American Literature", "Bender", "605"}
	SpecialAssessmentList[30434868] = SpecialAssessmentList[30434866]
	SpecialAssessmentList[30434888] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "American Literature", "Kerman", "Library"}
	SpecialAssessmentList[30434890] = SpecialAssessmentList[30434888]
	SpecialAssessmentList[108467109] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "American Literature", "Forbes", "604"}
	SpecialAssessmentList[108467110] = SpecialAssessmentList[108467109]

	// literature and composition
	SpecialAssessmentList[30434853] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Literature and Composition", "Luten", "612"}
	SpecialAssessmentList[30434854] = SpecialAssessmentList[30434853]
	SpecialAssessmentList[30434857] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Literature and Composition", "Hood", "503"}
	SpecialAssessmentList[30434858] = SpecialAssessmentList[30434857]
	SpecialAssessmentList[30434859] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Literature and Composition", "Leja", "601"}
	SpecialAssessmentList[30434860] = SpecialAssessmentList[30434859]
	SpecialAssessmentList[30435279] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Literature and Composition", "Bender", "605"}
	SpecialAssessmentList[30435280] = SpecialAssessmentList[30435279]
	SpecialAssessmentList[30434861] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Literature and Composition", "Kerman", "610"}
	SpecialAssessmentList[30434862] = SpecialAssessmentList[30434861]
	SpecialAssessmentList[30435277] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Literature and Composition", "Rothwell", "Library"}
	SpecialAssessmentList[30435278] = SpecialAssessmentList[30435277]
	SpecialAssessmentList[30434850] = &SpecialAssessmentInfo{SpecialAssessmentType_English, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Literature and Composition", "Forbes", "Library"}
	SpecialAssessmentList[30434852] = SpecialAssessmentList[30434850]
	SpecialAssessmentList[30434855] = SpecialAssessmentList[30434850]
	SpecialAssessmentList[30434856] = SpecialAssessmentList[30434850]

	// ** history **

	// world history I
	SpecialAssessmentList[30435495] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History I", "Graham (C)", "501"}
	SpecialAssessmentList[30435496] = SpecialAssessmentList[30435495]
	SpecialAssessmentList[30435497] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History I", "Graham (D)", "503"}
	SpecialAssessmentList[30435498] = SpecialAssessmentList[30435497]
	SpecialAssessmentList[30435493] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History I", "Okpalugo", "Cafeteria"}
	SpecialAssessmentList[30435494] = SpecialAssessmentList[30435493]
	SpecialAssessmentList[30435503] = SpecialAssessmentList[30435493]
	SpecialAssessmentList[30435504] = SpecialAssessmentList[30435493]
	SpecialAssessmentList[30435501] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History I", "Matz", "507"}
	SpecialAssessmentList[30435502] = SpecialAssessmentList[30435501]
	SpecialAssessmentList[30435505] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History I", "Fox", "601"}
	SpecialAssessmentList[30435506] = SpecialAssessmentList[30435505]
	SpecialAssessmentList[30435490] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History I", "Kideckel", "605"}
	SpecialAssessmentList[30435492] = SpecialAssessmentList[30435490]
	SpecialAssessmentList[30435499] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History I", "Slick", "Library"}
	SpecialAssessmentList[30435500] = SpecialAssessmentList[30435499]

	// world history III
	SpecialAssessmentList[30470291] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History III", "Kalbag", "Library"}
	SpecialAssessmentList[30470292] = SpecialAssessmentList[30470291]
	SpecialAssessmentList[30470289] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History III", "Kohn (B)", "608"}
	SpecialAssessmentList[30470290] = SpecialAssessmentList[30470289]
	SpecialAssessmentList[30435537] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History III", "Kohn (H)", "610"}
	SpecialAssessmentList[30435538] = SpecialAssessmentList[30435537]
	SpecialAssessmentList[30435530] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History III", "Okpalugo", "Cafeteria"}
	SpecialAssessmentList[30435532] = SpecialAssessmentList[30435530]
	SpecialAssessmentList[30435533] = SpecialAssessmentList[30435530]
	SpecialAssessmentList[30435534] = SpecialAssessmentList[30435530]
	SpecialAssessmentList[30435548] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History III", "Slick", "Library"}
	SpecialAssessmentList[30435550] = SpecialAssessmentList[30435548]
	SpecialAssessmentList[30504477] = SpecialAssessmentList[30435548]
	SpecialAssessmentList[30504479] = SpecialAssessmentList[30435548]
	SpecialAssessmentList[30435535] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "World History III", "Davidson", "612"}
	SpecialAssessmentList[30435536] = SpecialAssessmentList[30435535]

	// world history II
	SpecialAssessmentList[30435519] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "World History II", "Kohn", "Library"}
	SpecialAssessmentList[30435520] = SpecialAssessmentList[30435519]
	SpecialAssessmentList[30435521] = SpecialAssessmentList[30435519]
	SpecialAssessmentList[30435522] = SpecialAssessmentList[30435519]
	SpecialAssessmentList[30435523] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "World History II", "Graham", "Library"}
	SpecialAssessmentList[30435524] = SpecialAssessmentList[30435523]
	SpecialAssessmentList[30435525] = SpecialAssessmentList[30435523]
	SpecialAssessmentList[30435526] = SpecialAssessmentList[30435523]
	SpecialAssessmentList[30435517] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "World History II", "Kalbag", "612"}
	SpecialAssessmentList[30435518] = SpecialAssessmentList[30435517]
	SpecialAssessmentList[30435510] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "World History II", "Davidson (A)", "608"}
	SpecialAssessmentList[30435512] = SpecialAssessmentList[30435510]
	SpecialAssessmentList[30435515] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "World History II", "Davidson (C)", "610"}
	SpecialAssessmentList[30435516] = SpecialAssessmentList[30435515]
	SpecialAssessmentList[30435513] = &SpecialAssessmentInfo{SpecialAssessmentType_History, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "World History II", "Matz", "601"}
	SpecialAssessmentList[30435514] = SpecialAssessmentList[30435513]

	// ** math **

	// first block
	// Adv. Algebra & Trig	Cummings	251
	SpecialAssessmentList[30434726] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "Precalculus", "Metz", "Library"}
	SpecialAssessmentList[30434728] = SpecialAssessmentList[30434726]
	SpecialAssessmentList[30434734] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "PBL Precalc 1", "Metz/Cohen", "Library"}
	SpecialAssessmentList[30434736] = SpecialAssessmentList[30434734]
	SpecialAssessmentList[30434740] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((8 * 60) + 30) * 60, ((11 * 60) + 00) * 60, "PBL Precalc 1A", "Metz/Cohen", "Library"}
	SpecialAssessmentList[30434742] = SpecialAssessmentList[30434740]
	SpecialAssessmentList[30470256] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "PBL Precalc 1", "Cohen/Borenstein", "Library"}
	SpecialAssessmentList[30470257] = SpecialAssessmentList[30470256]
	SpecialAssessmentList[30470258] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((8 * 60) + 30) * 60, ((11 * 60) + 00) * 60, "PBL Precalc 1A", "Cohen/Borenstein", "Library"}
	SpecialAssessmentList[30470259] = SpecialAssessmentList[30470258]
	SpecialAssessmentList[30434923] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((8 * 60) + 30) * 60, ((11 * 60) + 00) * 60, "Precalc 1A", "Manuel", "Library"}
	SpecialAssessmentList[30434924] = SpecialAssessmentList[30434923]
	SpecialAssessmentList[30434920] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((8 * 60) + 30) * 60, ((11 * 60) + 00) * 60, "Precalc 1A", "Gomprecht", "Library"}
	SpecialAssessmentList[30434922] = SpecialAssessmentList[30434920]
	SpecialAssessmentList[30434928] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((8 * 60) + 30) * 60, ((11 * 60) + 00) * 60, "Precalc 2A", "Gomprecht", "503 or 507"}
	SpecialAssessmentList[30434930] = SpecialAssessmentList[30434928]
	SpecialAssessmentList[30434931] = SpecialAssessmentList[30434928]
	SpecialAssessmentList[30434932] = SpecialAssessmentList[30434928]

	// second block
	// Foundations in Geo	Redl	150
	SpecialAssessmentList[30434692] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((12 * 60) + 00) * 60, ((14 * 60) + 00) * 60, "Geometry", "Gomprecht", "Library"}
	SpecialAssessmentList[30434694] = SpecialAssessmentList[30434692]
	SpecialAssessmentList[30504400] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((12 * 60) + 00) * 60, ((14 * 60) + 00) * 60, "PBL Geo/Geo A", "Borenstein/Bensky", "Library"}
	SpecialAssessmentList[30504401] = SpecialAssessmentList[30504400]
	SpecialAssessmentList[30504409] = SpecialAssessmentList[30504400]
	SpecialAssessmentList[30504410] = SpecialAssessmentList[30504400]
	// these next three are listed as Harvey/Bensky, Bensky, and Harvey/Redl
	// unfortunately because of how the class owners are tracked, I can't distinguish the three classes
	// fortunately it doesn't matter because they're all at the same time and the same place
	SpecialAssessmentList[30504402] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((12 * 60) + 00) * 60, ((14 * 60) + 00) * 60, "PBL Geo/Geo A", "Harvey/Bensky/Redl", "Library"}
	SpecialAssessmentList[30504403] = SpecialAssessmentList[30504402]
	SpecialAssessmentList[30504404] = SpecialAssessmentList[30504402]
	SpecialAssessmentList[30504405] = SpecialAssessmentList[30504402]
	SpecialAssessmentList[30504406] = SpecialAssessmentList[30504402]
	SpecialAssessmentList[30504407] = SpecialAssessmentList[30504402]
	SpecialAssessmentList[30504407] = SpecialAssessmentList[30504411]
	SpecialAssessmentList[30504407] = SpecialAssessmentList[30504412]
	SpecialAssessmentList[30504402] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((12 * 60) + 00) * 60, ((14 * 60) + 00) * 60, "Algebra 2/2 A", "Harvey/Bensky/Redl", "Library"}
	SpecialAssessmentList[30504403] = SpecialAssessmentList[30504402]
	// Algebra 2/2 A	Harvey/Cummings	Library
	SpecialAssessmentList[30434656] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((12 * 60) + 00) * 60, ((14 * 60) + 00) * 60, "Algebra 2", "Metz", "Library"}
	SpecialAssessmentList[30434658] = SpecialAssessmentList[30434656]
	SpecialAssessmentList[30434706] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((12 * 60) + 00) * 60, ((14 * 60) + 00) * 60, "Euclidean Geo A", "Sturm", "503"}
	SpecialAssessmentList[30434708] = SpecialAssessmentList[30434706]
	SpecialAssessmentList[30434650] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((12 * 60) + 00) * 60, ((14 * 60) + 00) * 60, "Algebra 1/2", "Cummings", "251"}
	SpecialAssessmentList[30434652] = SpecialAssessmentList[30434650]
	SpecialAssessmentList[30434656] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((12 * 60) + 00) * 60, ((14 * 60) + 00) * 60, "Algebra 2/2A", "Metz/Manuel/Bensky", "B06"}
	SpecialAssessmentList[30434658] = SpecialAssessmentList[30434656]
	SpecialAssessmentList[30434686] = SpecialAssessmentList[30434656]
	SpecialAssessmentList[30434688] = SpecialAssessmentList[30434656]
	SpecialAssessmentList[30470254] = SpecialAssessmentList[30434656]
	SpecialAssessmentList[30470255] = SpecialAssessmentList[30434656]
	SpecialAssessmentList[30434662] = &SpecialAssessmentInfo{SpecialAssessmentType_Math, ((12 * 60) + 00) * 60, ((14 * 60) + 00) * 60, "PBL Algebra 2/2A", "Bensky/Cohen", "507"}
	SpecialAssessmentList[30434664] = SpecialAssessmentList[30434662]
	SpecialAssessmentList[30434674] = SpecialAssessmentList[30434662]
	SpecialAssessmentList[30434676] = SpecialAssessmentList[30434662]

	// ** science **

	// first block
	SpecialAssessmentList[30434907] = &SpecialAssessmentInfo{SpecialAssessmentType_Science, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Biology", "Cuello", "Cafeteria"}
	SpecialAssessmentList[30434908] = SpecialAssessmentList[30434907]
	SpecialAssessmentList[30434909] = &SpecialAssessmentInfo{SpecialAssessmentType_Science, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Biology", "Reid", "Cafeteria"}
	SpecialAssessmentList[30434910] = SpecialAssessmentList[30434909]
	SpecialAssessmentList[30435290] = &SpecialAssessmentInfo{SpecialAssessmentType_Science, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Chemistry", "Ramsey", "Cafeteria"}
	SpecialAssessmentList[30435292] = SpecialAssessmentList[30435290]
	SpecialAssessmentList[30435307] = &SpecialAssessmentInfo{SpecialAssessmentType_Science, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Chemistry A", "Ramsey", "Cafeteria"}
	SpecialAssessmentList[30435308] = SpecialAssessmentList[30435307]
	SpecialAssessmentList[30470282] = SpecialAssessmentList[30435307]
	SpecialAssessmentList[30470283] = SpecialAssessmentList[30435307]
	SpecialAssessmentList[30435293] = &SpecialAssessmentInfo{SpecialAssessmentType_Science, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Chemistry", "Hackett", "Cafeteria"}
	SpecialAssessmentList[30435294] = SpecialAssessmentList[30435293]
	SpecialAssessmentList[30435299] = SpecialAssessmentList[30435293]
	SpecialAssessmentList[30435300] = SpecialAssessmentList[30435293]

	SpecialAssessmentList[30435295] = &SpecialAssessmentInfo{SpecialAssessmentType_Science, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Chemistry", "Taylor", "Library"}
	SpecialAssessmentList[30435296] = SpecialAssessmentList[30435295]
	SpecialAssessmentList[30435304] = &SpecialAssessmentInfo{SpecialAssessmentType_Science, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Chemistry A", "Taylor", "Library"}
	SpecialAssessmentList[30435306] = SpecialAssessmentList[30435304]
	SpecialAssessmentList[30434913] = &SpecialAssessmentInfo{SpecialAssessmentType_Science, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Biology", "Brizzolara", "Library"}
	SpecialAssessmentList[30434914] = SpecialAssessmentList[30434913]
	SpecialAssessmentList[30434915] = SpecialAssessmentList[30434913]
	SpecialAssessmentList[30434916] = SpecialAssessmentList[30434913]
	SpecialAssessmentList[30434905] = &SpecialAssessmentInfo{SpecialAssessmentType_Science, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Biology", "Schollenberger", "Library"}
	SpecialAssessmentList[30434906] = SpecialAssessmentList[30434905]
	SpecialAssessmentList[30434911] = SpecialAssessmentList[30434905]
	SpecialAssessmentList[30434912] = SpecialAssessmentList[30434905]
	// Conceptual Chemistry	Fenton	Library

	SpecialAssessmentList[30435360] = &SpecialAssessmentInfo{SpecialAssessmentType_Science, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Human Physiology", "Geller", "405/413"}
	SpecialAssessmentList[30435362] = SpecialAssessmentList[30435360]
	SpecialAssessmentList[30434902] = &SpecialAssessmentInfo{SpecialAssessmentType_Science, ((8 * 60) + 15) * 60, ((10 * 60) + 15) * 60, "Biology", "Geller", "405/413"}
	SpecialAssessmentList[30434904] = SpecialAssessmentList[30434902]

	// second block
	SpecialAssessmentList[30435312] = &SpecialAssessmentInfo{SpecialAssessmentType_Science, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Physics/Physics A", "Atary/Bove/Condie/Francischelli", "Library"}
	SpecialAssessmentList[30435314] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[30435315] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[30435316] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[30435317] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[30435318] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[30435319] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[30435320] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[30435321] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[30435322] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[30435326] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[30435328] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[30435329] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[30435330] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[108404024] = SpecialAssessmentList[30435312]
	SpecialAssessmentList[108404025] = SpecialAssessmentList[30435312]

	// ** language **

	// first block
	SpecialAssessmentList[30435170] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "Spanish 1", "Berman", "Library"}
	SpecialAssessmentList[30435172] = SpecialAssessmentList[30435170]
	SpecialAssessmentList[30435176] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "Spanish 2", "Berman/Nebres", "Library"}
	SpecialAssessmentList[30435178] = SpecialAssessmentList[30435176]
	SpecialAssessmentList[30435179] = SpecialAssessmentList[30435176]
	SpecialAssessmentList[30435180] = SpecialAssessmentList[30435176]
	SpecialAssessmentList[30435184] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "Spanish 3", "San Juan/Nebres", "Library"}
	SpecialAssessmentList[30435186] = SpecialAssessmentList[30435184]
	SpecialAssessmentList[30435187] = SpecialAssessmentList[30435184]
	SpecialAssessmentList[30435188] = SpecialAssessmentList[30435184]
	SpecialAssessmentList[30435194] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((8 * 60) + 30) * 60, ((10 * 60) + 30) * 60, "Spanish 3A", "San Juan/David", "Library"}
	SpecialAssessmentList[30435196] = SpecialAssessmentList[30435194]
	SpecialAssessmentList[30435197] = SpecialAssessmentList[30435194]
	SpecialAssessmentList[30435198] = SpecialAssessmentList[30435194]

	// second block
	SpecialAssessmentList[30435116] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Mandarin 1", "Hwang-Dolak", "Library"}
	SpecialAssessmentList[30435118] = SpecialAssessmentList[30435116]
	SpecialAssessmentList[30435122] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Mandarin 2", "Wu", "Library"}
	SpecialAssessmentList[30435124] = SpecialAssessmentList[30435122]
	SpecialAssessmentList[30435128] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Mandarin 3", "Hwang-Dolak/Lanphier", "Library"}
	SpecialAssessmentList[30435130] = SpecialAssessmentList[30435128]
	SpecialAssessmentList[30470270] = SpecialAssessmentList[30435128]
	SpecialAssessmentList[30470271] = SpecialAssessmentList[30435128]
	SpecialAssessmentList[30435050] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Latin 1", "Braff", "Library"}
	SpecialAssessmentList[30435052] = SpecialAssessmentList[30435050]
	SpecialAssessmentList[30435056] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Latin 2", "Thornton", "Library"}
	SpecialAssessmentList[30435058] = SpecialAssessmentList[30435056]
	SpecialAssessmentList[30435062] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "Latin 3", "Braff", "Library"}
	SpecialAssessmentList[30435064] = SpecialAssessmentList[30435062]
	SpecialAssessmentList[30434998] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "French 1", "Bendali", "Library"}
	SpecialAssessmentList[30435000] = SpecialAssessmentList[30434998]
	// French 2	Bendali	Library
	SpecialAssessmentList[30435010] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "French 3", "Albino", "Library"}
	SpecialAssessmentList[30435012] = SpecialAssessmentList[30435010]
	SpecialAssessmentList[30435018] = &SpecialAssessmentInfo{SpecialAssessmentType_Language, ((11 * 60) + 30) * 60, ((13 * 60) + 30) * 60, "French 3A", "Mhinat", "Library"}
	SpecialAssessmentList[30435020] = SpecialAssessmentList[30435018]
}
