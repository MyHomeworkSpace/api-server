package api

import (
	"log"
	"net/http"

	"gopkg.in/labstack/echo.v2"
)

// structs for data
type Feedback struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Due string `json:"due"`
	Desc string `json:"desc"`
	Complete int `json:"complete"`
	ClassID int `json:"classId"`
	UserID int `json:"userId"`
}
// responses


func InitFeedbackAPI(e *echo.Echo) {
	e.POST("/feedback/add", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if c.FormValue("type") == "" || c.FormValue("text") == "" {
			jsonResp := ErrorResponse{"error", "Missing required parameters."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		stmt, err := DB.Prepare("INSERT INTO feedback(userId, type, text) VALUES(?, ?, ?)")
		if err != nil {
			log.Println("Error while adding feedback: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = stmt.Exec(GetSessionUserID(&c), c.FormValue("type"), c.FormValue("text"))
		if err != nil {
			log.Println("Error while adding feedback: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})
}
