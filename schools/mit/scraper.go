package mit

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/errorlog"

	"github.com/PuerkitoBio/goquery"

	"github.com/thatoddmailbox/touchstone-client/touchstone"
)

const (
	studentDashboardURL                string = "https://student-dashboard.mit.edu"
	studentDashboardAcademicProfileURL string = studentDashboardURL + "/service/apis/students/students/academicProfile"
	studentDashboardRegistrationURL    string = studentDashboardURL + "/service/apis/studentregistration/statusOfRegistration"
)

const (
	peRegistrationURL string = "https://eduapps.mit.edu/mitpe/student/registration/home"
)

type academicProfileInfo struct {
	Name  string `json:"name"`
	Year  string `json:"year"`
	MITID string `json:"mitid"`
}

type subjectSelectionInfo struct {
	SubjectID string `json:"subjectID"`
	SectionID string `json:"sectionID"`
	Title     string `json:"title"`
	Units     int    `json:"units"`
}

type subjectRegistrationInfo struct {
	Selection subjectSelectionInfo `json:"subjectSelectionInfo"`
}

type academicRegistrationInfo struct {
	TermCode         string                    `json:"termCode"`
	TermDescription  string                    `json:"termDescription"`
	Subjects         []subjectRegistrationInfo `json:"regSubjectSelectionInfo"`
	TotalUnits       string                    `json:"totalUnits"`
	RegistrationLoad string                    `json:"registrationLoad"`
}

type peRegistrationInfo struct {
	SectionID        string `json:"sectionID"`
	SectionNumber    int    `json:"sectionNumber"`
	CourseTitle      string `json:"courseTitle"`
	Quarter          string `json:"quarter"`
	QuarterShortCode string `json:"quarterShortCode"`
	PERegTermCode    string `json:"peRegTermCode"`
	LMODTermCode     string `json:"lmodTermCode"`
}

type registrationInfo struct {
	StatusOfRegistration academicRegistrationInfo `json:"statusOfRegistration"`
	PERegistrations      []peRegistrationInfo     `json:"peRegistrations"`
}

func fetchDataWithClient(tsClient *touchstone.Client, username string) (*academicProfileInfo, *registrationInfo, *peInfo, error) {
	// first: authenticate to the fancy beta undergrad dashboard, as it has a nice api
	// we have to load the main page first because the api endpoints won't redirect to touchstone
	_, err := tsClient.AuthenticateToResource(studentDashboardURL)
	if err != nil {
		return nil, nil, nil, err
	}

	// academic profile - name, id, class year
	academicProfileResp, err := tsClient.AuthenticateToResource(studentDashboardAcademicProfileURL)
	if err != nil {
		return nil, nil, nil, err
	}
	defer academicProfileResp.Body.Close()

	academicProfile := academicProfileInfo{}
	err = json.NewDecoder(academicProfileResp.Body).Decode(&academicProfile)
	if err != nil {
		return nil, nil, nil, err
	}

	// status of registration
	registrationResp, err := tsClient.AuthenticateToResource(studentDashboardRegistrationURL)
	if err != nil {
		return nil, nil, nil, err
	}
	defer registrationResp.Body.Close()

	registration := registrationInfo{}
	err = json.NewDecoder(registrationResp.Body).Decode(&registration)
	if err != nil {
		return nil, nil, nil, err
	}

	// we have to get the pe registration separately, as the student dashboard doesn't tell us times of classes
	peRegistrationResp, err := tsClient.AuthenticateToResource(peRegistrationURL)
	if err != nil {
		return nil, nil, nil, err
	}
	defer peRegistrationResp.Body.Close()

	peRegistrationDoc, err := goquery.NewDocumentFromReader(peRegistrationResp.Body)
	if err != nil {
		return nil, nil, nil, err
	}

	peInfo, err := parseDocForPEInfo(peRegistrationDoc)
	if err != nil {
		return nil, nil, nil, err
	}

	return &academicProfile, &registration, peInfo, nil
}

func parsePEDate(peDate string) (string, error) {
	parts := strings.Split(peDate, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("mit: scraper: couldn't parse pe date '%s'", peDate)
	}

	parsedDate, err := time.Parse("01/02/2006", parts[1])
	if err != nil {
		return "", fmt.Errorf("mit: scraper: couldn't parse pe date '%s'", peDate)
	}

	return parsedDate.Format("2006-01-02"), nil
}

