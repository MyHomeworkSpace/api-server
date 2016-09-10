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
			log.Println("Error while adding section: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})
}
