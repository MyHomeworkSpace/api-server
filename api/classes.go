package api

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
)

// structs for data
type HomeworkClass struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Teacher string `json:"teacher"`
	UserID int `json:"userId"`
}
// responses
type ClassResponse struct {
	Status string `json:"status"`
	Classes []HomeworkClass `json:"classes"`
}
type SingleClassResponse struct {
	Status string `json:"status"`
	Class HomeworkClass `json:"class"`
}
type HWInfoResponse struct {
	Status string `json:"status"`
	HWItems int `json:"hwItems"`
}

func InitClassesAPI(e *echo.Echo) {
	e.GET("/classes/get", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		rows, err := DB.Query("SELECT id, name, teacher, userId FROM classes WHERE userId = ?", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting class information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()

		classes := []HomeworkClass{}
		for rows.Next() {
			resp := HomeworkClass{-1, "", "", -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Teacher, &resp.UserID)
			classes = append(classes, resp)
		}
		jsonResp := ClassResponse{"ok", classes}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.GET("/classes/get/:id", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		rows, err := DB.Query("SELECT id, name, teacher, userId FROM classes WHERE id = ? AND userId = ?", c.Param("id"), GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting class information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()

		if !rows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid ID."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}
		resp := HomeworkClass{-1, "", "", -1}
		rows.Scan(&resp.ID, &resp.Name, &resp.Teacher, &resp.UserID)
		jsonResp := SingleClassResponse{"ok", resp}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.GET("/classes/hwInfo/:id", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		rows, err := DB.Query("SELECT COUNT(*) FROM homework WHERE classId = ? AND userId = ?", c.Param("id"), GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting class information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()

		if !rows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid ID."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}
		resp := -1
		rows.Scan(&resp)
		jsonResp := HWInfoResponse{"ok", resp}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.POST("/classes/add", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if c.FormValue("name") == "" {
			jsonResp := ErrorResponse{"error", "Name is required."}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}

		stmt, err := DB.Prepare("INSERT INTO classes(name, teacher, userId) VALUES(?, ?, ?)")
		if err != nil {
			log.Println("Error while adding class: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = stmt.Exec(c.FormValue("name"), c.FormValue("teacher"), GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while adding class: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.POST("/classes/edit", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if c.FormValue("id") == "" || c.FormValue("name") == "" {
			jsonResp := ErrorResponse{"error", "Missing required parameters."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// check if you are allowed to edit the given id
		idRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			log.Println("Error while editing classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer idRows.Close()
		if !idRows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid ID."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		stmt, err := DB.Prepare("UPDATE classes SET name = ?, teacher = ? WHERE id = ?")
		if err != nil {
			log.Println("Error while editing class: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = stmt.Exec(c.FormValue("name"), c.FormValue("teacher"), c.FormValue("id"))
		if err != nil {
			log.Println("Error while editing class: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.POST("/classes/delete", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if c.FormValue("id") == "" {
			jsonResp := ErrorResponse{"error", "Missing required parameters."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// check if you are allowed to delete the given id
		idRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer idRows.Close()
		if !idRows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid ID."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// use a transaction so that you can't delete just the hw or the class entry - either both or nothing
		tx, err := DB.Begin()

		// delete HW
		hwStmt, err := tx.Prepare("DELETE FROM homework WHERE classId = ?")
		if err != nil {
			log.Println("Error while deleting class: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = hwStmt.Exec(c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting class: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}

		// delete class
		classStmt, err := tx.Prepare("DELETE FROM classes WHERE id = ?")
		if err != nil {
			log.Println("Error while deleting class: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = classStmt.Exec(c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting class: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}

		// go!
		err = tx.Commit()
		if err != nil {
			log.Println("Error while deleting class: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}

		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.POST("/classes/swap", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if c.FormValue("id1") == "" || c.FormValue("id2") == "" {
			jsonResp := ErrorResponse{"error", "Missing required parameters."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// check if you are allowed to change id1
		id1Rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id1"))
		if err != nil {
			log.Println("Error while deleting classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer id1Rows.Close()
		if !id1Rows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid ID."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// check if you are allowed to change id2
		id2Rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id2"))
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer id2Rows.Close()
		if !id2Rows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid ID."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// find the swap id
		// this is a dumb way of doing this and is kind of a race condition
		// but mysql has no better way to swap primary keys
		// hopefully no one adds 100 classes in the time this transaction takes to complete
		swapIdStmt, err := DB.Query("SELECT max(id) + 100 FROM classes")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
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
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = class1Stmt.Exec(swapId, c.FormValue("id1"))
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}

		// update class id2 -> id1
		class2Stmt, err := tx.Prepare("UPDATE classes SET id = ? WHERE id = ?")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = class2Stmt.Exec(c.FormValue("id1"), c.FormValue("id2"))
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}

		// update class tmp -> id2
		classTmpStmt, err := tx.Prepare("UPDATE classes SET id = ? WHERE id = ?")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = classTmpStmt.Exec(c.FormValue("id2"), swapId)
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}

		// update homework id1 -> swp
		hw1Stmt, err := tx.Prepare("UPDATE homework SET classId = ? WHERE classId = ?")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = hw1Stmt.Exec(swapId, c.FormValue("id1"))
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}

		// update homework id2 -> id1
		hw2Stmt, err := tx.Prepare("UPDATE homework SET classId = ? WHERE classId = ?")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = hw2Stmt.Exec(c.FormValue("id1"), c.FormValue("id2"))
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}

		// update homework swp -> id2
		hwSwapStmt, err := tx.Prepare("UPDATE homework SET classId = ? WHERE classId = ?")
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = hwSwapStmt.Exec(c.FormValue("id2"), swapId)
		if err != nil {
			log.Println("Error while swapping classes: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}

		// go!
		err = tx.Commit()
		if err != nil {
			log.Println("Error while deleting class: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}

		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})
}