func parsePESchedule(peSchedule string) ([]time.Weekday, int, int, string, error) {
	// return values are: ParsedDaysOfWeek, ParsedStartTime, ParsedEndTime, ParsedLocation, error
	// the input is in the format "Tue, Thu 2:00 PM at Location That Might Have Spaces"
	weekdayMap := map[string]time.Weekday{
		"Sun": time.Sunday,
		"Mon": time.Monday,
		"Tue": time.Tuesday,
		"Wed": time.Wednesday,
		"Thu": time.Thursday,
		"Fri": time.Friday,
		"Sat": time.Saturday,
	}

	atParts := strings.Split(peSchedule, " at ")
	if len(atParts) != 2 {
		return nil, 0, 0, "", fmt.Errorf("mit: scraper: wrong number of atParts, %d", len(atParts))
	}

	// assume the second half of the string is the location
	location := atParts[1]

	// parse the weekdays
	weekdays := []time.Weekday{}
	weekdayStrings := strings.Split(atParts[0], ", ")
	for _, weekdayString := range weekdayStrings {
		cleanWeekdayString := strings.Split(weekdayString, " ")[0]
		weekday, matched := weekdayMap[cleanWeekdayString]
		if !matched {
			return nil, 0, 0, "", fmt.Errorf("mit: scraper: unknown day of week '%s'", weekdayString)
		}

		weekdays = append(weekdays, weekday)
	}

	// parse the start time
	lastWeekdayParts := strings.SplitN(weekdayStrings[len(weekdayStrings)-1], " ", 2)
	if len(lastWeekdayParts) != 2 {
		return nil, 0, 0, "", fmt.Errorf("mit: scraper: wrong number of lastWeekdayParts, %d", len(lastWeekdayParts))
	}

	startTimeString := lastWeekdayParts[1]
	startTime, err := time.Parse("3:04 PM", startTimeString)
	if len(lastWeekdayParts) != 2 {
		return nil, 0, 0, "", err
	}

	// i'm pretty sure all PE classes are an hour long? i think?
	endTime := startTime.Add(time.Hour)

	startTime = startTime.AddDate(1970, 0, 0)
	endTime = endTime.AddDate(1970, 0, 0)

	return weekdays, int(startTime.Unix()), int(endTime.Unix()), location, nil
}

func parseDocForPEInfo(doc *goquery.Document) (*peInfo, error) {
	sectionContainer := doc.Find(".sectionContainer")
	if sectionContainer.Length() == 0 {
		return nil, errors.New("mit: scraper: couldn't find .sectionContainer in PE doc")
	}

	registered := false
	peInfo := peInfo{}

	tableRows := sectionContainer.Find("tr")
	tableRows.Each(func(i int, s *goquery.Selection) {
		key := strings.TrimSpace(s.Find("th").Text())
		value := strings.TrimSpace(s.Find("td").Text())

		if key == "Status" {
			if value == "Registered" {
				registered = true
			}
		} else if key == "Section ID" {
			peInfo.SectionID = value
		} else if key == "Activity" {
			peInfo.Activity = value
		} else if key == "Course Title" {
			peInfo.CourseTitle = value
		} else if key == "Schedule" {
			peInfo.RawSchedule = value
		} else if key == "First Day of Class" {
			peInfo.RawFirstDay = value
		} else if key == "Last Day of Class" {
			peInfo.RawLastDay = value
		} else if key == "Calendar Notes" {
			peInfo.RawCalendarNotes = value
		}
	})

	if !registered {
		return nil, nil
	}

	var err error

	peInfo.ParsedFirstDay, err = parsePEDate(peInfo.RawFirstDay)
	if err != nil {
		return nil, err
	}

	firstDay, err := time.Parse("2006-01-02", peInfo.ParsedFirstDay)
	if err != nil {
		return nil, err
	}

	peInfo.ParsedLastDay, err = parsePEDate(peInfo.RawLastDay)
	if err != nil {
		return nil, err
	}

	peInfo.ParsedSkipDays = []string{}

	if peInfo.RawCalendarNotes != "" {
		// so far, the only formats I know for this is "no classes 11/11, 11/26, 11/27" and "No Classes 11/11, 22, 23, 24, 25"
		parsedNotes := false
		calendarNotes := strings.ToLower(peInfo.RawCalendarNotes)
		if strings.HasPrefix(calendarNotes, "no classes ") {
			noPrefixString := strings.Replace(calendarNotes, "no classes ", "", -1)
			parts := strings.Split(noPrefixString, ", ")

			lastMonth := 0
			for _, rawSkipDay := range parts {
				// these are in the format "11/26"
				skipDayParts := strings.Split(rawSkipDay, "/")
				if len(skipDayParts) > 2 {
					continue
				}

				month, date := 0, 0
				if len(skipDayParts) == 1 {
					// it's one number
					// this implies that we use the same month as before
					if lastMonth == 0 {
						// there is no last month???
						// not sure what's going on here!!!
						continue
					}

					month = lastMonth
					date, err = strconv.Atoi(skipDayParts[0])
					if err != nil {
						continue
					}
				} else if len(skipDayParts) == 2 {
					// it's two numbers
					// it's a date like 11/26
					month, err = strconv.Atoi(skipDayParts[0])
					if err != nil {
						continue
					}

					date, err = strconv.Atoi(skipDayParts[1])
					if err != nil {
						continue
					}
				}

				// the year not given to us and must be guessed
				year := firstDay.Year()

				// if the month is BEFORE the month that the class starts, assume it's next year
				if time.Month(month) < firstDay.Month() {
					year++
				}

				skipDay := time.Date(year, time.Month(month), date, 0, 0, 0, 0, time.UTC)

				peInfo.ParsedSkipDays = append(peInfo.ParsedSkipDays, skipDay.Format("2006-01-02"))

				parsedNotes = true
			}
		}

		if !parsedNotes {
			// oofie
			// report it back and just hope that there's nothing important...
			errorlog.LogError(
				"parsing PE calendar notes",
				fmt.Errorf(
					"mit: scraper: couldn't parse calendar notes '%s' for '%s'",
					peInfo.RawCalendarNotes,
					peInfo.SectionID,
				),
			)
		}
	}

	// now we have to parse the schedule field
	peInfo.ParsedDaysOfWeek, peInfo.ParsedStartTime, peInfo.ParsedEndTime, peInfo.ParsedLocation, err = parsePESchedule(peInfo.RawSchedule)
	if err != nil {
		return nil, err
	}

	return &peInfo, nil
}
