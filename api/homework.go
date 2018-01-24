package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
)

// structs for data
type Homework struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Due      string `json:"due"`
	Desc     string `json:"desc"`
	Complete int    `json:"complete"`
	ClassID  int    `json:"classId"`
	UserID   int    `json:"userId"`
}

// responses
type HomeworkResponse struct {
	Status   string     `json:"status"`
	Homework []Homework `json:"homework"`
}
type HWViewResponse struct {
	Status       string     `json:"status"`
	TomorrowName string     `json:"tomorrowName"`
	Overdue      []Homework `json:"overdue"`
	Tomorrow     []Homework `json:"tomorrow"`
	Soon         []Homework `json:"soon"`
	Longterm     []Homework `json:"longterm"`
}
type SingleHomeworkResponse struct {
	Status   string   `json:"status"`
	Homework Homework `json:"homework"`
}

func InitHomeworkAPI(e *echo.Echo) {
	e.GET("/homework/get", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? ORDER BY `due` ASC", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting homework information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()

		homework := []Homework{}
		for rows.Next() {
			resp := Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			homework = append(homework, resp)
		}
		jsonResp := HomeworkResponse{"ok", homework}
		return c.JSON(http.StatusOK, jsonResp)
	})
	e.GET("/homework/getForClass/:classId", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		// verify the class exists and the user owns it
		classIdStr := c.Param("classId")
		if classIdStr == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		classId, err := strconv.Atoi(classIdStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		classRows, err := DB.Query("SELECT id FROM classes WHERE id = ? AND userId = ?", classId, GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting homework for class:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer classRows.Close()

		if !classRows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		// actually get the homework
		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE classId = ? AND userId = ? ORDER BY `due` ASC", classId, GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting homework for class:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer rows.Close()

		homework := []Homework{}
		for rows.Next() {
			resp := Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			homework = append(homework, resp)
		}
		return c.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
	})
	e.GET("/homework/getHWView", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}

		// look for hidden class pref
		hiddenPref, err := Data_GetPrefForUser("homeworkHiddenClasses", GetSessionUserID(&c))
		hiddenClasses := []int{}
		if err != nil && err != ErrDataNotFound {
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		} else if err == nil {
			err = json.Unmarshal([]byte(hiddenPref.Value), &hiddenClasses)
			if err != nil {
				// just ignore the error
				hiddenClasses = []int{}
			}
		}

		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (`due` > (NOW() - INTERVAL 2 DAY) OR `complete` != '1') ORDER BY `due` ASC", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting homework information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()

		homework := []Homework{}
		for rows.Next() {
			resp := Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)

			if Util_IntSliceContains(hiddenClasses, resp.ClassID) {
				continue
			}

			homework = append(homework, resp)
		}
		jsonResp := HomeworkResponse{"ok", homework}
		return c.JSON(http.StatusOK, jsonResp)
	})
	e.GET("/homework/getHWViewSorted", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}

		// look for hidden class pref
		hiddenPref, err := Data_GetPrefForUser("homeworkHiddenClasses", GetSessionUserID(&c))
		hiddenClasses := []int{}
		if err != nil && err != ErrDataNotFound {
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		} else if err == nil {
			err = json.Unmarshal([]byte(hiddenPref.Value), &hiddenClasses)
			if err != nil {
				// just ignore the error
				hiddenClasses = []int{}
			}
		}

		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (`due` > (NOW() - INTERVAL 2 DAY) OR `complete` != '1') ORDER BY `due` ASC", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting homework view: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer rows.Close()

		overdue := []Homework{}
		tomorrowName := "Tomorrow"
		tomorrow := []Homework{}
		soon := []Homework{}
		longterm := []Homework{}

		tomorrowTimeToThreshold := 24 * time.Hour

		location := time.FixedZone("America/New_York", -5*60*60)
		now := time.Now().In(location)

		if now.Weekday() == time.Friday || now.Weekday() == time.Saturday {
			tomorrowName = "Monday"
			if now.Weekday() == time.Friday {
				tomorrowTimeToThreshold = 3 * 24 * time.Hour
			} else {
				tomorrowTimeToThreshold = 2 * 24 * time.Hour
			}
		}

		for rows.Next() {
			resp := Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			dueDate, err := time.ParseInLocation("2006-01-02", resp.Due, location)
			if err != nil {
				log.Println("Error while getting homework view: ")
				log.Println(err)
				return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
			}

			if Util_IntSliceContains(hiddenClasses, resp.ClassID) {
				continue
			}

			timeUntilDue := dueDate.Sub(now)
			if timeUntilDue < 0 {
				// it's overdue
				overdue = append(overdue, resp)
			} else if timeUntilDue <= tomorrowTimeToThreshold {
				// it's in the tomorrow column
				tomorrow = append(tomorrow, resp)
			} else if timeUntilDue <= 5*24*time.Hour {
				// it's in the soon column
				soon = append(soon, resp)
			} else {
				// it's in the longterm column
				longterm = append(longterm, resp)
			}
		}

		return c.JSON(http.StatusOK, HWViewResponse{
			"ok",
			tomorrowName,
			overdue,
			tomorrow,
			soon,
			longterm,
		})
	})

	e.GET("/homework/get/:id", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.Param("id"))
		if err != nil {
			log.Println("Error while getting homework information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()

		if !rows.Next() {
			jsonResp := ErrorResponse{"error", "not_found"}
			return c.JSON(http.StatusNotFound, jsonResp)
		}

		resp := Homework{-1, "", "", "", -1, -1, -1}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)

		jsonResp := SingleHomeworkResponse{"ok", resp}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.GET("/homework/getWeek/:monday", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}

		startDate, err := time.Parse("2006-01-02", c.Param("monday"))
		if err != nil {
			jsonResp := ErrorResponse{"error", "Invalid date."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}
		endDate := startDate.Add(time.Hour * 24 * 7)

		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (due >= ? and due < ?)", GetSessionUserID(&c), startDate, endDate)
		if err != nil {
			log.Println("Error while getting homework information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()

		homework := []Homework{}
		for rows.Next() {
			resp := Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			homework = append(homework, resp)
		}
		jsonResp := HomeworkResponse{"ok", homework}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.GET("/homework/getPickerSuggestions", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND due > NOW() AND id NOT IN (SELECT homework.id FROM homework INNER JOIN calendar_hwevents ON calendar_hwevents.homeworkId = homework.id) ORDER BY `due` DESC", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting homework picker suggestions:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer rows.Close()

		homework := []Homework{}
		for rows.Next() {
			resp := Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			homework = append(homework, resp)
		}
		return c.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
	})

	e.GET("/homework/search", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if c.FormValue("q") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		query := c.FormValue("q")

		// sanitize the query, see https://githubengineering.com/like-injection/ for details
		query = strings.Replace(query, "\\", "\\\\", -1)
		query = strings.Replace(query, "%", "\\%", -1)
		query = strings.Replace(query, "_", "\\_", -1)
		query = "%" + query + "%"

		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (`name` LIKE ? OR `desc` LIKE ?) ORDER BY `due` DESC", GetSessionUserID(&c), query, query)
		if err != nil {
			log.Println("Error while getting homework search results:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer rows.Close()

		homework := []Homework{}
		for rows.Next() {
			resp := Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			homework = append(homework, resp)
		}
		return c.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
	})

	e.POST("/homework/add", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if c.FormValue("name") == "" || c.FormValue("due") == "" || c.FormValue("complete") == "" || c.FormValue("classId") == "" {
			jsonResp := ErrorResponse{"error", "Missing required parameters."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// check if you are allowed to add to the given classId
		rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("classId"))
		if err != nil {
			log.Println("Error while adding homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()
		if !rows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid class."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		stmt, err := DB.Prepare("INSERT INTO homework(name, `due`, `desc`, `complete`, classId, userId) VALUES(?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Println("Error while adding homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = stmt.Exec(c.FormValue("name"), c.FormValue("due"), c.FormValue("desc"), c.FormValue("complete"), c.FormValue("classId"), GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while adding homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.POST("/homework/edit", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if c.FormValue("id") == "" || c.FormValue("name") == "" || c.FormValue("due") == "" || c.FormValue("complete") == "" || c.FormValue("classId") == "" {
			jsonResp := ErrorResponse{"error", "Missing required parameters."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// check if you are allowed to edit the given id
		idRows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			log.Println("Error while editing homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer idRows.Close()
		if !idRows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid ID."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// check if you are allowed to add to the given classId
		classRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("classId"))
		if err != nil {
			log.Println("Error while editing homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer classRows.Close()
		if !classRows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid class."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		stmt, err := DB.Prepare("UPDATE homework SET name = ?, `due` = ?, `desc` = ?, `complete` = ?, classId = ? WHERE id = ?")
		if err != nil {
			log.Println("Error while editing homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = stmt.Exec(c.FormValue("name"), c.FormValue("due"), c.FormValue("desc"), c.FormValue("complete"), c.FormValue("classId"), c.FormValue("id"))
		if err != nil {
			log.Println("Error while editing homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.POST("/homework/delete", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if c.FormValue("id") == "" {
			jsonResp := ErrorResponse{"error", "Missing ID parameter."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// check if you are allowed to edit the given id
		idRows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer idRows.Close()
		if !idRows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid ID."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		deleteTx, err := DB.Begin()

		// delete the homework records
		_, err = deleteTx.Exec("DELETE FROM homework WHERE id = ?", c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting homework: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		// delete any associated calendar events
		_, err = deleteTx.Exec("DELETE FROM calendar_hwevents WHERE homeworkId = ?", c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting homework: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		err = deleteTx.Commit()
		if err != nil {
			log.Println("Error while deleting homework: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/homework/markOverdueDone", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		_, err := DB.Exec("UPDATE homework SET complete = 1 WHERE due < NOW() AND userId = ?", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while marking overdue homework as done: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
