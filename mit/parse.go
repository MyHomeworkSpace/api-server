package mit

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var parensRegex = regexp.MustCompile("\\((.*?)\\)")

// A ScheduledMeeting struct contains details about a scheduled meeting in a TimeInfo. Each TimeInfo can have multiple ScheduledMeetings.
type ScheduledMeeting struct {
	Weekdays     []time.Weekday
	StartSeconds int
	EndSeconds   int
}

// A TimeInfo struct contains details about when a class's section meets.
type TimeInfo struct {
	Meetings []ScheduledMeeting
	BeginsOn time.Time
	EndsOn   time.Time
}

func parseTime(timeString string, forceAM bool) (int, error) {
	// sometimes, for whatever reason, use a colon instead of the normal dot as a separator
	// this seems to just be some advising seminars
	// replace these colons with a dot so that we can handle them
	timeString = strings.Replace(timeString, ":", ".", -1)

	// some classes also like to just randomly add a "pm"
	// looking at you WGS.228 and 15.389
	isPM := false
	if strings.Contains(timeString, "pm") || strings.Contains(timeString, "PM") {
		timeString = strings.Replace(strings.Replace(timeString, "pm", "", -1), "PM", "", -1)
		isPM = true
	}

	// also check for just "p"
	if strings.Contains(timeString, "p") || strings.Contains(timeString, "P") {
		timeString = strings.Replace(strings.Replace(timeString, "p", "", -1), "P", "", -1)
		isPM = true
	}

	// it's a time like "4" or "5.30"
	parts := strings.Split(timeString, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("mit: parseTime: time string '%s' had too many dots", timeString)
	}

	hour, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, err
	}

	minute := 0

	if len(parts) == 2 {
		minute, err = strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil {
			return 0, err
		}
	}

	if forceAM || (!isPM && hour >= 8 && hour <= 12) {
		// assume AM
	} else {
		// assume PM
		hour += 12
		if hour == 24 {
			// 12 PM is actually noon
			hour = 12
		}
	}

	timeSeconds := ((hour * 60) + minute) * 60

	return timeSeconds, nil
}

func assignYearForDate(date time.Time, termInfo TermInfo) time.Time {
	return date.AddDate(termInfo.FirstDayOfClasses.Year(), 0, 0)
}

