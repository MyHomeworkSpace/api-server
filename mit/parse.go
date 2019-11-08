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

func parseTime(timeString string) (int, error) {
	// sometimes, for whatever reason, use a colon instead of the normal dot as a separator
	// this seems to just be some advising seminars
	// replace these colons with a dot so that we can handle them
	timeString = strings.Replace(timeString, ":", ".", -1)

	// some classes also like to just randomly add a "pm"
	// looking at you WGS.228
	isPM := false
	if strings.Contains(timeString, "pm") {
		timeString = strings.Replace(timeString, "pm", "", -1)
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

	if !isPM && hour >= 8 && hour <= 12 {
		// assume AM
	} else {
		// assume PM
		hour += 12
	}

	timeSeconds := ((hour * 60) + minute) * 60

	return timeSeconds, nil
}

// ParseScheduledMeeting parses the given time info string, in a format like "MW4-5.30"
func ParseScheduledMeeting(scheduledMeetingString string, termInfo TermInfo) (*ScheduledMeeting, *time.Time, *time.Time, error) {
	scheduledMeeting := ScheduledMeeting{
		Weekdays: []time.Weekday{},
	}

	var beginsOn *time.Time
	var endsOn *time.Time

	if strings.TrimSpace(strings.ToUpper(scheduledMeetingString)) == "*TO BE ARRANGED" || strings.TrimSpace(strings.ToUpper(scheduledMeetingString)) == "TBA" || strings.TrimSpace(strings.ToUpper(scheduledMeetingString)) == "TBD" {
		// oof
		return nil, nil, nil, nil
	}

	if strings.Contains(scheduledMeetingString, "EVE") {
		// it's an evening class
		// for example "TR EVE (4-6 PM)" or "W EVE (4-6.30 PM)"

		matches := parensRegex.FindAllStringSubmatch(scheduledMeetingString, -1)
		if len(matches) != 1 {
			return nil, nil, nil, fmt.Errorf("mit: ParseScheduledMeeting: time info string '%s' had unexpected number of parens", scheduledMeetingString)
		}

		time := matches[0][1]
		time = strings.Replace(time, " PM", "", -1)

		subScheduledMeeting, _, _, err := ParseScheduledMeeting(time, termInfo)
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
	} else if strings.Contains(scheduledMeetingString, "(") {
		// there's a thing in parentheses
		// for example "TR10.30-12 (BEGINS OCT 21)"
		matches := parensRegex.FindAllStringSubmatch(scheduledMeetingString, -1)
		for _, match := range matches {
			info := match[1]

			if strings.HasPrefix(info, "BEGINS") {
				dateString := strings.Replace(info, "BEGINS ", "", -1)
				parsedDate, err := time.Parse("Jan 2", dateString)
				if err != nil {
					return nil, nil, nil, err
				}
				parsedDate = parsedDate.AddDate(termInfo.FirstDayOfClasses.Year(), 0, 0)
				beginsOn = &parsedDate
			} else if strings.HasPrefix(info, "END") {
				dateString := strings.Replace(info, "ENDS ", "", -1)
				parsedDate, err := time.Parse("Jan 2", dateString)
				if err != nil {
					return nil, nil, nil, err
				}
				parsedDate = parsedDate.AddDate(termInfo.FirstDayOfClasses.Year(), 0, 0)
				endsOn = &parsedDate
			} else {
				return nil, nil, nil, fmt.Errorf("mit: ParseScheduledMeeting: couldn't handle info '%s'", info)
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
	if remainingScheduledMeetingString != "" {
		if strings.Contains(remainingScheduledMeetingString, "-") {
			// it's a time like "4-5.30"
			timeParts := strings.Split(remainingScheduledMeetingString, "-")
			if len(timeParts) != 2 {
				return nil, nil, nil, fmt.Errorf("mit: ParseScheduledMeeting: time info string '%s' had too many dashes", scheduledMeetingString)
			}

			scheduledMeeting.StartSeconds, err = parseTime(timeParts[0])
			if err != nil {
				return nil, nil, nil, err
			}

			scheduledMeeting.EndSeconds, err = parseTime(timeParts[1])
			if err != nil {
				return nil, nil, nil, err
			}
		} else {
			// it's a time like "4"
			scheduledMeeting.StartSeconds, err = parseTime(remainingScheduledMeetingString)
			if err != nil {
				return nil, nil, nil, err
			}

			// we can assume it's an hour long
			scheduledMeeting.EndSeconds = scheduledMeeting.StartSeconds + 60*60
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