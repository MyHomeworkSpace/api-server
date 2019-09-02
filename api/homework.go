package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"
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

func routeHomeworkGet(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? ORDER BY `due` ASC", c.User.ID)
	if err != nil {
		errorlog.LogError("getting homework information", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	homework := []data.Homework{}
	for rows.Next() {
		resp := data.Homework{-1, "", "", "", -1, -1, -1}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
		homework = append(homework, resp)
	}
	ec.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
}

func routeHomeworkGetForClass(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	// verify the class exists and the user owns it
	classIdStr := ec.Param("classId")
	if classIdStr == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	classId, err := strconv.Atoi(classIdStr)
	if err != nil {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}

	classRows, err := DB.Query("SELECT id FROM classes WHERE id = ? AND userId = ?", classId, c.User.ID)
	if err != nil {
		errorlog.LogError("getting homework for class", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer classRows.Close()

	if !classRows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	// actually get the homework
	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE classId = ? AND userId = ? ORDER BY `due` ASC", classId, c.User.ID)
	if err != nil {
		errorlog.LogError("getting homework for class", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	homework := []data.Homework{}
	for rows.Next() {
		resp := data.Homework{-1, "", "", "", -1, -1, -1}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
		homework = append(homework, resp)
	}
	ec.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
}

func routeHomeworkGetHWView(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	// look for hidden class pref
	hiddenPref, err := data.GetPrefForUser("homeworkHiddenClasses", c.User.ID)
	hiddenClasses := []int{}
	if err != nil && err != data.ErrNotFound {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	} else if err == nil {
		err = json.Unmarshal([]byte(hiddenPref.Value), &hiddenClasses)
		if err != nil {
			// just ignore the error
			hiddenClasses = []int{}
		}
	}

	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (`due` > (NOW() - INTERVAL 2 DAY) OR `complete` != '1') ORDER BY `due` ASC", c.User.ID)
	if err != nil {
		errorlog.LogError("getting homework information", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
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
	ec.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
}

func routeHomeworkGetHWViewSorted(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	showTodayStr := ec.FormValue("showToday")
	showToday := false

	if showTodayStr == "true" {
		showToday = true
	}

	// look for hidden class pref
	hiddenPref, err := data.GetPrefForUser("homeworkHiddenClasses", c.User.ID)
	hiddenClasses := []int{}
	if err != nil && err != data.ErrNotFound {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	} else if err == nil {
		err = json.Unmarshal([]byte(hiddenPref.Value), &hiddenClasses)
		if err != nil {
			// just ignore the error
			hiddenClasses = []int{}
		}
	}

	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (`due` > (NOW() - INTERVAL 3 DAY) OR `complete` != '1') ORDER BY `due` ASC", c.User.ID)
	if err != nil {
		errorlog.LogError("getting homework view", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
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
			errorlog.LogError("getting homework view", err)
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
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

	ec.JSON(http.StatusOK, HWViewResponse{
		"ok",
		tomorrowName,
		showToday,
		overdue,
		today,
		tomorrow,
		soon,
		longterm,
	})
}

func routeHomeworkGetID(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND id = ?", c.User.ID, ec.Param("id"))
	if err != nil {
		errorlog.LogError("getting homework information", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		ec.JSON(http.StatusNotFound, ErrorResponse{"error", "not_found"})
		return
	}

	resp := data.Homework{-1, "", "", "", -1, -1, -1}
	rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)

	ec.JSON(http.StatusOK, SingleHomeworkResponse{"ok", resp})
}

func routeHomeworkGetWeek(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	startDate, err := time.Parse("2006-01-02", ec.Param("monday"))
	if err != nil {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}
	endDate := startDate.Add(time.Hour * 24 * 7)

	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (due >= ? and due < ?)", c.User.ID, startDate, endDate)
	if err != nil {
		errorlog.LogError("getting homework information", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	homework := []data.Homework{}
	for rows.Next() {
		resp := data.Homework{-1, "", "", "", -1, -1, -1}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
		homework = append(homework, resp)
	}
	ec.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
}

func routeHomeworkGetPickerSuggestions(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND due > NOW() AND id NOT IN (SELECT homework.id FROM homework INNER JOIN calendar_hwevents ON calendar_hwevents.homeworkId = homework.id) ORDER BY `due` DESC", c.User.ID)
	if err != nil {
		errorlog.LogError("getting homework picker suggestions", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	homework := []data.Homework{}
	for rows.Next() {
		resp := data.Homework{-1, "", "", "", -1, -1, -1}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
		homework = append(homework, resp)
	}
	ec.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
}

func routeHomeworkSearch(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("q") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	query := ec.FormValue("q")

	// sanitize the query, see https://githubengineering.com/like-injection/ for details
	query = strings.Replace(query, "\\", "\\\\", -1)
	query = strings.Replace(query, "%", "\\%", -1)
	query = strings.Replace(query, "_", "\\_", -1)
	query = "%" + query + "%"

	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (`name` LIKE ? OR `desc` LIKE ?) ORDER BY `due` DESC", c.User.ID, query, query)
	if err != nil {
		errorlog.LogError("getting homework search results", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	homework := []data.Homework{}
	for rows.Next() {
		resp := data.Homework{-1, "", "", "", -1, -1, -1}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
		homework = append(homework, resp)
	}
	ec.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
}

func routeHomeworkAdd(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("name") == "" || ec.FormValue("due") == "" || ec.FormValue("complete") == "" || ec.FormValue("classId") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}
	if ec.FormValue("complete") != "0" && ec.FormValue("complete") != "1" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}

	// check if you are allowed to add to the given classId
	rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("classId"))
	if err != nil {
		errorlog.LogError("adding homework", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec(
		"INSERT INTO homework(name, `due`, `desc`, `complete`, classId, userId) VALUES(?, ?, ?, ?, ?, ?)",
		ec.FormValue("name"), ec.FormValue("due"), ec.FormValue("desc"), ec.FormValue("complete"), ec.FormValue("classId"), c.User.ID,
	)
	if err != nil {
		errorlog.LogError("adding homework", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeHomeworkEdit(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("id") == "" || ec.FormValue("name") == "" || ec.FormValue("due") == "" || ec.FormValue("complete") == "" || ec.FormValue("classId") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}
	if ec.FormValue("complete") != "0" && ec.FormValue("complete") != "1" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("editing homework", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	// check if you are allowed to add to the given classId
	classRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("classId"))
	if err != nil {
		errorlog.LogError("editing homework", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer classRows.Close()
	if !classRows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec(
		"UPDATE homework SET name = ?, `due` = ?, `desc` = ?, `complete` = ?, classId = ? WHERE id = ?",
		ec.FormValue("name"), ec.FormValue("due"), ec.FormValue("desc"), ec.FormValue("complete"), ec.FormValue("classId"), ec.FormValue("id"),
	)
	if err != nil {
		errorlog.LogError("editing homework", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeHomeworkDelete(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("id") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting homework", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	deleteTx, err := DB.Begin()

	// delete the homework records
	_, err = deleteTx.Exec("DELETE FROM homework WHERE id = ?", ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting homework", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// delete any associated calendar events
	_, err = deleteTx.Exec("DELETE FROM calendar_hwevents WHERE homeworkId = ?", ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting homework", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	err = deleteTx.Commit()
	if err != nil {
		errorlog.LogError("deleting homework", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeHomeworkMarkOverdueDone(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	// look for hidden class pref
	hiddenPref, err := data.GetPrefForUser("homeworkHiddenClasses", c.User.ID)
	hiddenClasses := []int{}
	if err != nil && err != data.ErrNotFound {
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
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

	_, err = DB.Exec("UPDATE homework SET complete = 1 WHERE due < NOW() AND userId = ? AND FIND_IN_SET(classId, ?) = 0", c.User.ID, hiddenClassesSet)
	if err != nil {
		errorlog.LogError("marking overdue homework as done", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}
