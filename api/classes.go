package api

import (
	"log"
	"net/http"

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

func InitClassesAPI(e *echo.Echo) {
	e.GET("/classes/get", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		rows, err := DB.Query("SELECT id, name, teacher, color, userId FROM classes WHERE userId = ?", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting class information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
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
		return c.JSON(http.StatusOK, ClassResponse{"ok", classes})
	})

	e.GET("/classes/get/:id", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		rows, err := DB.Query("SELECT id, name, teacher, color, userId FROM classes WHERE id = ? AND userId = ?", c.Param("id"), GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting class information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer rows.Close()

		if !rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}
		resp := HomeworkClass{-1, "", "", "", -1}
		rows.Scan(&resp.ID, &resp.Name, &resp.Teacher, &resp.Color, &resp.UserID)
		if resp.Color == "" {
			resp.Color = DefaultColors[resp.ID%len(DefaultColors)]
		}
		return c.JSON(http.StatusOK, SingleClassResponse{"ok", resp})
	})

	e.GET("/classes/hwInfo/:id", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		rows, err := DB.Query("SELECT COUNT(*) FROM homework WHERE classId = ? AND userId = ?", c.Param("id"), GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting class information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer rows.Close()

		if !rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}
		resp := -1
		rows.Scan(&resp)
		return c.JSON(http.StatusOK, HWInfoResponse{"ok", resp})
	})

	e.POST("/classes/add", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("name") == "" || c.FormValue("color") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}
		if !Util_StringSliceContains(DefaultColors, c.FormValue("color")) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		stmt, err := DB.Prepare("INSERT INTO classes(name, teacher, color, userId) VALUES(?, ?, ?, ?)")
		if err != nil {
			log.Println("Error while adding class: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		_, err = stmt.Exec(c.FormValue("name"), c.FormValue("teacher"), c.FormValue("color"), GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while adding class: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/classes/edit", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("id") == "" || c.FormValue("name") == "" || c.FormValue("color") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}
		if !Util_StringSliceContains(DefaultColors, c.FormValue("color")) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		// check if you are allowed to edit the given id
		idRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			log.Println("Error while editing classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer idRows.Close()
		if !idRows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		stmt, err := DB.Prepare("UPDATE classes SET name = ?, teacher = ?, color = ? WHERE id = ?")
		if err != nil {
			log.Println("Error while editing class: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		_, err = stmt.Exec(c.FormValue("name"), c.FormValue("teacher"), c.FormValue("color"), c.FormValue("id"))
		if err != nil {
			log.Println("Error while editing class: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/classes/delete", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("id") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		// check if you are allowed to delete the given id
		idRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer idRows.Close()
		if !idRows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		// use a transaction so that you can't delete just the hw or the class entry - either both or nothing
		tx, err := DB.Begin()

		// delete HW calendar events
		_, err = tx.Exec("DELETE calendar_hwevents FROM calendar_hwevents INNER JOIN homework ON calendar_hwevents.homeworkId = homework.id WHERE homework.classId = ?", c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting class: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		// delete HW
		_, err = tx.Exec("DELETE FROM homework WHERE classId = ?", c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting class: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		// delete class
		_, err = tx.Exec("DELETE FROM classes WHERE id = ?", c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting class: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		// go!
		err = tx.Commit()
		if err != nil {
			log.Println("Error while deleting class: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/classes/swap", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("id1") == "" || c.FormValue("id2") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		// check if you are allowed to change id1
		id1Rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id1"))
		if err != nil {
			log.Println("Error while deleting classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer id1Rows.Close()
		if !id1Rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		// check if you are allowed to change id2
		id2Rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id2"))
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer id2Rows.Close()
		if !id2Rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		// find the swap id
		// this is a dumb way of doing this and is kind of a race condition
		// but mysql has no better way to swap primary keys
		// hopefully no one adds 100 classes in the time this transaction takes to complete
		swapIdStmt, err := DB.Query("SELECT max(id) + 100 FROM classes")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		swapId := -1
		defer swapIdStmt.Close()
		swapIdStmt.Next()
		swapIdStmt.Scan(&swapId)

		// use a transaction so that you can't delete just the hw or the class entry - either both or nothing
		tx, err := DB.Begin()

		// update class id1 -> tmp
		class1Stmt, err := tx.Prepare("UPDATE classes SET id = ? WHERE id = ?")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		_, err = class1Stmt.Exec(swapId, c.FormValue("id1"))
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		// update class id2 -> id1
		class2Stmt, err := tx.Prepare("UPDATE classes SET id = ? WHERE id = ?")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		_, err = class2Stmt.Exec(c.FormValue("id1"), c.FormValue("id2"))
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		// update class tmp -> id2
		classTmpStmt, err := tx.Prepare("UPDATE classes SET id = ? WHERE id = ?")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		_, err = classTmpStmt.Exec(c.FormValue("id2"), swapId)
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		// update homework id1 -> swp
		hw1Stmt, err := tx.Prepare("UPDATE homework SET classId = ? WHERE classId = ?")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		_, err = hw1Stmt.Exec(swapId, c.FormValue("id1"))
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		// update homework id2 -> id1
		hw2Stmt, err := tx.Prepare("UPDATE homework SET classId = ? WHERE classId = ?")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		_, err = hw2Stmt.Exec(c.FormValue("id1"), c.FormValue("id2"))
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		// update homework swp -> id2
		hwSwapStmt, err := tx.Prepare("UPDATE homework SET classId = ? WHERE classId = ?")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		_, err = hwSwapStmt.Exec(c.FormValue("id2"), swapId)
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		// go!
		err = tx.Commit()
		if err != nil {
			log.Println("Error while deleting class: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
