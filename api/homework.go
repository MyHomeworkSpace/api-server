package api

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
)

// structs for data
type Homework struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Due string `json:"due"`
	Desc string `json:"desc"`
	Complete int `json:"complete"`
	ClassID int `json:"classId"`
	UserID int `json:"userId"`
}
// responses
type HomeworkResponse struct {
	Status string `json:"status"`
	Homework []Homework `json:"homework"`
}
type SingleHomeworkResponse struct {
	Status string `json:"status"`
	Homework Homework `json:"homework"`
}

func InitHomeworkAPI(e *echo.Echo) {
	e.GET("/homework/get", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? ORDER BY `due` ASC", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting homework information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()

		homework := []Homework{}
		for rows.Next() {
			resp := Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			homework = append(homework, resp)
		}
		jsonResp := HomeworkResponse{"ok", homework}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.GET("/homework/get/:id", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.Param("id"))
		if err != nil {
			log.Println("Error while getting homework information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()

		if !rows.Next() {
			jsonResp := ErrorResponse{"error", "not_found"}
			return c.JSON(http.StatusNotFound, jsonResp)
		}

		resp := Homework{-1, "", "", "", -1, -1, -1}
		rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)

		jsonResp := SingleHomeworkResponse{"ok", resp}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.POST("/homework/add", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if c.FormValue("name") == "" || c.FormValue("due") == "" || c.FormValue("complete") == "" || c.FormValue("classId") == "" {
			jsonResp := ErrorResponse{"error", "Missing required parameters."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// check if you are allowed to add to the given classId
		rows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("classId"))
		if err != nil {
			log.Println("Error while adding homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()
		if !rows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid class."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		stmt, err := DB.Prepare("INSERT INTO homework(name, `due`, `desc`, `complete`, classId, userId) VALUES(?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Println("Error while adding homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = stmt.Exec(c.FormValue("name"), c.FormValue("due"), c.FormValue("desc"), c.FormValue("complete"), c.FormValue("classId"), GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while adding homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.POST("/homework/edit", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if c.FormValue("id") == "" || c.FormValue("name") == "" || c.FormValue("due") == "" || c.FormValue("complete") == "" || c.FormValue("classId") == "" {
			jsonResp := ErrorResponse{"error", "Missing required parameters."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// check if you are allowed to edit the given id
		idRows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			log.Println("Error while editing homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer idRows.Close()
		if !idRows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid ID."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// check if you are allowed to add to the given classId
		classRows, err := DB.Query("SELECT id FROM classes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("classId"))
		if err != nil {
			log.Println("Error while editing homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer classRows.Close()
		if !classRows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid class."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		stmt, err := DB.Prepare("UPDATE homework SET name = ?, `due` = ?, `desc` = ?, `complete` = ?, classId = ? WHERE id = ?")
		if err != nil {
			log.Println("Error while editing homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = stmt.Exec(c.FormValue("name"), c.FormValue("due"), c.FormValue("desc"), c.FormValue("complete"), c.FormValue("classId"), c.FormValue("id"))
		if err != nil {
			log.Println("Error while editing homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		jsonResp := StatusResponse{"ok"}
		return c.JSON(http.StatusOK, jsonResp)
	})
}
