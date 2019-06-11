package api

import (
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/auth"
	"github.com/MyHomeworkSpace/api-server/blackbaud"
	"github.com/MyHomeworkSpace/api-server/calendar"

	"github.com/labstack/echo"
)

// structs for data
type CalendarClass struct {
	ID        int    `json:"id"`
	TermID    int    `json:"termId"`
	OwnerID   int    `json:"ownerId"`
	SectionID int    `json:"sectionId"`
	Name      string `json:"name"`
	OwnerName string `json:"ownerName"`
	UserID    int    `json:"userId"`
}
type CalendarPeriod struct {
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
type DumbTerm struct {
	ID     int    `json:"id"`
	TermID int    `json:"termId"`
	Name   string `json:"name"`
	UserID int    `json:"userId"`
}

// responses
type CalendarStatusResponse struct {
	Status    string `json:"status"`
	StatusNum int    `json:"statusNum"`
}

func routeCalendarGetStatus(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT status FROM calendar_status WHERE userId = ?", c.User.ID)
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		ec.JSON(http.StatusOK, CalendarStatusResponse{"ok", 0})
		return
	}

	statusNum := -1
	rows.Scan(&statusNum)

	ec.JSON(http.StatusOK, CalendarStatusResponse{"ok", statusNum})
}

func routeCalendarGetView(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("start") == "" || ec.FormValue("end") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	timeZone, err := time.LoadLocation("America/New_York")
	if err != nil {
		ErrorLog_LogError("timezone info", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	startDate, err := time.ParseInLocation("2006-01-02", ec.FormValue("start"), timeZone)
	if err != nil {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}
	endDate, err := time.ParseInLocation("2006-01-02", ec.FormValue("end"), timeZone)
	if err != nil {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}

	if int(math.Floor(endDate.Sub(startDate).Hours()/24)) > 2*365 {
		// cap of 2 years between start and end
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}

	view, err := calendar.GetView(DB, c.User, timeZone, startDate, endDate)
	if err != nil {
		ErrorLog_LogError("getting calendar view", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, CalendarViewResponse{"ok", view})
}

func routeCalendarImport(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("password") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	// test the credentials first so we don't run into blackbaud's rate limiting
	_, resp, err := auth.DaltonLogin(c.User.Username, ec.FormValue("password"))
	if resp != "" || err != nil {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", resp})
		return
	}

	schoolSlug := "dalton"

	// set up ajax token and stuff
	ajaxToken, err := blackbaud.GetAjaxToken(schoolSlug)
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "ajaxtoken_error"})
		return
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// sign in to blackbaud
	response, err := blackbaud.Request(schoolSlug, "POST", "SignIn", url.Values{}, map[string]interface{}{
		"From":            "",
		"InterfaceSource": "WebApp",
		"Password":        ec.FormValue("password"),
		"Username":        c.User.Username,
		"remember":        "false",
	}, jar, ajaxToken)

	if err != nil {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "bb_signin_error"})
		return
	}

	result, worked := (response.(map[string]interface{}))["AuthenticationResult"].(float64)

	if worked && result == 5 {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "bb_signin_rate_limit"})
		return
	}

	if !worked || result == 2 {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "bb_signin_error"})
		return
	}

	// get user id
	response, err = blackbaud.Request(schoolSlug, "GET", "webapp/context", url.Values{}, map[string]interface{}{}, jar, ajaxToken)
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	bbUserId := int(((response.(map[string]interface{}))["UserInfo"].(map[string]interface{}))["UserId"].(float64))

	foundHouseGroup := false
	houseSectionId := 0
	houseInfo := CalendarClass{}
	allGroups := response.(map[string]interface{})["Groups"].([]interface{})
	for _, group := range allGroups {
		// look for the house group
		groupInfo := group.(map[string]interface{})
		groupName := groupInfo["GroupName"].(string)

		if strings.Contains(strings.ToLower(groupName), "house") {
			// found it!
			foundHouseGroup = true
			houseSectionId = int(groupInfo["SectionId"].(float64))
			houseInfo = CalendarClass{
				-1,
				-1,
				-1,
				houseSectionId,
				groupInfo["GroupName"].(string),
				groupInfo["OwnerName"].(string),
				-1,
			}
			break
		}
	}

	// get list of grades
	response, err = blackbaud.Request(schoolSlug, "GET", "datadirect/StudentGradeLevelList", url.Values{}, map[string]interface{}{}, jar, ajaxToken)
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
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
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "bb_no_grade"})
		return
	}

	// get list of terms
	response, err = blackbaud.Request(schoolSlug, "GET", "DataDirect/StudentGroupTermList", url.Values{
		"studentUserId":   {strconv.Itoa(bbUserId)},
		"schoolYearLabel": {schoolYearLabel},
		"personaId":       {"2"},
	}, map[string]interface{}{}, jar, ajaxToken)
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	totalTermList := response.([]interface{})
	termMap := map[int]DumbTerm{}
	termRequestString := ""
	for _, term := range totalTermList {
		termInfo := term.(map[string]interface{})
		termId := int(termInfo["DurationId"].(float64))
		if termInfo["OfferingType"].(float64) == 1 {
			termMap[termId] = DumbTerm{
				-1,
				termId,
				termInfo["DurationDescription"].(string),
				-1,
			}
			if termRequestString != "" {
				termRequestString += ","
			}
			termRequestString += strconv.Itoa(termId)
		}
	}

	// get list of classes
	response, err = blackbaud.Request(schoolSlug, "GET", "datadirect/ParentStudentUserAcademicGroupsGet", url.Values{
		"userId":          {strconv.Itoa(bbUserId)},
		"schoolYearLabel": {schoolYearLabel},
		"memberLevel":     {"3"},
		"persona":         {"2"},
		"durationList":    {termRequestString},
		"markingPeriodId": {""},
	}, map[string]interface{}{}, jar, ajaxToken)
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	totalClassList := response.([]interface{})
	classMap := map[int]CalendarClass{}
	for _, class := range totalClassList {
		classInfo := class.(map[string]interface{})
		classId := int(classInfo["sectionid"].(float64))
		classItem := CalendarClass{
			-1,
			int(classInfo["DurationId"].(float64)),
			int(classInfo["OwnerId"].(float64)),
			classId,
			classInfo["sectionidentifier"].(string),
			classInfo["groupownername"].(string),
			-1,
		}
		classMap[classId] = classItem
	}
	if foundHouseGroup {
		classMap[houseSectionId] = houseInfo
	}

	// find all periods of classes
	dayMap := map[int]map[int][]CalendarPeriod{}
	for _, term := range []int{1, 2} {
		dayMap[term] = map[int][]CalendarPeriod{
			0: []CalendarPeriod{},
			1: []CalendarPeriod{},
			2: []CalendarPeriod{},
			3: []CalendarPeriod{},
			4: []CalendarPeriod{},
			5: []CalendarPeriod{},
			6: []CalendarPeriod{},
			7: []CalendarPeriod{},
		}

		startDate := calendar.Term1_Import_Start
		endDate := calendar.Term1_Import_End
		if term == 2 {
			startDate = calendar.Term2_Import_Start
			endDate = calendar.Term2_Import_End
		}

		response, err = blackbaud.Request(schoolSlug, "GET", "DataDirect/ScheduleList", url.Values{
			"format":          {"json"},
			"viewerId":        {strconv.Itoa(bbUserId)},
			"personaId":       {"2"},
			"viewerPersonaId": {"2"},
			"start":           {strconv.FormatInt(startDate.Unix(), 10)},
			"end":             {strconv.FormatInt(endDate.Unix(), 10)},
		}, map[string]interface{}{}, jar, ajaxToken)
		if err != nil {
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
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
				ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				return
			}

			dayNumber := int(day.Weekday())
			if dayNumber == int(time.Friday) {
				// find what friday it is and add that to the day number
				info, ok := daysInfo[dayStr]
				if !ok {
					ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
					return
				}
				fridayNumber, err := strconv.Atoi(strings.Split(info, " ")[1])
				if err != nil {
					ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
					return
				}
				dayNumber += fridayNumber - 1
			}

			if daysFound[dayNumber] != "" && daysFound[dayNumber] != dayStr {
				// we've already found a source for periods from this day, and it's not this one
				// so just skip the period
				continue
			}

			daysFound[dayNumber] = dayStr

			startTime, err := time.Parse("3:04 PM", strings.SplitN(periodInfo["start"].(string), " ", 2)[1])
			endTime, err2 := time.Parse("3:04 PM", strings.SplitN(periodInfo["end"].(string), " ", 2)[1])

			if err != nil || err2 != nil {
				ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				return
			}

			startTime = startTime.AddDate(1970, 0, 0)
			endTime = endTime.AddDate(1970, 0, 0)

			// add the period to our list
			periodItem := CalendarPeriod{
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

			dayMap[term][dayNumber] = append(dayMap[term][dayNumber], periodItem)
		}
	}

	// find locations of classes
	datesToSearch := map[int][]time.Time{
		1: {
			calendar.Term1_Import_Start,
			calendar.Term1_Import_Start.Add(1 * 24 * time.Hour),
			calendar.Term1_Import_Start.Add(2 * 24 * time.Hour),
			calendar.Term1_Import_Start.Add(3 * 24 * time.Hour),
			calendar.Term1_Import_Start.Add(time.Duration(calendar.Term1_Import_DayOffset_Friday1) * 24 * time.Hour),
			calendar.Term1_Import_Start.Add(time.Duration(calendar.Term1_Import_DayOffset_Friday2) * 24 * time.Hour),
			calendar.Term1_Import_Start.Add(time.Duration(calendar.Term1_Import_DayOffset_Friday3) * 24 * time.Hour),
			calendar.Term1_Import_Start.Add(time.Duration(calendar.Term1_Import_DayOffset_Friday4) * 24 * time.Hour),
		},
		2: {
			calendar.Term2_Import_Start,
			calendar.Term2_Import_Start.Add(1 * 24 * time.Hour),
			calendar.Term2_Import_Start.Add(2 * 24 * time.Hour),
			calendar.Term2_Import_Start.Add(3 * 24 * time.Hour),
			calendar.Term2_Import_Start.Add(time.Duration(calendar.Term2_Import_DayOffset_Friday1) * 24 * time.Hour),
			calendar.Term2_Import_Start.Add(time.Duration(calendar.Term2_Import_DayOffset_Friday2) * 24 * time.Hour),
			calendar.Term2_Import_Start.Add(time.Duration(calendar.Term2_Import_DayOffset_Friday3) * 24 * time.Hour),
			calendar.Term2_Import_Start.Add(time.Duration(calendar.Term2_Import_DayOffset_Friday4) * 24 * time.Hour),
		},
	}

	for termNum, termDates := range datesToSearch {
		for dayNumber, date := range termDates {
			periods := dayMap[termNum][dayNumber+1]

			scheduleInfo, err := blackbaud.Request(schoolSlug, "GET", "schedule/MyDayCalendarStudentList", url.Values{
				"scheduleDate": {date.Format("1/2/2006")},
				"personaId":    {"2"},
			}, map[string]interface{}{}, jar, ajaxToken)
			if err != nil {
				ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				return
			}

			scheduleList := scheduleInfo.([]interface{})

			for periodIndex, period := range periods {
				for _, scheduleInterface := range scheduleList {
					scheduleItem := scheduleInterface.(map[string]interface{})
					sectionId := int(scheduleItem["SectionId"].(float64))
					if sectionId == period.ClassID {
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
	// in one giant transaction
	tx, err := DB.Begin()

	// clear away anything that is in the db
	_, err = tx.Exec("DELETE FROM calendar_terms WHERE userId = ?", c.User.ID)
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	_, err = tx.Exec("DELETE FROM calendar_classes WHERE userId = ?", c.User.ID)
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	_, err = tx.Exec("DELETE FROM calendar_periods WHERE userId = ?", c.User.ID)
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	_, err = tx.Exec("DELETE FROM calendar_status WHERE userId = ?", c.User.ID)
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// first add the terms
	termInsertStmt, err := tx.Prepare("INSERT INTO calendar_terms(termId, name, userId) VALUES(?, ?, ?)")
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer termInsertStmt.Close()
	for _, term := range termMap {
		_, err = termInsertStmt.Exec(term.TermID, term.Name, c.User.ID)
		if err != nil {
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}
	}

	// then the classes
	classInsertStmt, err := tx.Prepare("INSERT INTO calendar_classes(termId, ownerId, sectionId, name, ownerName, userId) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer classInsertStmt.Close()
	for _, class := range classMap {
		_, err = classInsertStmt.Exec(class.TermID, class.OwnerID, class.SectionID, class.Name, class.OwnerName, c.User.ID)
		if err != nil {
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}
	}

	// and finally the periods
	periodsInsertStmt, err := tx.Prepare("INSERT INTO calendar_periods(classId, dayNumber, block, buildingName, roomNumber, start, end, userId) VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer periodsInsertStmt.Close()
	for termNum, term := range dayMap {
		for _, periods := range term {
			for _, period := range periods {
				if period.ClassID == houseSectionId && termNum != 1 {
					// skip inserting it again because house doesn't change ID
					continue
				}
				_, err = periodsInsertStmt.Exec(period.ClassID, period.DayNumber, period.Block, period.BuildingName, period.RoomNumber, period.Start, period.End, c.User.ID)
				if err != nil {
					ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
					return
				}
			}
		}
	}

	_, err = tx.Exec("INSERT INTO calendar_status(status, userId) VALUES(1, ?)", c.User.ID)
	if err != nil {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// go!
	err = tx.Commit()
	if err != nil {
		ErrorLog_LogError("adding schedule to DB", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeCalendarResetSchedule(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	userID := c.User.ID

	if ec.FormValue("userID") != "" {
		// want to target a user, can they?
		if strings.Split(r.RemoteAddr, ":")[0] == "127.0.0.1" || strings.HasPrefix(r.RemoteAddr, "[::1]") {
			// yes
			var err error
			userID, err = strconv.Atoi(ec.FormValue("userID"))
			if err != nil {
				ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
				return
			}
		} else {
			// nope
			ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
			return
		}
	}

	tx, err := DB.Begin()
	if err != nil {
		ErrorLog_LogError("clearing schedule from DB", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// clear away anything that is in the db
	_, err = tx.Exec("DELETE FROM calendar_terms WHERE userId = ?", userID)
	if err != nil {
		ErrorLog_LogError("clearing schedule from DB", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	_, err = tx.Exec("DELETE FROM calendar_classes WHERE userId = ?", userID)
	if err != nil {
		ErrorLog_LogError("clearing schedule from DB", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	_, err = tx.Exec("DELETE FROM calendar_periods WHERE userId = ?", userID)
	if err != nil {
		ErrorLog_LogError("clearing schedule from DB", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	_, err = tx.Exec("DELETE FROM calendar_status WHERE userId = ?", userID)
	if err != nil {
		ErrorLog_LogError("clearing schedule from DB", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	err = tx.Commit()
	if err != nil {
		ErrorLog_LogError("adding schedule to DB", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}
