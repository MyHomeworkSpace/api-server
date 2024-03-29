package cornell

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
	"github.com/MyHomeworkSpace/api-server/util"
	"github.com/PuerkitoBio/goquery"
)

var ErrUnexpectedPageStructure = errors.New("cornell: Unexpected page structure")

var apiKeyRegexp = regexp.MustCompile(`"key":"(?P<Key>[^"]*)`)
var nameRegexp = regexp.MustCompile(`login.*"name":"([^"]*)`) //kinda hacky, I should fix this

func (s *school) Enroll(tx *sql.Tx, user *data.User, params map[string]interface{}) (map[string]interface{}, error) {
	netidRaw, ok := params["username"]
	passwordRaw, ok2 := params["password"]

	if !ok || !ok2 {
		return nil, data.SchoolError{Code: "missing_params"}
	}

	netID, ok := netidRaw.(string)
	password, ok2 := passwordRaw.(string)

	if !ok || !ok2 || netID == "" || password == "" {
		return nil, data.SchoolError{Code: "invalid_params"}
	}

	cookieJar, _ := cookiejar.New(nil)
	c := &http.Client{
		Jar: cookieJar,
	}

	term := GetCurrentTerm().Code

	resp, err := c.Get("https://classes.cornell.edu/sascuwalogin/login/redirect?redirectUri=https%3A%2F%2Fclasses.cornell.edu%2Fscheduler%2Froster%2F" + term)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	loginForm, exists := doc.Find("form").First().Attr("action")
	if !exists {
		return nil, ErrUnexpectedPageStructure
	}

	values := url.Values{
		"j_username":       {netID},
		"j_password":       {password},
		"_eventId_proceed": {"Login"},
	}

	loginResp, err := c.PostForm(resp.Request.URL.Scheme+"://"+resp.Request.URL.Host+loginForm, values)
	if err != nil {
		return nil, err
	}

	confirmationDoc, err := goquery.NewDocumentFromReader(loginResp.Body)
	confirmationDocString, _ := confirmationDoc.Html()

	if strings.Contains(confirmationDocString, "Unable to log in") {
		return nil, data.SchoolError{Code: "creds_incorrect"}
	}

	// otherwise we're logged in! Unfortunately, CUWebLogin has this intermediate screen that says Your login credentials are being transmitted to the website via POST. This only works if you use javascript, so we need to manually transmit the token

	classesURL, exists := confirmationDoc.Find("form").First().Attr("action")
	if !exists {
		return nil, ErrUnexpectedPageStructure
	}

	relayState, exists := confirmationDoc.Find("input[name=RelayState]").First().Attr("value")
	if !exists {
		return nil, ErrUnexpectedPageStructure
	}

	samlResponse, exists := confirmationDoc.Find("input[name=SAMLResponse]").First().Attr("value")
	if !exists {
		return nil, ErrUnexpectedPageStructure
	}

	schedulerResp, err := c.PostForm(classesURL, url.Values{"RelayState": {relayState}, "SAMLResponse": {samlResponse}})
	if err != nil {
		return nil, err
	}

	schedulerRespBuf := new(bytes.Buffer)
	schedulerRespBuf.ReadFrom(schedulerResp.Body)
	schedulerRespContent := schedulerRespBuf.String()

	APIKey := apiKeyRegexp.FindStringSubmatch(schedulerRespContent)[1]
	name := nameRegexp.FindStringSubmatch(schedulerRespContent)[1]

	scheduleReq, err := http.NewRequest("GET", "https://classes.cornell.edu/api/3.0/scheduler/current-enrollment", nil)
	if err != nil {
		return nil, err
	}

	scheduleReq.Header.Add("Authorization", "ClassRoster "+APIKey)

	scheduleResp, err := c.Do(scheduleReq)
	if err != nil {
		return nil, err
	}

	scheduleRespBuf := new(bytes.Buffer)
	scheduleRespBuf.ReadFrom(scheduleResp.Body)
	classesBytes := scheduleRespBuf.Bytes()

	courses := []classItem{}

	err = json.Unmarshal(classesBytes, &courses)
	if err != nil {
		return nil, err
	}

	cpairs := coursePairs{}

	for _, class := range courses {
		cpairs.CoursePairs = append(cpairs.CoursePairs, strconv.Itoa(class.CourseID)+","+strconv.Itoa(class.CourseOfferNumber))
	}

	coursePairsJSON, err := json.Marshal(cpairs)

	if err != nil {
		return nil, err
	}

	courseDetailReq, err := http.NewRequest("POST", "https://classes.cornell.edu/api/3.0/scheduler/course-detail", bytes.NewBuffer(coursePairsJSON))
	if err != nil {
		return nil, err
	}
	courseDetailReq.Header.Set("Content-Type", "application/json")
	courseDetailReq.Header.Add("Authorization", "ClassRoster "+APIKey)

	courseDetailsResp, err := c.Do(courseDetailReq)
	if err != nil {
		return nil, err
	}

	courseDetailBuf := new(bytes.Buffer)
	courseDetailBuf.ReadFrom(courseDetailsResp.Body)

	courseDetails := []course{}
	err = json.Unmarshal(courseDetailBuf.Bytes(), &courseDetails)
	if err != nil {
		return nil, err
	}

	// ok, now we have all the course details and the classes that the student is taking. Now we need to put it all in the db.

	for _, course := range courses {
		_, err = tx.Exec("INSERT INTO cornell_courses (userId, subject, catalogNum, title, units, rosterId) VALUES (?, ?, ?, ?, ?, ?)", user.ID, course.Subject, course.CatalogNumber, course.Title, course.Units, course.CourseID)
		if err != nil {
			return nil, err
		}

		for _, details := range courseDetails {
			if details.CourseID == course.CourseID {
				// we matched the details with the course!
				for _, eg := range details.EnrollGroups {
					for _, section := range eg.ClassSections {
						if util.IntSliceContains(course.ClassNumbers, section.ClassNum) {
							for _, meeting := range section.Meetings {
								if meeting.StartTime == "" || meeting.EndTime == "" {
									// this class is missing a time
									// not super clear what these are, skip it
									continue
								}

								startDate, err := time.Parse("01/02/2006", meeting.StartDate)
								if err != nil {
									return nil, err
								}

								endDate, err := time.Parse("01/02/2006", meeting.EndDate)
								if err != nil {
									return nil, err
								}

								startTime, err := time.Parse("3:04PM", meeting.StartTime)
								if err != nil {
									return nil, err
								}

								endTime, err := time.Parse("3:04PM", meeting.EndTime)
								if err != nil {
									return nil, err
								}

								// kinda hacky but it works
								startTime = time.Date(1970, 01, 01, startTime.Hour(), startTime.Minute(), 0, 0, time.UTC)
								endTime = time.Date(1970, 01, 01, endTime.Hour(), endTime.Minute(), 0, 0, time.UTC)

								satMeeting := false
								sunMeeting := false

								if strings.Contains("SSu", meeting.Pattern) {
									satMeeting = true
									sunMeeting = true
								} else if strings.Contains("Su", meeting.Pattern) {
									sunMeeting = true
								} else if strings.Contains("S", meeting.Pattern) {
									satMeeting = true
								}

								_, err = tx.Exec(`INSERT INTO cornell_events (
		title,
		userId,
		subject,
		catalogNum,
		classNum,
		component,
		componentLong,
		section,
		campus,
		campusLong,
		location,
		locationLong,
		startDate,
		endDate,
		startTime,
		endTime,
		monday,
		tuesday,
		wednesday,
		thursday,
		friday,
		saturday,
		sunday,
		facility,
		facilityLong,
		building
	)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
									course.Title,
									user.ID,
									course.Subject,
									course.CatalogNumber,
									section.ClassNum,
									section.Component,
									section.ComponentLong,
									section.Section,
									section.Campus,
									section.CampusDesc,
									section.Location,
									section.LocationDesc,
									startDate.Format("2006-01-02"),
									endDate.Format("2006-01-02"),
									startTime.Unix(),
									endTime.Unix(),
									strings.Contains(meeting.Pattern, "M"),
									strings.Contains(meeting.Pattern, "T"),
									strings.Contains(meeting.Pattern, "W"),
									strings.Contains(meeting.Pattern, "R"),
									strings.Contains(meeting.Pattern, "F"),
									satMeeting,
									sunMeeting,
									meeting.FacilityDescShort,
									meeting.FacilityDesc,
									meeting.BuildingDesc,
								)

								if err != nil {
									return nil, err
								}
							}
						}
					}
				}
				break
			}
		}
	}

	return map[string]interface{}{
		"netid":  strings.ToLower(netID),
		"name":   name,
		"status": schools.ImportStatusOK,
	}, nil
}

func (s *school) NeedsUpdate(db *sql.DB) (bool, error) {
	return (s.importStatus != schools.ImportStatusOK), nil
}

func (s *school) Unenroll(tx *sql.Tx, user *data.User) error {
	return clearUserData(tx, user)
}

func clearUserData(tx *sql.Tx, user *data.User) error {
	// clear away anything that is in the db
	_, err := tx.Exec("DELETE FROM cornell_courses WHERE userId = ?", user.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM cornell_events WHERE userId = ?", user.ID)
	if err != nil {
		return err
	}

	return nil
}
