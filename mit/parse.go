package mit

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var parensRegex = regexp.MustCompile("\\((.*?)\\)")

// A TimeInfo struct contains details about when a class's section meets.
type TimeInfo struct {
	Weekdays     []time.Weekday
	StartSeconds int
	EndSeconds   int
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

// ParseTimeInfo parses the given time info string, in a format like "MW4-5.30"
func ParseTimeInfo(timeInfoString string) (*TimeInfo, error) {
	timeInfo := TimeInfo{
		Weekdays: []time.Weekday{},
	}

	if strings.TrimSpace(strings.ToUpper(timeInfoString)) == "*TO BE ARRANGED" || strings.TrimSpace(strings.ToUpper(timeInfoString)) == "TBA" || strings.TrimSpace(strings.ToUpper(timeInfoString)) == "TBD" {
		// oof
		return nil, nil
	}

	if strings.Contains(timeInfoString, "EVE") {
		// it's an evening class
		// for example "TR EVE (4-6 PM)" or "W EVE (4-6.30 PM)"

		matches := parensRegex.FindAllStringSubmatch(timeInfoString, -1)
		if len(matches) != 1 {
			return nil, fmt.Errorf("mit: ParseTimeInfo: time info string '%s' had unexpected number of parens", timeInfoString)
		}

		time := matches[0][1]
		time = strings.Replace(time, " PM", "", -1)

		subtimeInfo, err := ParseTimeInfo(time)
		if err != nil {
			return nil, err
		}

		if subtimeInfo.StartSeconds < 12*60*60 {
			subtimeInfo.StartSeconds += 12 * 60 * 60
		} else if subtimeInfo.EndSeconds < 12*60*60 {
			subtimeInfo.EndSeconds += 12 * 60 * 60
		}

		timeInfo.StartSeconds = subtimeInfo.StartSeconds
		timeInfo.EndSeconds = subtimeInfo.EndSeconds

		timeInfoString = strings.Replace(timeInfoString, "EVE", "", -1)
		timeInfoString = parensRegex.ReplaceAllString(timeInfoString, "")
	} else if strings.Contains(timeInfoString, "(") {
		// there's a thing in parentheses
		// for example "TR10.30-12 (BEGINS OCT 21)"
		matches := parensRegex.FindAllStringSubmatch(timeInfoString, -1)
		for _, match := range matches {
			info := match[1]

			if strings.HasPrefix(info, "BEGINS") {
				// TODO: handle
			} else if strings.HasPrefix(info, "END") {
				// TODO: handle
			} else {
				return nil, fmt.Errorf("mit: ParseTimeInfo: couldn't handle info '%s'", info)
			}
		}

		timeInfoString = parensRegex.ReplaceAllString(timeInfoString, "")
	}

	timeInfoString = strings.TrimSpace(timeInfoString)

	charMap := map[rune]time.Weekday{
		'M': time.Monday,
		'T': time.Tuesday,
		'W': time.Wednesday,
		'R': time.Thursday,
		'F': time.Friday,
	}

	remainingTimeInfoString := ""

	// first parse out all the weekdays
	for i, timeInfoRune := range timeInfoString {
		weekday, isWeekday := charMap[timeInfoRune]
		if isWeekday {
			timeInfo.Weekdays = append(timeInfo.Weekdays, weekday)
		} else {
			// we're done
			remainingTimeInfoString = timeInfoString[i:]
			break
		}
	}

	var err error
	if remainingTimeInfoString != "" {
		if strings.Contains(remainingTimeInfoString, "-") {
			// it's a time like "4-5.30"
			timeParts := strings.Split(remainingTimeInfoString, "-")
			if len(timeParts) != 2 {
				return nil, fmt.Errorf("mit: ParseTimeInfo: time info string '%s' had too many dashes", timeInfoString)
			}

			timeInfo.StartSeconds, err = parseTime(timeParts[0])
			if err != nil {
				return nil, err
			}

			timeInfo.EndSeconds, err = parseTime(timeParts[1])
			if err != nil {
				return nil, err
			}
		} else {
			// it's a time like "4"
			timeInfo.StartSeconds, err = parseTime(remainingTimeInfoString)
			if err != nil {
				return nil, err
			}

			// we can assume it's an hour long
			timeInfo.EndSeconds = timeInfo.StartSeconds + 60*60
		}
	}

	return &timeInfo, nil
}

// ParseAllTimeInfo parses out all the time information from an input string in the format "MW9-11,F9".
func ParseAllTimeInfo(timeInfo string) ([]TimeInfo, error) {
	allTimeInfo := []TimeInfo{}
	timeInfoParts := strings.Split(timeInfo, ",")
	for _, part := range timeInfoParts {
		info, err := ParseTimeInfo(part)
		if err != nil {
			return nil, err
		}

		if info != nil {
			allTimeInfo = append(allTimeInfo, *info)
		}
	}

	return allTimeInfo, nil
}
