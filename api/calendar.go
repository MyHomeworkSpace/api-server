package api

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/auth"

	"github.com/labstack/echo"
)

// structs for data
type CalendarTerm struct {
	ID     int    `json:"id"`
	TermID int    `json:"termId"`
	Name   string `json:"name"`
	UserID int    `json:"userId"`
}
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
	ID        int `json:"id"`
	ClassID   int `json:"classId"`
	DayNumber int `json:"dayNumber"`
	Start     int `json:"start"`
	End       int `json:"end"`
	UserID    int `json:"userId"`
}
type CalendarScheduleItem struct {
	ID        int    `json:"id"`
	TermID    int    `json:"termId"`
	ClassID   int    `json:"classId"`
	Name      string `json:"name"`
	OwnerID   int    `json:"ownerId"`
	OwnerName string `json:"ownerName"`
	DayNumber int    `json:"dayNumber"`
	Start     int    `json:"start"`
	End       int    `json:"end"`
	UserID    int    `json:"userId"`
}

// responses
type CalendarScheduleResponse struct {
	Status string                 `json:"status"`
	Terms  []CalendarTerm         `json:"terms"`
	Items  []CalendarScheduleItem `json:"items"`
}
type CalendarStatusResponse struct {
	Status    string `json:"status"`
	StatusNum int    `json:"statusNum"`
}