// ParseScheduledMeeting parses the given time info string, in a format like "MW4-5.30"
func ParseScheduledMeeting(scheduledMeetingString string, termInfo TermInfo) (*ScheduledMeeting, *time.Time, *time.Time, error) {
	scheduledMeeting := ScheduledMeeting{
		Weekdays: []time.Weekday{},
	}

	var beginsOn *time.Time
	var endsOn *time.Time

	normalizedMeetingString := strings.ToUpper(strings.TrimSpace(strings.Replace(scheduledMeetingString, "*", "", -1)))

	if normalizedMeetingString == "SECTION CANCELLED" || normalizedMeetingString == "TO BE ARRANGED" || normalizedMeetingString == "TBA" || normalizedMeetingString == "TBD" || normalizedMeetingString == "ARRANGED" {
		// oof
		return nil, nil, nil, nil
	}

	parsedATime := false

	if strings.Contains(scheduledMeetingString, "(") {
		// there's a thing in parentheses
		// for example "TR10.30-12 (BEGINS OCT 21)"

		// it could be an evening class
		// for example "TR EVE (4-6 PM)" or "W EVE (4-6.30 PM)"
		isEvening := (strings.Contains(scheduledMeetingString, "EVE"))

		matches := parensRegex.FindAllStringSubmatch(scheduledMeetingString, -1)
		for i, match := range matches {
			info := match[1]

			if i == 0 && isEvening {
				// it's an evening class
				// the first paren is a time
				info = strings.Replace(info, " PM", "", -1)

				subScheduledMeeting, _, _, err := ParseScheduledMeeting(info, termInfo)
				if err != nil {
					return nil, nil, nil, err
				}

				if subScheduledMeeting.StartSeconds < 12*60*60 {
					subScheduledMeeting.StartSeconds += 12 * 60 * 60
				} else if subScheduledMeeting.EndSeconds < 12*60*60 {
					subScheduledMeeting.EndSeconds += 12 * 60 * 60
				}

				scheduledMeeting.StartSeconds = subScheduledMeeting.StartSeconds
				scheduledMeeting.EndSeconds = subScheduledMeeting.EndSeconds

				scheduledMeetingString = strings.Replace(scheduledMeetingString, "EVE", "", -1)
				scheduledMeetingString = parensRegex.ReplaceAllString(scheduledMeetingString, "")

				parsedATime = true

				continue
			}

			if strings.HasPrefix(info, "BEGINS") {
				dateString := strings.Replace(info, "BEGINS ", "", -1)
				parsedDate, err := time.Parse("Jan 2", dateString)
				if err != nil {
					return nil, nil, nil, err
				}
				parsedDate = assignYearForDate(parsedDate, termInfo)
				beginsOn = &parsedDate
			} else if strings.HasPrefix(info, "END") {
				dateString := strings.Replace(info, "ENDS ", "", -1)
				parsedDate, err := time.Parse("Jan 2", dateString)
				if err != nil {
					return nil, nil, nil, err
				}
				parsedDate = assignYearForDate(parsedDate, termInfo)
				endsOn = &parsedDate
			} else if strings.HasPrefix(info, "MEETS") {
				// it's a message in the format "MEETS 4/7 TO 5/14"
				dateString := strings.Replace(info, "MEETS ", "", -1)
				parts := strings.Split(dateString, "TO")

				parsedBegin, err := time.Parse("1/2", strings.TrimSpace(parts[0]))
				if err != nil {
					return nil, nil, nil, err
				}
				parsedBegin = assignYearForDate(parsedBegin, termInfo)
				beginsOn = &parsedBegin

				parsedEnd, err := time.Parse("1/2", strings.TrimSpace(parts[1]))
				if err != nil {
					return nil, nil, nil, err
				}
				parsedEnd = assignYearForDate(parsedEnd, termInfo)
				endsOn = &parsedEnd
			} else if strings.HasPrefix(info, "LIMITED") {
				// it's about limited enrollment, which we don't care about
				// just ignore it
			} else {
				// some classes are not entered into the catalog correctly
				// they have a format like T(11-12:30)
				// if we can parse the time in parenthesis, assume that's the problem
				parsedInfo, _, _, err := ParseScheduledMeeting(info, termInfo)
				if err != nil {
					// didn't work, give up
					return nil, nil, nil, fmt.Errorf("mit: ParseScheduledMeeting: couldn't handle info '%s'", info)
				}

				// it worked, so keep going
				if parsedInfo.StartSeconds != 0 {
					scheduledMeeting.StartSeconds = parsedInfo.StartSeconds
				}
				if parsedInfo.EndSeconds != 0 {
					scheduledMeeting.EndSeconds = parsedInfo.EndSeconds
				}
			}
		}

		scheduledMeetingString = parensRegex.ReplaceAllString(scheduledMeetingString, "")
	}

	scheduledMeetingString = strings.TrimSpace(scheduledMeetingString)

	charMap := map[rune]time.Weekday{
		'M': time.Monday,
		'T': time.Tuesday,
		'W': time.Wednesday,
		'R': time.Thursday,
		'F': time.Friday,
		'S': time.Saturday,
	}

	remainingScheduledMeetingString := ""

	// first parse out all the weekdays
	for i, scheduledMeetingRune := range scheduledMeetingString {
		weekday, isWeekday := charMap[scheduledMeetingRune]
		if isWeekday {
			scheduledMeeting.Weekdays = append(scheduledMeeting.Weekdays, weekday)
		} else {
			// we're done
			remainingScheduledMeetingString = scheduledMeetingString[i:]
			break
		}
	}

	var err error
	if remainingScheduledMeetingString != "" && !parsedATime {
		if strings.Contains(remainingScheduledMeetingString, "-") {
			// it's a time like "4-5.30"
			forceAM := false
			if strings.Contains(remainingScheduledMeetingString, "AM") {
				forceAM = true
				remainingScheduledMeetingString = strings.Replace(remainingScheduledMeetingString, "AM", "", -1)
			}

			// HACK: 18.204 has a string in the format "MW2.30-4 (LIMITED 15 EACH S .."
			// we detect this and ignore the parenthesis
			if strings.Contains(remainingScheduledMeetingString, "(") && !strings.Contains(remainingScheduledMeetingString, ")") {
				remainingScheduledMeetingString = strings.Split(remainingScheduledMeetingString, "(")[0]
			}

			timeParts := strings.Split(remainingScheduledMeetingString, "-")
			if len(timeParts) != 2 {
				return nil, nil, nil, fmt.Errorf("mit: ParseScheduledMeeting: time info string '%s' had too many dashes", scheduledMeetingString)
			}

			scheduledMeeting.StartSeconds, err = parseTime(timeParts[0], forceAM)
			if err != nil {
				return nil, nil, nil, err
			}

			scheduledMeeting.EndSeconds, err = parseTime(timeParts[1], forceAM)
			if err != nil {
				return nil, nil, nil, err
			}
		} else {
			// it's a time like "4"
			scheduledMeeting.StartSeconds, err = parseTime(remainingScheduledMeetingString, false)
			if err != nil {
				return nil, nil, nil, err
			}

			// we can assume it's an hour long
			scheduledMeeting.EndSeconds = scheduledMeeting.StartSeconds + 60*60
		}
	} else {
		if strings.TrimSpace(remainingScheduledMeetingString) != "" && parsedATime {
			// there's something left, and it can't be a time...
			// complain
			return nil, nil, nil, fmt.Errorf("mit: unexpected extra data '%s' when parsing time string '%s'", remainingScheduledMeetingString, scheduledMeetingString)
		}
	}

	return &scheduledMeeting, beginsOn, endsOn, nil
}

// ParseTimeInfo parses out the time information from an input string in the format "MW9-11,F9".
func ParseTimeInfo(timeInfoString string, termInfo TermInfo) (TimeInfo, error) {
	timeInfo := TimeInfo{
		Meetings: []ScheduledMeeting{},
		BeginsOn: termInfo.FirstDayOfClasses,
		EndsOn:   termInfo.LastDayOfClasses,
	}

	// hack to cope with the fact that some people really, really don't understand how the course catalog works
	// this is specifically for AS.301
	timeInfoString = strings.Replace(timeInfoString, "Thurs", "R", -1)

	timeInfoParts := strings.Split(timeInfoString, ",")
	for _, part := range timeInfoParts {
		meeting, beginsOn, endsOn, err := ParseScheduledMeeting(part, termInfo)
		if err != nil {
			return TimeInfo{}, err
		}

		if beginsOn != nil {
			timeInfo.BeginsOn = *beginsOn
		}
		if endsOn != nil {
			timeInfo.EndsOn = *endsOn
		}

		if meeting != nil {
			timeInfo.Meetings = append(timeInfo.Meetings, *meeting)
		}
	}

	return timeInfo, nil
}
