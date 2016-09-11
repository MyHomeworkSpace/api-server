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
		if c.FormValue("id") == "" || c.FormValue("name") == "" || c.FormValue("teacher") == "" {
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
}