func InitCalendarAPI(e *echo.Echo) {
	e.GET("/calendar/getSchedule", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		// get the terms
		terms := []CalendarTerm{}
		termRows, err := DB.Query("SELECT id, termId, name, userId FROM calendar_terms WHERE userId = ?", GetSessionUserID(&c))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer termRows.Close()
		for termRows.Next() {
			term := CalendarTerm{-1, -1, "", -1}
			termRows.Scan(&term.ID, &term.TermID, &term.Name, &term.UserID)
			terms = append(terms, term)
		}

		// get the periods for each term
		items := []CalendarScheduleItem{}
		rows, err := DB.Query("SELECT calendar_periods.id, calendar_classes.termId, calendar_classes.sectionId, calendar_classes.`name`, calendar_classes.ownerId, calendar_classes.ownerName, calendar_periods.dayNumber, calendar_periods.`start`, calendar_periods.`end`, calendar_periods.userId FROM calendar_periods INNER JOIN calendar_classes ON calendar_periods.classId = calendar_classes.sectionId WHERE calendar_periods.userId = ?", GetSessionUserID(&c))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		for rows.Next() {
			item := CalendarScheduleItem{-1, -1, -1, "", -1, "", -1, -1, -1, -1}
			rows.Scan(&item.ID, &item.TermID, &item.ClassID, &item.Name, &item.OwnerID, &item.OwnerName, &item.DayNumber, &item.Start, &item.End, &item.UserID)
			items = append(items, item)
		}

		return c.JSON(http.StatusOK, CalendarScheduleResponse{"ok", terms, items})
	})

	e.GET("/calendar/getStatus", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		rows, err := DB.Query("SELECT status FROM calendar_status WHERE userId = ?", GetSessionUserID(&c))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		if !rows.Next() {
			return c.JSON(http.StatusOK, CalendarStatusResponse{"ok", 0})
		}

		statusNum := -1
		rows.Scan(&statusNum)

		return c.JSON(http.StatusOK, CalendarStatusResponse{"ok", statusNum})
	})

	e.POST("/calendar/import", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("password") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		userId := GetSessionUserID(&c)

		// test the credentials first so we don't run into blackbaud's rate limiting
		_, resp, err := auth.DaltonLogin(GetSessionInfo(&c).Username, c.FormValue("password"))
		if resp != "" || err != nil {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", resp})
		}

		// set up ajax token and stuff
		ajaxToken, err := Blackbaud_GetAjaxToken()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "ajaxtoken_error"})
		}

		jar, err := cookiejar.New(nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		// sign in to blackbaud
		response, err := Blackbaud_Request("POST", "SignIn", url.Values{}, map[string]interface{}{
			"From":            "",
			"InterfaceSource": "WebApp",
			"Password":        c.FormValue("password"),
			"Username":        GetSessionInfo(&c).Username,
			"remember":        "false",
		}, jar, ajaxToken)

		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "bb_signin_error"})
		}

		result, worked := (response.(map[string]interface{}))["AuthenticationResult"].(float64)

		if worked && result == 5 {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "bb_signin_rate_limit"})
		}

		if !worked || result == 2 {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "bb_signin_error"})
		}

		// get user id
		response, err = Blackbaud_Request("GET", "webapp/context", url.Values{}, map[string]interface{}{}, jar, ajaxToken)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		bbUserId := int(((response.(map[string]interface{}))["UserInfo"].(map[string]interface{}))["UserId"].(float64))

		// get list of grades
		response, err = Blackbaud_Request("GET", "datadirect/StudentGradeLevelList", url.Values{}, map[string]interface{}{}, jar, ajaxToken)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
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
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "bb_no_grade"})
		}

		// get list of terms
		response, err = Blackbaud_Request("GET", "DataDirect/StudentGroupTermList", url.Values{
			"studentUserId":   {strconv.Itoa(bbUserId)},
			"schoolYearLabel": {schoolYearLabel},
			"personaId":       {"2"},
		}, map[string]interface{}{}, jar, ajaxToken)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		totalTermList := response.([]interface{})
		termMap := map[int]CalendarTerm{}
		termRequestString := ""
		for _, term := range totalTermList {
			termInfo := term.(map[string]interface{})
			termId := int(termInfo["DurationId"].(float64))
			if termInfo["OfferingType"].(float64) == 1 {
				termMap[termId] = CalendarTerm{
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
		response, err = Blackbaud_Request("GET", "datadirect/ParentStudentUserAcademicGroupsGet", url.Values{
			"userId":          {strconv.Itoa(bbUserId)},
			"schoolYearLabel": {schoolYearLabel},
			"memberLevel":     {"3"},
			"persona":         {"2"},
			"durationList":    {termRequestString},
			"markingPeriodId": {""},
		}, map[string]interface{}{}, jar, ajaxToken)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
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

			month := time.October
			year, err := strconv.Atoi(strings.Trim(strings.Split(schoolYearLabel, "-")[0], " "))
			if err != nil {
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}
			if term == 2 {
				month = time.February
				year += 1
			}

			startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
			endDate := startDate.AddDate(0, 1, -1)

			response, err = Blackbaud_Request("GET", "DataDirect/ScheduleList", url.Values{
				"format":          {"json"},
				"viewerId":        {strconv.Itoa(bbUserId)},
				"personaId":       {"2"},
				"viewerPersonaId": {"2"},
				"start":           {strconv.FormatInt(startDate.Unix(), 10)},
				"end":             {strconv.FormatInt(endDate.Unix(), 10)},
			}, map[string]interface{}{}, jar, ajaxToken)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
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
					return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				}

				dayNumber := int(day.Weekday())
				if dayNumber == int(time.Friday) {
					// find what friday it is and add that to the day number
					info, ok := daysInfo[dayStr]
					if !ok {
						return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
					}
					fridayNumber, err := strconv.Atoi(strings.Split(info, " ")[1])
					if err != nil {
						return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
					}
					dayNumber += fridayNumber - 1
				}

				if daysFound[dayNumber] != "" && daysFound[dayNumber] != dayStr {
					// we've already found a source for periods from this day, and it's not this one
					// so just skip the period
					continue
				}

				daysFound[dayNumber] = dayStr

				//map[start:2/24/2017 2:30 PM endTicks:6.36235461e+17 AssociationId:1 LinkableCoursePage:false startTicks:6.36235434e+17 end:2/24/2017 3:15 PM SectionId:3.464249e+06 title:Introduction to Drama - 3211-05 (E) allDay:false]

				startTime, err := time.Parse("3:04 PM", strings.SplitN(periodInfo["start"].(string), " ", 2)[1])
				endTime, err2 := time.Parse("3:04 PM", strings.SplitN(periodInfo["end"].(string), " ", 2)[1])

				if err != nil || err2 != nil {
					return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				}

				startTime = startTime.AddDate(1970, 0, 0)
				endTime = endTime.AddDate(1970, 0, 0)

				// add the period to our list
				periodItem := CalendarPeriod{
					-1,
					int(periodInfo["SectionId"].(float64)),
					dayNumber,
					int(startTime.Unix()),
					int(endTime.Unix()),
					-1,
				}

				dayMap[term][dayNumber] = append(dayMap[term][dayNumber], periodItem)
			}
		}

		// add all of this to mysql
		// in one giant transaction
		tx, err := DB.Begin()

		// clear away anything that is in the db
		termDeleteStmt, err := tx.Prepare("DELETE FROM calendar_terms WHERE userId = ?")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer termDeleteStmt.Close()
		termDeleteStmt.Exec(userId)

		classDeleteStmt, err := tx.Prepare("DELETE FROM calendar_classes WHERE userId = ?")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer classDeleteStmt.Close()
		classDeleteStmt.Exec(userId)

		periodsDeleteStmt, err := tx.Prepare("DELETE FROM calendar_periods WHERE userId = ?")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer periodsDeleteStmt.Close()
		periodsDeleteStmt.Exec(userId)

		statusDeleteStmt, err := tx.Prepare("DELETE FROM calendar_status WHERE userId = ?")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer statusDeleteStmt.Close()
		statusDeleteStmt.Exec(userId)

		// first add the terms
		termInsertStmt, err := tx.Prepare("INSERT INTO calendar_terms(termId, name, userId) VALUES(?, ?, ?)")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer termInsertStmt.Close()
		for _, term := range termMap {
			_, err = termInsertStmt.Exec(term.TermID, term.Name, userId)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}
		}

		// then the classes
		classInsertStmt, err := tx.Prepare("INSERT INTO calendar_classes(termId, ownerId, sectionId, name, ownerName, userId) VALUES(?, ?, ?, ?, ?, ?)")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer classInsertStmt.Close()
		for _, class := range classMap {
			_, err = classInsertStmt.Exec(class.TermID, class.OwnerID, class.SectionID, class.Name, class.OwnerName, userId)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}
		}

		// and finally the periods
		periodsInsertStmt, err := tx.Prepare("INSERT INTO calendar_periods(classId, dayNumber, start, end, userId) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer periodsInsertStmt.Close()
		for _, term := range dayMap {
			for _, periods := range term {
				for _, period := range periods {
					_, err = periodsInsertStmt.Exec(period.ClassID, period.DayNumber, period.Start, period.End, userId)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
					}
				}
			}
		}

		statusInsertStmt, err := tx.Prepare("INSERT INTO calendar_status(status, userId) VALUES(1, ?)")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer statusInsertStmt.Close()
		_, err = statusInsertStmt.Exec(userId)

		// go!
		err = tx.Commit()
		if err != nil {
			log.Println("Error while adding schedule to DB")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
