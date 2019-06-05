package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/util"

	"github.com/labstack/echo"
)

// responses
type HomeworkResponse struct {
	Status   string          `json:"status"`
	Homework []data.Homework `json:"homework"`
}
type HWViewResponse struct {
	Status       string          `json:"status"`
	TomorrowName string          `json:"tomorrowName"`
	ShowToday    bool            `json:"showToday"`
	Overdue      []data.Homework `json:"overdue"`
	Today        []data.Homework `json:"today"`
	Tomorrow     []data.Homework `json:"tomorrow"`
	Soon         []data.Homework `json:"soon"`
	Longterm     []data.Homework `json:"longterm"`
}
type SingleHomeworkResponse struct {
	Status   string        `json:"status"`
	Homework data.Homework `json:"homework"`
}

func InitHomeworkAPI(e *echo.Echo) {
	e.GET("/homework/get", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? ORDER BY `due` ASC", GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("getting homework information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		homework := []data.Homework{}
		for rows.Next() {
			resp := data.Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			homework = append(homework, resp)
		}
		return c.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
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
			ErrorLog_LogError("getting homework for class", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer classRows.Close()

		if !classRows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		// actually get the homework
		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE classId = ? AND userId = ? ORDER BY `due` ASC", classId, GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("getting homework for class", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		homework := []data.Homework{}
		for rows.Next() {
			resp := data.Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			homework = append(homework, resp)
		}
		return c.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
	})
	e.GET("/homework/getHWView", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		// look for hidden class pref
		hiddenPref, err := Data_GetPrefForUser("homeworkHiddenClasses", GetSessionUserID(&c))
		hiddenClasses := []int{}
		if err != nil && err != data.ErrNotFound {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		} else if err == nil {
			err = json.Unmarshal([]byte(hiddenPref.Value), &hiddenClasses)
			if err != nil {
				// just ignore the error
				hiddenClasses = []int{}
			}
		}

		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (`due` > (NOW() - INTERVAL 2 DAY) OR `complete` != '1') ORDER BY `due` ASC", GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("getting homework information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		homework := []data.Homework{}
		for rows.Next() {
			resp := data.Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)

			if util.IntSliceContains(hiddenClasses, resp.ClassID) {
				continue
			}

			homework = append(homework, resp)
		}
		return c.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
	})
	e.GET("/homework/getHWViewSorted", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		showTodayStr := c.FormValue("showToday")
		showToday := false

		if showTodayStr == "true" {
			showToday = true
		}

		// look for hidden class pref
		hiddenPref, err := Data_GetPrefForUser("homeworkHiddenClasses", GetSessionUserID(&c))
		hiddenClasses := []int{}
		if err != nil && err != data.ErrNotFound {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		} else if err == nil {
			err = json.Unmarshal([]byte(hiddenPref.Value), &hiddenClasses)
			if err != nil {
				// just ignore the error
				hiddenClasses = []int{}
			}
		}

		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (`due` > (NOW() - INTERVAL 3 DAY) OR `complete` != '1') ORDER BY `due` ASC", GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("getting homework view", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		overdue := []data.Homework{}
		today := []data.Homework{}
		tomorrowName := "Tomorrow"
		tomorrow := []data.Homework{}
		soon := []data.Homework{}
		longterm := []data.Homework{}

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

		if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
			showToday = false
		}

		for rows.Next() {
			resp := data.Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			dueDate, err := time.ParseInLocation("2006-01-02", resp.Due, location)
			if err != nil {
				ErrorLog_LogError("getting homework view", err)
				return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			}

			if util.IntSliceContains(hiddenClasses, resp.ClassID) {
				continue
			}

			timeUntilDue := dueDate.Sub(now)
			if timeUntilDue < 0 {
				if timeUntilDue > 0-(24*time.Hour) {
					// it's in the today column
					today = append(today, resp)
				} else if resp.Complete == 0 {
					// it's overdue
					overdue = append(overdue, resp)
				}
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

		if !showToday {
			for _, item := range today {
				if item.Complete == 0 {
					overdue = append(overdue, item)
				}
			}
			today = []data.Homework{}
		}

		return c.JSON(http.StatusOK, HWViewResponse{
			"ok",
			tomorrowName,
			showToday,
			overdue,
			today,
			tomorrow,
			soon,
			longterm,
		})
	})

	e.GET("/homework/get/:id", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.Param("id"))
		if err != nil {
			ErrorLog_LogError("getting homework information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		if !rows.Next() {
			return c.JSON(http.StatusNotFound, ErrorResponse{"error", "not_found"})
		}

		resp := data.Homework{-1, "", "", "", -1, -1, -1}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)

		return c.JSON(http.StatusOK, SingleHomeworkResponse{"ok", resp})
	})

	e.GET("/homework/getWeek/:monday", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		startDate, err := time.Parse("2006-01-02", c.Param("monday"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}
		endDate := startDate.Add(time.Hour * 24 * 7)

		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (due >= ? and due < ?)", GetSessionUserID(&c), startDate, endDate)
		if err != nil {
			ErrorLog_LogError("getting homework information", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		homework := []data.Homework{}
		for rows.Next() {
			resp := data.Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			homework = append(homework, resp)
		}
		return c.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
	})

	e.GET("/homework/getPickerSuggestions", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND due > NOW() AND id NOT IN (SELECT homework.id FROM homework INNER JOIN calendar_hwevents ON calendar_hwevents.homeworkId = homework.id) ORDER BY `due` DESC", GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("getting homework picker suggestions", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		homework := []data.Homework{}
		for rows.Next() {
			resp := data.Homework{-1, "", "", "", -1, -1, -1}
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
			ErrorLog_LogError("getting homework search results", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		homework := []data.Homework{}
		for rows.Next() {
			resp := data.Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			homework = append(homework, resp)
		}
		return c.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
	})

	e.POST("/homework/add", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if c.FormValue("name") == "" || c.FormValue("due") == "" || c.FormValue("complete") == "" || c.FormValue("classId") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}
		if c.FormValue("complete") != "0" && c.FormValue("complete") != "1" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		// check if you are allowed to add to the given classId
		rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("classId"))
		if err != nil {
			ErrorLog_LogError("adding homework", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		if !rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("INSERT INTO homework(name, `due`, `desc`, `complete`, classId, userId) VALUES(?, ?, ?, ?, ?, ?)")
		if err != nil {
			ErrorLog_LogError("adding homework", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("name"), c.FormValue("due"), c.FormValue("desc"), c.FormValue("complete"), c.FormValue("classId"), GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("adding homework", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/homework/edit", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if c.FormValue("id") == "" || c.FormValue("name") == "" || c.FormValue("due") == "" || c.FormValue("complete") == "" || c.FormValue("classId") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}
		if c.FormValue("complete") != "0" && c.FormValue("complete") != "1" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		// check if you are allowed to edit the given id
		idRows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("editing homework", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer idRows.Close()
		if !idRows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		// check if you are allowed to add to the given classId
		classRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("classId"))
		if err != nil {
			ErrorLog_LogError("editing homework", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer classRows.Close()
		if !classRows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("UPDATE homework SET name = ?, `due` = ?, `desc` = ?, `complete` = ?, classId = ? WHERE id = ?")
		if err != nil {
			ErrorLog_LogError("editing homework", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("name"), c.FormValue("due"), c.FormValue("desc"), c.FormValue("complete"), c.FormValue("classId"), c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("editing homework", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/homework/delete", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("id") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		// check if you are allowed to edit the given id
		idRows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("deleting homework", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer idRows.Close()
		if !idRows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		deleteTx, err := DB.Begin()

		// delete the homework records
		_, err = deleteTx.Exec("DELETE FROM homework WHERE id = ?", c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("deleting homework", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		// delete any associated calendar events
		_, err = deleteTx.Exec("DELETE FROM calendar_hwevents WHERE homeworkId = ?", c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("deleting homework", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		err = deleteTx.Commit()
		if err != nil {
			ErrorLog_LogError("deleting homework", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/homework/markOverdueDone", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		// look for hidden class pref
		hiddenPref, err := Data_GetPrefForUser("homeworkHiddenClasses", GetSessionUserID(&c))
		hiddenClasses := []int{}
		if err != nil && err != data.ErrNotFound {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		} else if err == nil {
			err = json.Unmarshal([]byte(hiddenPref.Value), &hiddenClasses)
			if err != nil {
				// just ignore the error
				hiddenClasses = []int{}
			}
		}

		hiddenClassesSet := ""
		for i, hiddenClassID := range hiddenClasses {
			if i > 0 {
				hiddenClassesSet = hiddenClassesSet + ","
			}
			hiddenClassesSet = hiddenClassesSet + strconv.Itoa(hiddenClassID)
		}

		_, err = DB.Exec("UPDATE homework SET complete = 1 WHERE due < NOW() AND userId = ? AND FIND_IN_SET(classId, ?) = 0", GetSessionUserID(&c), hiddenClassesSet)
		if err != nil {
			ErrorLog_LogError("marking overdue homework as done", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
