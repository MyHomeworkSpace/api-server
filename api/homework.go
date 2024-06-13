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

	"github.com/julienschmidt/httprouter"
)

// responses
type homeworkResponse struct {
	Status   string          `json:"status"`
	Homework []data.Homework `json:"homework"`
}
type hwViewResponse struct {
	Status       string          `json:"status"`
	TomorrowName string          `json:"tomorrowName"`
	ShowToday    bool            `json:"showToday"`
	Overdue      []data.Homework `json:"overdue"`
	Today        []data.Homework `json:"today"`
	Tomorrow     []data.Homework `json:"tomorrow"`
	Soon         []data.Homework `json:"soon"`
	Longterm     []data.Homework `json:"longterm"`
}
type singleHomeworkResponse struct {
	Status   string        `json:"status"`
	Homework data.Homework `json:"homework"`
}

func routeHomeworkGet(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? ORDER BY `due` ASC", c.User.ID)
	if err != nil {
		errorlog.LogError("getting homework information", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	homework := []data.Homework{}
	for rows.Next() {
		resp := data.Homework{}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
		homework = append(homework, resp)
	}
	writeJSON(w, http.StatusOK, homeworkResponse{"ok", homework})
}

func routeHomeworkGetForClass(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	// verify the class exists and the user owns it
	classIDStr := p.ByName("classId")
	if classIDStr == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	classID, err := strconv.Atoi(classIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	classRows, err := DB.Query("SELECT id FROM classes WHERE id = ? AND userId = ?", classID, c.User.ID)
	if err != nil {
		errorlog.LogError("getting homework for class", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer classRows.Close()

	if !classRows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	// actually get the homework
	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE classId = ? AND userId = ? ORDER BY `due` ASC, id ASC", classID, c.User.ID)
	if err != nil {
		errorlog.LogError("getting homework for class", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	homework := []data.Homework{}
	for rows.Next() {
		resp := data.Homework{}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
		homework = append(homework, resp)
	}
	writeJSON(w, http.StatusOK, homeworkResponse{"ok", homework})
}

func routeHomeworkGetHWView(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	// look for hidden class pref
	hiddenPref, err := data.GetPrefForUser("homeworkHiddenClasses", c.User.ID)
	hiddenClasses := []int{}
	if err != nil && err != data.ErrNotFound {
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
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
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	homework := []data.Homework{}
	for rows.Next() {
		resp := data.Homework{}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)

		if util.IntSliceContains(hiddenClasses, resp.ClassID) {
			continue
		}

		homework = append(homework, resp)
	}
	writeJSON(w, http.StatusOK, homeworkResponse{"ok", homework})
}

func routeHomeworkGetHWViewSorted(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	showTodayStr := r.FormValue("showToday")
	showToday := false

	if showTodayStr == "true" {
		showToday = true
	}

	// look for hidden class pref
	hiddenPref, err := data.GetPrefForUser("homeworkHiddenClasses", c.User.ID)
	hiddenClasses := []int{}
	if err != nil && err != data.ErrNotFound {
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
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
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
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

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		errorlog.LogError("getting homework view", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
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
		resp := data.Homework{}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
		dueDate, err := time.ParseInLocation("2006-01-02", resp.Due, location)
		if err != nil {
			errorlog.LogError("getting homework view", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
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

	writeJSON(w, http.StatusOK, hwViewResponse{
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

func routeHomeworkGetID(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND id = ?", c.User.ID, p.ByName("id"))
	if err != nil {
		errorlog.LogError("getting homework information", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		writeJSON(w, http.StatusNotFound, errorResponse{"error", "not_found"})
		return
	}

	resp := data.Homework{}
	rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)

	writeJSON(w, http.StatusOK, singleHomeworkResponse{"ok", resp})
}

func routeHomeworkGetWeek(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	startDate, err := time.Parse("2006-01-02", p.ByName("date"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}
	endDate := startDate.Add(time.Hour * 24 * 7)

	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (due >= ? and due < ?)", c.User.ID, startDate, endDate)
	if err != nil {
		errorlog.LogError("getting homework information", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	homework := []data.Homework{}
	for rows.Next() {
		resp := data.Homework{}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
		homework = append(homework, resp)
	}
	writeJSON(w, http.StatusOK, homeworkResponse{"ok", homework})
}

func routeHomeworkGetPickerSuggestions(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	rows, err := DB.Query(
		"SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND due >= ? AND id NOT IN (SELECT homework.id FROM homework INNER JOIN calendar_hwevents ON calendar_hwevents.homeworkId = homework.id) ORDER BY `due` ASC",
		c.User.ID,
		time.Now().Format("2006-01-02"),
	)
	if err != nil {
		errorlog.LogError("getting homework picker suggestions", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	homework := []data.Homework{}
	for rows.Next() {
		resp := data.Homework{}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
		homework = append(homework, resp)
	}
	writeJSON(w, http.StatusOK, homeworkResponse{"ok", homework})
}

func routeHomeworkSearch(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("q") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	query := r.FormValue("q")

	// sanitize the query, see https://githubengineering.com/like-injection/ for details
	query = strings.Replace(query, "\\", "\\\\", -1)
	query = strings.Replace(query, "%", "\\%", -1)
	query = strings.Replace(query, "_", "\\_", -1)
	query = "%" + query + "%"

	rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (`name` LIKE ? OR `desc` LIKE ?) ORDER BY `due` DESC", c.User.ID, query, query)
	if err != nil {
		errorlog.LogError("getting homework search results", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	homework := []data.Homework{}
	for rows.Next() {
		resp := data.Homework{}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
		homework = append(homework, resp)
	}
	writeJSON(w, http.StatusOK, homeworkResponse{"ok", homework})
}

func routeHomeworkAdd(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("name") == "" || r.FormValue("due") == "" || r.FormValue("complete") == "" || r.FormValue("classId") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}
	if r.FormValue("complete") != "0" && r.FormValue("complete") != "1" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// check if you are allowed to add to the given classId
	rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("classId"))
	if err != nil {
		errorlog.LogError("adding homework", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec(
		"INSERT INTO homework(name, `due`, `desc`, `complete`, classId, userId) VALUES(?, ?, ?, ?, ?, ?)",
		r.FormValue("name"), r.FormValue("due"), r.FormValue("desc"), r.FormValue("complete"), r.FormValue("classId"), c.User.ID,
	)
	if err != nil {
		errorlog.LogError("adding homework", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeHomeworkEdit(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("id") == "" || r.FormValue("name") == "" || r.FormValue("due") == "" || r.FormValue("complete") == "" || r.FormValue("classId") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}
	if r.FormValue("complete") != "0" && r.FormValue("complete") != "1" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id"))
	if err != nil {
		errorlog.LogError("editing homework", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	// check if you are allowed to add to the given classId
	classRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("classId"))
	if err != nil {
		errorlog.LogError("editing homework", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer classRows.Close()
	if !classRows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec(
		"UPDATE homework SET name = ?, `due` = ?, `desc` = ?, `complete` = ?, classId = ? WHERE id = ?",
		r.FormValue("name"), r.FormValue("due"), r.FormValue("desc"), r.FormValue("complete"), r.FormValue("classId"), r.FormValue("id"),
	)
	if err != nil {
		errorlog.LogError("editing homework", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeHomeworkDelete(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("id") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting homework", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	deleteTx, err := DB.Begin()

	// delete the homework records
	_, err = deleteTx.Exec("DELETE FROM homework WHERE id = ?", r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting homework", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// delete any associated calendar events
	_, err = deleteTx.Exec("DELETE FROM calendar_hwevents WHERE homeworkId = ?", r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting homework", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	err = deleteTx.Commit()
	if err != nil {
		errorlog.LogError("deleting homework", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeHomeworkMarkOverdueDone(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	// look for hidden class pref
	hiddenPref, err := data.GetPrefForUser("homeworkHiddenClasses", c.User.ID)
	hiddenClasses := []int{}
	if err != nil && err != data.ErrNotFound {
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
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

	_, err = DB.Exec("UPDATE homework SET complete = 1 WHERE due < NOW() - INTERVAL 1 DAY AND userId = ? AND FIND_IN_SET(classId, ?) = 0", c.User.ID, hiddenClassesSet)
	if err != nil {
		errorlog.LogError("marking overdue homework as done", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}
