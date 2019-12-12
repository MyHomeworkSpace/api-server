package api

import (
	"net/http"

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

// structs for data
type HomeworkClass struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Teacher string `json:"teacher"`
	Color   string `json:"color"`
	UserID  int    `json:"userId"`
}

// responses
type classResponse struct {
	Status  string          `json:"status"`
	Classes []HomeworkClass `json:"classes"`
}
type singleClassResponse struct {
	Status string        `json:"status"`
	Class  HomeworkClass `json:"class"`
}
type hwInfoResponse struct {
	Status  string `json:"status"`
	HWItems int    `json:"hwItems"`
}

func routeClassesGet(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, teacher, color, userId FROM classes WHERE userId = ?", c.User.ID)
	if err != nil {
		errorlog.LogError("getting class information", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	classes := []HomeworkClass{}
	for rows.Next() {
		resp := HomeworkClass{-1, "", "", "", -1}
		rows.Scan(&resp.ID, &resp.Name, &resp.Teacher, &resp.Color, &resp.UserID)
		if resp.Color == "" {
			resp.Color = DefaultColors[resp.ID%len(DefaultColors)]
		}
		classes = append(classes, resp)
	}
	writeJSON(w, http.StatusOK, classResponse{"ok", classes})
}

func routeClassesGetID(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, teacher, color, userId FROM classes WHERE id = ? AND userId = ?", p.ByName("id"), c.User.ID)
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
	resp := HomeworkClass{-1, "", "", "", -1}
	rows.Scan(&resp.ID, &resp.Name, &resp.Teacher, &resp.Color, &resp.UserID)
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
		"INSERT INTO classes(name, teacher, color, userId) VALUES(?, ?, ?, ?)",
		r.FormValue("name"), r.FormValue("teacher"), r.FormValue("color"), c.User.ID,
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

	// check if you are allowed to delete the given id
	idRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	// use a transaction so that you can't delete just the hw or the class entry - either both or nothing
	tx, err := DB.Begin()

	// delete HW calendar events
	_, err = tx.Exec("DELETE calendar_hwevents FROM calendar_hwevents INNER JOIN homework ON calendar_hwevents.homeworkId = homework.id WHERE homework.classId = ?", r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting class", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// delete HW
	_, err = tx.Exec("DELETE FROM homework WHERE classId = ?", r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting class", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// delete class
	_, err = tx.Exec("DELETE FROM classes WHERE id = ?", r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting class", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
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

	// check if you are allowed to change id1
	id1Rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id1"))
	if err != nil {
		errorlog.LogError("deleting classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer id1Rows.Close()
	if !id1Rows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	// check if you are allowed to change id2
	id2Rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id2"))
	if err != nil {
		errorlog.LogError("swapping classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer id2Rows.Close()
	if !id2Rows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	// find the swap id
	// this is a dumb way of doing this and is kind of a race condition
	// but mysql has no better way to swap primary keys
	// hopefully no one adds 100 classes in the time this transaction takes to complete
	swapIdStmt, err := DB.Query("SELECT max(id) + 100 FROM classes")
	if err != nil {
		errorlog.LogError("swapping classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	swapId := -1
	defer swapIdStmt.Close()
	swapIdStmt.Next()
	swapIdStmt.Scan(&swapId)

	// use a transaction so that you can't delete just the hw or the class entry - either both or nothing
	tx, err := DB.Begin()

	// update class id1 -> tmp
	_, err = tx.Exec("UPDATE classes SET id = ? WHERE id = ?", swapId, r.FormValue("id1"))
	if err != nil {
		errorlog.LogError("swapping classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// update class id2 -> id1
	_, err = tx.Exec("UPDATE classes SET id = ? WHERE id = ?", r.FormValue("id1"), r.FormValue("id2"))
	if err != nil {
		errorlog.LogError("swapping classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// update class tmp -> id2
	_, err = tx.Exec("UPDATE classes SET id = ? WHERE id = ?", r.FormValue("id2"), swapId)
	if err != nil {
		errorlog.LogError("swapping classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// update homework id1 -> swp
	_, err = tx.Exec("UPDATE homework SET classId = ? WHERE classId = ?", swapId, r.FormValue("id1"))
	if err != nil {
		errorlog.LogError("swapping classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// update homework id2 -> id1
	_, err = tx.Exec("UPDATE homework SET classId = ? WHERE classId = ?", r.FormValue("id1"), r.FormValue("id2"))
	if err != nil {
		errorlog.LogError("swapping classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// update homework swp -> id2
	_, err = tx.Exec("UPDATE homework SET classId = ? WHERE classId = ?", r.FormValue("id2"), swapId)
	if err != nil {
		errorlog.LogError("swapping classes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
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
