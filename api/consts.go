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
)
