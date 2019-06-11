package api

import (
	"net/http"

	"github.com/MyHomeworkSpace/api-server/util"
	"github.com/labstack/echo"
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
type ClassResponse struct {
	Status  string          `json:"status"`
	Classes []HomeworkClass `json:"classes"`
}
type SingleClassResponse struct {
	Status string        `json:"status"`
	Class  HomeworkClass `json:"class"`
}
type HWInfoResponse struct {
	Status  string `json:"status"`
	HWItems int    `json:"hwItems"`
}

func routeClassesGet(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, teacher, color, userId FROM classes WHERE userId = ?", c.User.ID)
	if err != nil {
		ErrorLog_LogError("getting class information", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
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
	ec.JSON(http.StatusOK, ClassResponse{"ok", classes})
}

func routeClassesGetID(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, teacher, color, userId FROM classes WHERE id = ? AND userId = ?", ec.Param("id"), c.User.ID)
	if err != nil {
		ErrorLog_LogError("getting class information", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}
	resp := HomeworkClass{-1, "", "", "", -1}
	rows.Scan(&resp.ID, &resp.Name, &resp.Teacher, &resp.Color, &resp.UserID)
	if resp.Color == "" {
		resp.Color = DefaultColors[resp.ID%len(DefaultColors)]
	}
	ec.JSON(http.StatusOK, SingleClassResponse{"ok", resp})
}

func routeClassesHWInfo(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT COUNT(*) FROM homework WHERE classId = ? AND userId = ?", ec.Param("id"), c.User.ID)
	if err != nil {
		ErrorLog_LogError("getting class information", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}
	resp := -1
	rows.Scan(&resp)
	ec.JSON(http.StatusOK, HWInfoResponse{"ok", resp})
}

func routeClassesAdd(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("name") == "" || ec.FormValue("color") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}
	if !util.StringSliceContains(DefaultColors, ec.FormValue("color")) {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}

	_, err := DB.Exec(
		"INSERT INTO classes(name, teacher, color, userId) VALUES(?, ?, ?, ?)",
		ec.FormValue("name"), ec.FormValue("teacher"), ec.FormValue("color"), c.User.ID,
	)
	if err != nil {
		ErrorLog_LogError("adding class", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeClassesEdit(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("id") == "" || ec.FormValue("name") == "" || ec.FormValue("color") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}
	if !util.StringSliceContains(DefaultColors, ec.FormValue("color")) {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}

	// check if you are allowed to edit the given id
	idRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("id"))
	if err != nil {
		ErrorLog_LogError("editing classes", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec(
		"UPDATE classes SET name = ?, teacher = ?, color = ? WHERE id = ?",
		ec.FormValue("name"), ec.FormValue("teacher"), ec.FormValue("color"), ec.FormValue("id"),
	)
	if err != nil {
		ErrorLog_LogError("editing class", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeClassesDelete(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("id") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	// check if you are allowed to delete the given id
	idRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("id"))
	if err != nil {
		ErrorLog_LogError("deleting classes", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer idRows.Close()
	if !idRows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	// use a transaction so that you can't delete just the hw or the class entry - either both or nothing
	tx, err := DB.Begin()

	// delete HW calendar events
	_, err = tx.Exec("DELETE calendar_hwevents FROM calendar_hwevents INNER JOIN homework ON calendar_hwevents.homeworkId = homework.id WHERE homework.classId = ?", ec.FormValue("id"))
	if err != nil {
		ErrorLog_LogError("deleting class", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// delete HW
	_, err = tx.Exec("DELETE FROM homework WHERE classId = ?", ec.FormValue("id"))
	if err != nil {
		ErrorLog_LogError("deleting class", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// delete class
	_, err = tx.Exec("DELETE FROM classes WHERE id = ?", ec.FormValue("id"))
	if err != nil {
		ErrorLog_LogError("deleting class", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// go!
	err = tx.Commit()
	if err != nil {
		ErrorLog_LogError("deleting class", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeClassesSwap(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("id1") == "" || ec.FormValue("id2") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	// check if you are allowed to change id1
	id1Rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("id1"))
	if err != nil {
		ErrorLog_LogError("deleting classes", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer id1Rows.Close()
	if !id1Rows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	// check if you are allowed to change id2
	id2Rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("id2"))
	if err != nil {
		ErrorLog_LogError("swapping classes", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer id2Rows.Close()
	if !id2Rows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	// find the swap id
	// this is a dumb way of doing this and is kind of a race condition
	// but mysql has no better way to swap primary keys
	// hopefully no one adds 100 classes in the time this transaction takes to complete
	swapIdStmt, err := DB.Query("SELECT max(id) + 100 FROM classes")
	if err != nil {
		ErrorLog_LogError("swapping classes", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	swapId := -1
	defer swapIdStmt.Close()
	swapIdStmt.Next()
	swapIdStmt.Scan(&swapId)

	// use a transaction so that you can't delete just the hw or the class entry - either both or nothing
	tx, err := DB.Begin()

	// update class id1 -> tmp
	_, err = tx.Exec("UPDATE classes SET id = ? WHERE id = ?", swapId, ec.FormValue("id1"))
	if err != nil {
		ErrorLog_LogError("swapping classes", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// update class id2 -> id1
	_, err = tx.Exec("UPDATE classes SET id = ? WHERE id = ?", ec.FormValue("id1"), ec.FormValue("id2"))
	if err != nil {
		ErrorLog_LogError("swapping classes", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// update class tmp -> id2
	_, err = tx.Exec("UPDATE classes SET id = ? WHERE id = ?", ec.FormValue("id2"), swapId)
	if err != nil {
		ErrorLog_LogError("swapping classes", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// update homework id1 -> swp
	_, err = tx.Exec("UPDATE homework SET classId = ? WHERE classId = ?", swapId, ec.FormValue("id1"))
	if err != nil {
		ErrorLog_LogError("swapping classes", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// update homework id2 -> id1
	_, err = tx.Exec("UPDATE homework SET classId = ? WHERE classId = ?", ec.FormValue("id1"), ec.FormValue("id2"))
	if err != nil {
		ErrorLog_LogError("swapping classes", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// update homework swp -> id2
	_, err = tx.Exec("UPDATE homework SET classId = ? WHERE classId = ?", ec.FormValue("id2"), swapId)
	if err != nil {
		ErrorLog_LogError("swapping classes", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// go!
	err = tx.Commit()
	if err != nil {
		ErrorLog_LogError("deleting class", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}
