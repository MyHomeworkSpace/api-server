package calendar

import (
	"time"
)

// these change every year
var (
	// import ranges
	// these should be ranges with 4 fridays in a row and the first week having no off days
	Term1_Import_Start = time.Date(2018, time.September, 24, 0, 0, 0, 0, time.UTC)
	Term1_Import_End   = time.Date(2018, time.October, 20, 0, 0, 0, 0, time.UTC)

	Term1_Import_DayOffset_Friday1 = ((7 * 2) + 4)
	Term1_Import_DayOffset_Friday2 = ((7 * 3) + 4)
	Term1_Import_DayOffset_Friday3 = 4
	Term1_Import_DayOffset_Friday4 = ((7 * 1) + 4)

	Term2_Import_Start = time.Date(2019, time.January, 28, 0, 0, 0, 0, time.UTC)
	Term2_Import_End   = time.Date(2019, time.February, 23, 0, 0, 0, 0, time.UTC)

	Term2_Import_DayOffset_Friday1 = ((7 * 3) + 4)
	Term2_Import_DayOffset_Friday2 = 4
	Term2_Import_DayOffset_Friday3 = ((7 * 1) + 4)
	Term2_Import_DayOffset_Friday4 = ((7 * 2) + 4)
)

func InitCalendar() {

}
