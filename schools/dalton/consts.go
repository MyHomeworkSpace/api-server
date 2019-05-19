package dalton

import "time"

// An AssemblyType describes what happens for assembly on a given week.
type AssemblyType int

const (
	AssemblyType_Assembly AssemblyType = iota
	AssemblyType_LongHouse
	AssemblyType_Lab
)

// these change every year
var (
	Day_SchoolStart, _   = time.Parse("2006-01-02", "2018-09-12")
	Day_ExamRelief, _    = time.Parse("2006-01-02", "2019-01-25")
	Day_SeniorLastDay, _ = time.Parse("2006-01-02", "2019-04-26")
	Day_SchoolEnd, _     = time.Parse("2006-01-02", "2019-06-06")

	AssemblyTypeList = map[string]AssemblyType{
		"2018-09-13": AssemblyType_Assembly,
		"2018-09-20": AssemblyType_Assembly,
		"2018-09-27": AssemblyType_LongHouse,
		"2018-10-04": AssemblyType_Assembly,
		"2018-10-11": AssemblyType_Lab,
		"2018-10-18": AssemblyType_Assembly,
		"2018-10-25": AssemblyType_LongHouse,
		"2018-11-01": AssemblyType_Assembly,
		"2018-11-08": AssemblyType_Lab,
		"2018-11-15": AssemblyType_Assembly,
		"2018-11-29": AssemblyType_Lab,
		"2018-12-06": AssemblyType_Assembly,
		"2018-12-13": AssemblyType_LongHouse,
		"2018-12-20": AssemblyType_Assembly,
		"2019-01-10": AssemblyType_Lab,
		"2019-01-31": AssemblyType_Assembly,
		"2019-02-07": AssemblyType_Lab,
		"2019-02-14": AssemblyType_Assembly,
		"2019-02-21": AssemblyType_Lab,
		"2019-02-28": AssemblyType_Assembly,
		"2019-03-07": AssemblyType_Assembly,
		"2019-03-14": AssemblyType_Assembly,
		"2019-04-04": AssemblyType_Lab,
		"2019-04-11": AssemblyType_Assembly,
		"2019-04-18": AssemblyType_Assembly,
		"2019-04-25": AssemblyType_LongHouse,
		"2019-05-02": AssemblyType_Assembly,
		"2019-05-09": AssemblyType_Lab,
		"2019-05-16": AssemblyType_Assembly,
		"2019-05-23": AssemblyType_Assembly,
	}
)
