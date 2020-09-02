package dalton

import (
	"database/sql"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/auth"
	"github.com/MyHomeworkSpace/api-server/blackbaud"
	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

type calendarClass struct {
	ID        int    `json:"id"`
	TermID    int    `json:"termID"`
	OwnerID   int    `json:"ownerId"`
	SectionID int    `json:"sectionId"`
	Name      string `json:"name"`
	OwnerName string `json:"ownerName"`
	UserID    int    `json:"userId"`
}
type calendarPeriod struct {
	ID           int    `json:"id"`
	ClassID      int    `json:"classId"`
	DayNumber    int    `json:"dayNumber"`
	Block        string `json:"block"`
	BuildingName string `json:"buildingName"`
	RoomNumber   string `json:"roomNumber"`
	Start        int    `json:"start"`
	End          int    `json:"end"`
	UserID       int    `json:"userId"`
}

// don't use me please
type dumbTerm struct {
	ID     int    `json:"id"`
	TermID int    `json:"termID"`
	Name   string `json:"name"`
	UserID int    `json:"userId"`
}

func clearUserData(tx *sql.Tx, user *data.User) error {
	// clear away anything that is in the db
	_, err := tx.Exec("DELETE FROM dalton_terms WHERE userId = ?", user.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM dalton_classes WHERE userId = ?", user.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM dalton_periods WHERE userId = ?", user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *school) Enroll(tx *sql.Tx, user *data.User, params map[string]interface{}) (map[string]interface{}, error) {
	usernameRaw, ok := params["username"]
	passwordRaw, ok2 := params["password"]

	if !ok || !ok2 {
		return nil, data.SchoolError{Code: "missing_params"}
	}

	username, ok := usernameRaw.(string)
	password, ok2 := passwordRaw.(string)

	if !ok || !ok2 || username == "" || password == "" {
		return nil, data.SchoolError{Code: "invalid_params"}
	}

	// test the credentials first so we don't run into blackbaud's rate limiting
	_, resp, ajaxToken, jar, err := auth.DaltonLogin(username, password)
	if resp != "" || err != nil {
		return nil, data.SchoolError{Code: resp}
	}

	schoolSlug := "dalton"

	// get user id
	response, err := blackbaud.Request(schoolSlug, "GET", "webapp/context", url.Values{}, map[string]interface{}{}, jar, ajaxToken, "")
	if err != nil {
		return nil, err
	}

	bbUserInfo := (response.(map[string]interface{}))["UserInfo"].(map[string]interface{})

	bbUserID := int(bbUserInfo["UserId"].(float64))
	bbUserDisplayName := bbUserInfo["StudentDisplay"].(string)
	bbUserUsername := strings.Replace(bbUserInfo["UserName"].(string), "@dalton.org", "", -1)

	foundHouseGroup := false
	houseSectionID := 0
	houseInfo := calendarClass{}
	allGroups := response.(map[string]interface{})["Groups"].([]interface{})
	for _, group := range allGroups {
		// look for the house group
		groupInfo := group.(map[string]interface{})
		groupName := groupInfo["GroupName"].(string)

		if strings.Contains(strings.ToLower(groupName), "house") {
			// found it!
			foundHouseGroup = true
			houseSectionID = int(groupInfo["SectionId"].(float64))
			houseInfo = calendarClass{
				-1,
				-1,
				-1,
				houseSectionID,
				groupInfo["GroupName"].(string),
				groupInfo["OwnerName"].(string),
				-1,
			}
			break
		}
	}

	// get list of grades
	response, err = blackbaud.Request(schoolSlug, "GET", "datadirect/StudentGradeLevelList", url.Values{}, map[string]interface{}{}, jar, ajaxToken, "")
	if err != nil {
		return nil, err
	}

	// find current grade
	gradeList := response.([]interface{})
	schoolYearLabel := ""
	for _, grade := range gradeList {
		gradeInfo := grade.(map[string]interface{})
		current := gradeInfo["CurrentInd"].(bool)
		if current {
			schoolYearLabel = gradeInfo["SchoolYearLabel"].(string)
		}
	}

	if schoolYearLabel == "" {
		return nil, data.SchoolError{Code: "bb_no_grade"}
	}

	// get list of terms
	response, err = blackbaud.Request(schoolSlug, "GET", "DataDirect/StudentGroupTermList", url.Values{
		"studentUserId":   {strconv.Itoa(bbUserID)},
		"schoolYearLabel": {schoolYearLabel},
		"personaId":       {"2"},
	}, map[string]interface{}{}, jar, ajaxToken, "")
	if err != nil {
		return nil, err
	}

	totalTermList := response.([]interface{})
	termMap := map[int]dumbTerm{}
	termRequestString := ""
	for _, term := range totalTermList {
		termInfo := term.(map[string]interface{})
		termID := int(termInfo["DurationId"].(float64))
		if termInfo["OfferingType"].(float64) == 1 {
			termMap[termID] = dumbTerm{
				-1,
				termID,
				termInfo["DurationDescription"].(string),
				-1,
			}
			if termRequestString != "" {
				termRequestString += ","
			}
			termRequestString += strconv.Itoa(termID)
		}
	}

	// get list of classes
	response, err = blackbaud.Request(schoolSlug, "GET", "datadirect/ParentStudentUserAcademicGroupsGet", url.Values{
		"userId":          {strconv.Itoa(bbUserID)},
		"schoolYearLabel": {schoolYearLabel},
		"memberLevel":     {"3"},
		"persona":         {"2"},
		"durationList":    {termRequestString},
		"markingPeriodId": {""},
	}, map[string]interface{}{}, jar, ajaxToken, "")
	if err != nil {
		return nil, err
	}

	totalClassList := response.([]interface{})
	classMap := map[int]calendarClass{}
	for _, class := range totalClassList {
		classInfo := class.(map[string]interface{})
		classID := int(classInfo["sectionid"].(float64))
		classItem := calendarClass{
			-1,
			int(classInfo["DurationId"].(float64)),
			int(classInfo["OwnerId"].(float64)),
			classID,
			classInfo["sectionidentifier"].(string),
			classInfo["groupownername"].(string),
			-1,
		}
		classMap[classID] = classItem
	}
	if foundHouseGroup {
		classMap[houseSectionID] = houseInfo
	}

	// find all periods of classes
	dayMap := map[int]map[int][]calendarPeriod{}
	for termIndex, term := range ImportTerms {
		dayMap[termIndex+1] = map[int][]calendarPeriod{
			0: {},
			1: {},
			2: {},
			3: {},
			4: {},
			5: {},
			6: {},
			7: {},
		}

		response, err = blackbaud.Request(schoolSlug, "GET", "DataDirect/ScheduleList", url.Values{
			"format":          {"json"},
			"viewerId":        {strconv.Itoa(bbUserID)},
			"personaId":       {"2"},
			"viewerPersonaId": {"2"},
			"start":           {strconv.FormatInt(term.Start.Unix(), 10)},
			"end":             {strconv.FormatInt(term.End.Unix(), 10)},
		}, map[string]interface{}{}, jar, ajaxToken, "")
		if err != nil {
			return nil, err
		}

		totalPeriodList := response.([]interface{})
		daysFound := map[int]string{
			0: "",
			1: "",
			2: "",
			3: "",
			4: "",
			5: "",
			6: "",
			7: "",
		}
		daysInfo := map[string]string{}
		for _, period := range totalPeriodList {
			periodInfo := period.(map[string]interface{})
			dayStr := strings.Split(periodInfo["start"].(string), " ")[0]

			if periodInfo["allDay"].(bool) {
				daysInfo[dayStr] = periodInfo["title"].(string)
				continue
			}

			day, err := time.Parse("1/2/2006", dayStr)
			if err != nil {
				return nil, err
			}

			dayNumber := int(day.Weekday())
			// if dayNumber == int(time.Friday) {
			// 	// find what friday it is and add that to the day number
			// 	info, ok := daysInfo[dayStr]
			// 	if !ok {
			// 		return nil, err
			// 	}
			// 	fridayNumber, err := strconv.Atoi(strings.Split(info, " ")[1])
			// 	if err == nil {
			// 		// we actually have a friday number, adjust it
			// 		dayNumber += fridayNumber - 1
			// 	}
			// }

			if daysFound[dayNumber] != "" && daysFound[dayNumber] != dayStr {
				// we've already found a source for periods from this day, and it's not this one
				// so just skip the period
				continue
			}

			daysFound[dayNumber] = dayStr

			startTime, err := time.Parse("3:04 PM", strings.SplitN(periodInfo["start"].(string), " ", 2)[1])
			endTime, err2 := time.Parse("3:04 PM", strings.SplitN(periodInfo["end"].(string), " ", 2)[1])

			if err != nil {
				return nil, err
			}
			if err2 != nil {
				return nil, err2
			}

			startTime = startTime.AddDate(1970, 0, 0)
			endTime = endTime.AddDate(1970, 0, 0)

			// add the period to our list
			periodItem := calendarPeriod{
				-1,
				int(periodInfo["SectionId"].(float64)),
				dayNumber,
				"",
				"",
				"",
				int(startTime.Unix()),
				int(endTime.Unix()),
				-1,
			}

			dayMap[termIndex+1][dayNumber] = append(dayMap[termIndex+1][dayNumber], periodItem)
		}
	}

	// find locations of classes
	for termIndex, term := range ImportTerms {
		termNum := termIndex + 1
		termDates := []time.Time{
			term.Start,
			term.Start.Add(1 * 24 * time.Hour),
			term.Start.Add(2 * 24 * time.Hour),
			term.Start.Add(3 * 24 * time.Hour),
			term.Start.Add(time.Duration(term.DayOffsets[0]) * 24 * time.Hour),
			term.Start.Add(time.Duration(term.DayOffsets[1]) * 24 * time.Hour),
			term.Start.Add(time.Duration(term.DayOffsets[2]) * 24 * time.Hour),
			term.Start.Add(time.Duration(term.DayOffsets[3]) * 24 * time.Hour),
		}

		for dayNumber, date := range termDates {
			periods := dayMap[termNum][dayNumber+1]

			scheduleInfo, err := blackbaud.Request(schoolSlug, "GET", "schedule/MyDayCalendarStudentList", url.Values{
				"scheduleDate": {date.Format("1/2/2006")},
				"personaId":    {"2"},
			}, map[string]interface{}{}, jar, ajaxToken, "")
			if err != nil {
				return nil, err
			}

			scheduleList := scheduleInfo.([]interface{})

			for periodIndex, period := range periods {
				for _, scheduleInterface := range scheduleList {
					scheduleItem := scheduleInterface.(map[string]interface{})
					sectionID := int(scheduleItem["SectionId"].(float64))
					if sectionID == period.ClassID {
						block := scheduleItem["Block"].(string)
						buildingName := scheduleItem["BuildingName"].(string)
						roomNumber := scheduleItem["RoomNumber"].(string)

						dayMap[termNum][dayNumber+1][periodIndex].Block = block
						dayMap[termNum][dayNumber+1][periodIndex].BuildingName = buildingName
						dayMap[termNum][dayNumber+1][periodIndex].RoomNumber = roomNumber
					}
				}
			}
		}
	}

	// add all of this to mysql

	// wipe whatever's there
	err = clearUserData(tx, user)
	if err != nil {
		return nil, err
	}

	// first add the terms
	termInsertStmt, err := tx.Prepare("INSERT INTO dalton_terms(termID, name, userId) VALUES(?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer termInsertStmt.Close()
	for _, term := range termMap {
		_, err = termInsertStmt.Exec(term.TermID, term.Name, user.ID)
		if err != nil {
			return nil, err
		}
	}

	// then the classes
	classInsertStmt, err := tx.Prepare("INSERT INTO dalton_classes(termID, ownerId, sectionId, name, ownerName, userId) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer classInsertStmt.Close()
	for _, class := range classMap {
		_, err = classInsertStmt.Exec(class.TermID, class.OwnerID, class.SectionID, class.Name, class.OwnerName, user.ID)
		if err != nil {
			return nil, err
		}
	}

	// and finally the periods
	periodsInsertStmt, err := tx.Prepare("INSERT INTO dalton_periods(classId, dayNumber, block, buildingName, roomNumber, start, end, userId) VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	defer periodsInsertStmt.Close()
	for termNum, term := range dayMap {
		for _, periods := range term {
			for _, period := range periods {
				if period.ClassID == houseSectionID && termNum != 1 {
					// skip inserting it again because house doesn't change ID
					continue
				}
				_, err = periodsInsertStmt.Exec(period.ClassID, period.DayNumber, period.Block, period.BuildingName, period.RoomNumber, period.Start, period.End, user.ID)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return map[string]interface{}{
		"status":   1,
		"name":     bbUserDisplayName,
		"username": bbUserUsername,
	}, nil
}

func (s *school) Unenroll(tx *sql.Tx, user *data.User) error {
	return clearUserData(tx, user)
}

func (s *school) NeedsUpdate(db *sql.DB) (bool, error) {
	return (s.importStatus != schools.ImportStatusOK), nil
}
