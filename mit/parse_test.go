package mit

import (
	"testing"
	"time"
)

var testTerm = TermInfo{
	Code:              "2021FA",
	FirstDayOfClasses: time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
	LastDayOfClasses:  time.Date(2020, 12, 9, 0, 0, 0, 0, time.UTC),
	ExceptionDays:     map[string]time.Weekday{},
}

func termDate(year int, month time.Month, day int) *time.Time {
	result := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &result
}

func compareWeekdays(a []time.Weekday, b []time.Weekday) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func testSingleTime(t *testing.T, input string, expectedScheduledMeeting *ScheduledMeeting, expectedStart *time.Time, expectedEnd *time.Time) {
	resultScheduledMeeting, resultStart, resultEnd, err := ParseScheduledMeeting(input, testTerm)
	if err != nil {
		t.Errorf("ParseScheduledMeeting('%s'): got error '%s'", input, err.Error())
		return
	}

	if !compareWeekdays(resultScheduledMeeting.Weekdays, expectedScheduledMeeting.Weekdays) {
		t.Errorf("ParseScheduledMeeting('%s'): ScheduledMeeting.Weekdays: got %#v, expected %#v", input, resultScheduledMeeting.Weekdays, expectedScheduledMeeting.Weekdays)
	}

	if resultScheduledMeeting.StartSeconds != expectedScheduledMeeting.StartSeconds {
		t.Errorf("ParseScheduledMeeting('%s'): ScheduledMeeting.StartSeconds: got %d, expected %d", input, resultScheduledMeeting.StartSeconds, expectedScheduledMeeting.StartSeconds)
	}

	if resultScheduledMeeting.EndSeconds != expectedScheduledMeeting.EndSeconds {
		t.Errorf("ParseScheduledMeeting('%s'): ScheduledMeeting.EndSeconds: got %d, expected %d", input, resultScheduledMeeting.EndSeconds, expectedScheduledMeeting.EndSeconds)
	}

	if (resultStart == nil && resultStart != expectedStart) || (resultStart != nil && *resultStart != *expectedStart) {
		t.Errorf("ParseScheduledMeeting('%s'): start date: got %s, expected %s", input, resultStart, expectedStart)
	}

	if (resultEnd == nil && resultEnd != expectedEnd) || (resultEnd != nil && *resultEnd != *expectedEnd) {
		t.Errorf("ParseScheduledMeeting('%s'): end date: got %s, expected %s", input, resultEnd, expectedEnd)
	}
}

const minuteSeconds = 60
const hourSeconds = 60 * minuteSeconds

func TestTimeParser(t *testing.T) {
	/*
	 * well-formed
	 */

	testSingleTime(t, "T3", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Tuesday},
		StartSeconds: 15 * hourSeconds,
		EndSeconds:   16 * hourSeconds,
	}, nil, nil)
	testSingleTime(t, "W3.30", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Wednesday},
		StartSeconds: (15 * hourSeconds) + (30 * minuteSeconds),
		EndSeconds:   (16 * hourSeconds) + (30 * minuteSeconds),
	}, nil, nil)
	testSingleTime(t, "R1.30-3.30", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Thursday},
		StartSeconds: (13 * hourSeconds) + (30 * minuteSeconds),
		EndSeconds:   (15 * hourSeconds) + (30 * minuteSeconds),
	}, nil, nil)
	testSingleTime(t, "S4", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Saturday},
		StartSeconds: 16 * hourSeconds,
		EndSeconds:   17 * hourSeconds,
	}, nil, nil)

	testSingleTime(t, "MTF11", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Monday, time.Tuesday, time.Friday},
		StartSeconds: 11 * hourSeconds,
		EndSeconds:   12 * hourSeconds,
	}, nil, nil)

	testSingleTime(t, "TR10.30-12 (BEGINS OCT 21)", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Tuesday, time.Thursday},
		StartSeconds: (10 * hourSeconds) + (30 * minuteSeconds),
		EndSeconds:   12 * hourSeconds,
	}, termDate(2020, time.October, 21), nil)

	testSingleTime(t, "RF11.30 (ENDS DEC 2)", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Thursday, time.Friday},
		StartSeconds: (11 * hourSeconds) + (30 * minuteSeconds),
		EndSeconds:   (12 * hourSeconds) + (30 * minuteSeconds),
	}, nil, termDate(2020, time.December, 2))

	testSingleTime(t, "MT9 (MEETS 9/4 TO 10/6)", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Monday, time.Tuesday},
		StartSeconds: 9 * hourSeconds,
		EndSeconds:   10 * hourSeconds,
	}, termDate(2020, time.September, 4), termDate(2020, time.October, 6))

	testSingleTime(t, "WF EVE (5-7)", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Wednesday, time.Friday},
		StartSeconds: 17 * hourSeconds,
		EndSeconds:   19 * hourSeconds,
	}, nil, nil)
	testSingleTime(t, "TR EVE (4-6 PM)", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Tuesday, time.Thursday},
		StartSeconds: 16 * hourSeconds,
		EndSeconds:   18 * hourSeconds,
	}, nil, nil)
	testSingleTime(t, "MW EVE (4.30-5.30 PM) (BEGINS NOV 2)", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Monday, time.Wednesday},
		StartSeconds: (16 * hourSeconds) + (30 * minuteSeconds),
		EndSeconds:   (17 * hourSeconds) + (30 * minuteSeconds),
	}, termDate(2020, time.November, 2), nil)

	testSingleTime(t, "MW10 (LIMITED)", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Monday, time.Wednesday},
		StartSeconds: 10 * hourSeconds,
		EndSeconds:   11 * hourSeconds,
	}, nil, nil)

	testSingleTime(t, "WR3.30-4.45 (LIMITED) (BEGINS OCT 14) (ENDS OCT 23)", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Wednesday, time.Thursday},
		StartSeconds: (15 * hourSeconds) + (30 * minuteSeconds),
		EndSeconds:   (16 * hourSeconds) + (45 * minuteSeconds),
	}, termDate(2020, time.October, 14), termDate(2020, time.October, 23))

	testSingleTime(t, "WR EVE (3.30-4.45 PM) (LIMITED) (BEGINS OCT 14) (ENDS OCT 23)", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Wednesday, time.Thursday},
		StartSeconds: (15 * hourSeconds) + (30 * minuteSeconds),
		EndSeconds:   (16 * hourSeconds) + (45 * minuteSeconds),
	}, termDate(2020, time.October, 14), termDate(2020, time.October, 23))

	/*
	 * not well-formed
	 */

	testSingleTime(t, "W 4:30", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Wednesday},
		StartSeconds: (16 * hourSeconds) + (30 * minuteSeconds),
		EndSeconds:   (17 * hourSeconds) + (30 * minuteSeconds),
	}, nil, nil)
	testSingleTime(t, "TR 10:30-12p", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Tuesday, time.Thursday},
		StartSeconds: (10 * hourSeconds) + (30 * minuteSeconds),
		EndSeconds:   12 * hourSeconds,
	}, nil, nil)
	testSingleTime(t, "TTH 10:30 - 12 PM", &ScheduledMeeting{
		Weekdays:     []time.Weekday{time.Tuesday, time.Thursday},
		StartSeconds: (10 * hourSeconds) + (30 * minuteSeconds),
		EndSeconds:   12 * hourSeconds,
	}, nil, nil)
}
