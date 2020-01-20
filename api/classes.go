package api

import (
	"net/http"
	"strconv"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/MyHomeworkSpace/api-server/util"

	"github.com/julienschmidt/httprouter"
)

// don't change these!
var DefaultColors = []string{
	"ff4d40",
	"ffa540",
	"40ff73",
	"4071ff",
	"ff4086",
	"40ccff",
	"5940ff",
	"ff40f5",
	"a940ff",
	"e6ab68",
	"4d4d4d",
}

// responses
type classResponse struct {
	Status  string               `json:"status"`
	Classes []data.HomeworkClass `json:"classes"`
}
type singleClassResponse struct {
	Status string             `json:"status"`
	Class  data.HomeworkClass `json:"class"`
}
type hwInfoResponse struct {
	Status  string `json:"status"`
	HWItems int    `json:"hwItems"`
}

func routeClassesGet(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	classes, err := data.GetClassesForUser(c.User)
	if err != nil {
		errorlog.LogError("getting list of classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, classResponse{"ok", classes})
}

func routeClassesGetID(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, teacher, color, sortIndex, userId FROM classes WHERE id = ? AND userId = ?", p.ByName("id"), c.User.ID)
	if err != nil {
		errorlog.LogError("getting class information", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}
	resp := data.HomeworkClass{-1, "", "", "", -1, -1}
	rows.Scan(&resp.ID, &resp.Name, &resp.Teacher, &resp.Color, &resp.SortIndex, &resp.UserID)
	if resp.Color == "" {
		resp.Color = DefaultColors[resp.ID%len(DefaultColors)]
	}
	writeJSON(w, http.StatusOK, singleClassResponse{"ok", resp})
}

func routeClassesHWInfo(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	rows, err := DB.Query("SELECT COUNT(*) FROM homework WHERE classId = ? AND userId = ?", p.ByName("id"), c.User.ID)
	if err != nil {
		errorlog.LogError("getting class information", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}
	resp := -1
	rows.Scan(&resp)
	writeJSON(w, http.StatusOK, hwInfoResponse{"ok", resp})
}

func routeClassesAdd(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("name") == "" || r.FormValue("color") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}
	if !util.StringSliceContains(DefaultColors, r.FormValue("color")) {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	_, err := DB.Exec(
		"INSERT INTO classes(name, teacher, color, sortIndex, userId) VALUES(?, ?, ?, (SELECT * FROM (SELECT COUNT(*) FROM classes WHERE userId = ?) AS sortIndex), ?)",
		r.FormValue("name"), r.FormValue("teacher"), r.FormValue("color"), c.User.ID, c.User.ID,
	)
	if err != nil {
		errorlog.LogError("adding class", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeClassesEdit(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("id") == "" || r.FormValue("name") == "" || r.FormValue("color") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}
	if !util.StringSliceContains(DefaultColors, r.FormValue("color")) {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id"))
	if err != nil {
		errorlog.LogError("editing classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec(
		"UPDATE classes SET name = ?, teacher = ?, color = ? WHERE id = ?",
		r.FormValue("name"), r.FormValue("teacher"), r.FormValue("color"), r.FormValue("id"),
	)
	if err != nil {
		errorlog.LogError("editing class", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeClassesDelete(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("id") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// check if you are allowed to delete the given id
	idRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, id)
	if err != nil {
		errorlog.LogError("deleting classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	if !idRows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}
	idRows.Close()

	// get the user's current classes so that we can renumber the sort indices
	currentClasses := []int{}
	allIDRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? ORDER BY sortIndex ASC", c.User.ID)
	if err != nil {
		errorlog.LogError("deleting classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	for allIDRows.Next() {
		id := -1

		err = allIDRows.Scan(&id)
		if err != nil {
			errorlog.LogError("deleting classes", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		currentClasses = append(currentClasses, id)
	}
	allIDRows.Close()

	// use a transaction so that you can't delete just the hw or the class entry - either both or nothing
	tx, err := DB.Begin()

	// delete HW calendar events
	_, err = tx.Exec("DELETE calendar_hwevents FROM calendar_hwevents INNER JOIN homework ON calendar_hwevents.homeworkId = homework.id WHERE homework.classId = ?", id)
	if err != nil {
		tx.Rollback()
		errorlog.LogError("deleting class", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// delete HW
	_, err = tx.Exec("DELETE FROM homework WHERE classId = ?", id)
	if err != nil {
		tx.Rollback()
		errorlog.LogError("deleting class", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// delete class
	_, err = tx.Exec("DELETE FROM classes WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		errorlog.LogError("deleting class", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// update all other classes
	currentSortIndex := 0
	for _, classID := range currentClasses {
		if classID == id {
			continue
		}

		_, err = tx.Exec("UPDATE classes SET sortIndex = ? WHERE id = ?", currentSortIndex, classID)
		if err != nil {
			tx.Rollback()
			errorlog.LogError("deleting class", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}

		currentSortIndex++
	}

	// go!
	err = tx.Commit()
	if err != nil {
		errorlog.LogError("deleting class", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeClassesSwap(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("id1") == "" || r.FormValue("id2") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	id1, err := strconv.Atoi(r.FormValue("id1"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}
	id2, err := strconv.Atoi(r.FormValue("id2"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// get class 1, check it's yours
	class1Row, err := DB.Query("SELECT sortIndex FROM classes WHERE userId = ? AND id = ?", c.User.ID, id1)
	if err != nil {
		errorlog.LogError("deleting classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	if !class1Row.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	class1SortIndex := -1
	class1Row.Scan(&class1SortIndex)
	class1Row.Close()

	// get class 2, check it's yours
	class2Row, err := DB.Query("SELECT sortIndex FROM classes WHERE userId = ? AND id = ?", c.User.ID, id2)
	if err != nil {
		errorlog.LogError("deleting classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	if !class2Row.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	class2SortIndex := -1
	class2Row.Scan(&class2SortIndex)
	class2Row.Close()

	tx, err := DB.Begin()
	if err != nil {
		errorlog.LogError("swapping classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	_, err = tx.Exec("UPDATE classes SET sortIndex = ? WHERE id = ?", class2SortIndex, id1)
	if err != nil {
		errorlog.LogError("swapping classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	_, err = tx.Exec("UPDATE classes SET sortIndex = ? WHERE id = ?", class1SortIndex, id2)
	if err != nil {
		errorlog.LogError("swapping classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// go!
	err = tx.Commit()
	if err != nil {
		errorlog.LogError("swapping class", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}
