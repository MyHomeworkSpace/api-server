package api

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
)

type FacultyListResponse struct {
	Status      string        `json:"status"`
	FacultyList []FacultyInfo `json:"faculty"`
}

type FacultyPeriod struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	SectionID int    `json:"sectionId"`
	Room      string `json:"room"`
	Block     string `json:"block"`
	DayNumber int    `json:"dayNumber"`
	Grade     int    `json:"grade"`
	Term      int    `json:"term"`
	Start     int    `json:"start"`
	End       int    `json:"end"`
	FacultyID int    `json:"facultyId"`
}

func InitScheduleAPI(e *echo.Echo) {
	e.GET("/schedule/getFacultyList", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		rows, err := DB.Query("SELECT bbId, firstName, lastName, largeFileName, department, grades FROM faculty ORDER BY lastName ASC")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		users := []FacultyInfo{}
		for rows.Next() {
			user := FacultyInfo{}
			rows.Scan(&user.BlackbaudUserID, &user.FirstName, &user.LastName, &user.LargeFileName, &user.DepartmentDisplay, &user.GradeNumericDisplay)
			users = append(users, user)
		}

		return c.JSON(http.StatusOK, FacultyListResponse{"ok", users})
	})

	// internal things
	// used for fetching things from blackbaud
	e.POST("/schedule/internal/importFaculty", func(c echo.Context) error {
		if !strings.HasPrefix(c.Request().RemoteAddr, "127.0.0.1") {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
		}

		if c.FormValue("t") == "" || c.FormValue("targetId") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		targetId, err := strconv.Atoi(c.FormValue("targetId"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "invalid_params"})
		}

		jar, err := cookiejar.New(nil)
		if err != nil {
			log.Println("Error while importing faculty schedule")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		mySchoolAppURL, err := url.Parse("https://dalton.myschoolapp.com")
		if err != nil {
			log.Println("Error while importing faculty schedule")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		jar.SetCookies(mySchoolAppURL, []*http.Cookie{
			&http.Cookie{
				Name:  "t",
				Value: c.FormValue("t"),
			},
		})

		// import their schedule
		dayMap := map[int]map[int][]FacultyPeriod{}
		for _, term := range []int{1, 2} {
			dayMap[term] = map[int][]FacultyPeriod{
				0: []FacultyPeriod{},
				1: []FacultyPeriod{},
				2: []FacultyPeriod{},
				3: []FacultyPeriod{},
				4: []FacultyPeriod{},
				5: []FacultyPeriod{},
				6: []FacultyPeriod{},
				7: []FacultyPeriod{},
			}

			startDate := Term1_Import_Start
			endDate := Term1_Import_End
			if term == 2 {
				startDate = Term2_Import_Start
				endDate = Term2_Import_End
			}

			response, err := Blackbaud_Request("GET", "DataDirect/ScheduleList", url.Values{
				"format":          {"json"},
				"viewerId":        {strconv.Itoa(targetId)},
				"personaId":       {"3"},
				"viewerPersonaId": {"2"},
				"start":           {strconv.FormatInt(startDate.Unix(), 10)},
				"end":             {strconv.FormatInt(endDate.Unix(), 10)},
			}, map[string]interface{}{}, jar, "")
			if err != nil {
				log.Println("Error while importing faculty schedule")
				log.Println(err)
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
					fridayNumber, ok := ScheduleFridayList[day.Format("2006-01-02")]
					if !ok {
						log.Println("Error while importing faculty schedule")
						log.Println("Couldn't find Friday number")
						log.Println(day)
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

				startTime, err := time.Parse("3:04 PM", strings.SplitN(periodInfo["start"].(string), " ", 2)[1])
				endTime, err2 := time.Parse("3:04 PM", strings.SplitN(periodInfo["end"].(string), " ", 2)[1])

				if err != nil || err2 != nil {
					log.Println("Error while importing faculty schedule")
					log.Println(err)
					return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
				}

				startTime = startTime.AddDate(1970, 0, 0)
				endTime = endTime.AddDate(1970, 0, 0)

				// add the period to our list
				periodItem := FacultyPeriod{
					Name:      periodInfo["title"].(string),
					SectionID: int(periodInfo["SectionId"].(float64)),
					Room:      "",
					Block:     "",
					DayNumber: dayNumber,
					Grade:     -1,
					Term:      term,
					Start:     int(startTime.Unix()),
					End:       int(endTime.Unix()),
					FacultyID: targetId,
				}

				dayMap[term][dayNumber] = append(dayMap[term][dayNumber], periodItem)
			}
		}

		tx, err := DB.Begin()

		// clear away existing periods
		periodDeleteStmt, err := tx.Prepare("DELETE FROM faculty_periods WHERE facultyId = ?")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer periodDeleteStmt.Close()
		periodDeleteStmt.Exec(targetId)

		// add imported data
		periodInsertStmt, err := tx.Prepare("INSERT INTO faculty_periods(name, sectionId, room, block, dayNumber, grade, term, start, end, facultyId) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer periodInsertStmt.Close()
		for _, term := range dayMap {
			for _, periods := range term {
				for _, period := range periods {
					_, err = periodInsertStmt.Exec(period.Name, period.SectionID, period.Room, period.Block, period.DayNumber, period.Grade, period.Term, period.Start, period.End, targetId)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
					}
				}
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Println("Error while adding faculty schedule to DB")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/schedule/internal/importMetadata", func(c echo.Context) error {
		if !strings.HasPrefix(c.Request().RemoteAddr, "127.0.0.1") {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "forbidden"})
		}

		if c.FormValue("t") == "" || c.FormValue("sectionId") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		sectionId, err := strconv.Atoi(c.FormValue("sectionId"))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "invalid_params"})
		}

		jar, err := cookiejar.New(nil)
		if err != nil {
			log.Println("Error while importing class metadata")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		mySchoolAppURL, err := url.Parse("https://dalton.myschoolapp.com")
		if err != nil {
			log.Println("Error while importing class metadata")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		jar.SetCookies(mySchoolAppURL, []*http.Cookie{
			&http.Cookie{
				Name:  "t",
				Value: c.FormValue("t"),
			},
		})

		response, err := Blackbaud_Request("GET", "datadirect/SectionInfoView", url.Values{
			"format":        {"json"},
			"sectionId":     {strconv.Itoa(sectionId)},
			"associationId": {"1"},
		}, map[string]interface{}{}, jar, "")
		if err != nil {
			log.Println("Error while importing class metadata")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		options, ok := response.([]interface{})
		if !ok || len(options) == 0 {
			// there's no information on the section page...
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "no_metadata_page"})
		}
		classData := options[0].(map[string]interface{})

		roomParts := strings.Split(classData["Room"].(string), " ")

		room := roomParts[len(roomParts)-1]
		block := classData["Block"].(string)
		grade := int(classData["LevelNum"].(float64))

		_, err = DB.Exec("UPDATE faculty_periods SET room = ?, block = ?, grade = ? WHERE sectionId = ?", room, block, grade, sectionId)
		if err != nil {
			log.Println("Error while importing class metadata")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
