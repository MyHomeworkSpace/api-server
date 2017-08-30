package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
)

// structs for data
type Homework struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Due      string `json:"due"`
	Desc     string `json:"desc"`
	Complete int    `json:"complete"`
	ClassID  int    `json:"classId"`
	UserID   int    `json:"userId"`
}

// responses
type HomeworkResponse struct {
	Status   string     `json:"status"`
	Homework []Homework `json:"homework"`
}
type SingleHomeworkResponse struct {
	Status   string   `json:"status"`
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
	e.GET("/homework/getHWView", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (`due` > (NOW() - INTERVAL 2 DAY) OR `complete` != '1') ORDER BY `due` ASC", GetSessionUserID(&c))
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

	e.GET("/homework/getWeek/:monday", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}

		startDate, err := time.Parse("2006-01-02", c.Param("monday"))
		if err != nil {
			jsonResp := ErrorResponse{"error", "Invalid date."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}
		endDate := startDate.Add(time.Hour * 24 * 7)

		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (due >= ? and due < ?)", GetSessionUserID(&c), startDate, endDate)
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

	e.GET("/homework/search", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if c.FormValue("q") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		query := c.FormValue("q")

		// sanitize the query, see https://githubengineering.com/like-injection/ for details
		query = strings.Replace(query, "\\", "\\\\", -1)
		query = strings.Replace(query, "%", "\\%", -1)
		query = strings.Replace(query, "_", "\\_", -1)
		query = "%" + query + "%"

		rows, err := DB.Query("SELECT id, name, `due`, `desc`, `complete`, classId, userId FROM homework WHERE userId = ? AND (`name` LIKE ? OR `desc` LIKE ?)", GetSessionUserID(&c), query, query)
		if err != nil {
			log.Println("Error while getting homework search results:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer rows.Close()

		homework := []Homework{}
		for rows.Next() {
			resp := Homework{-1, "", "", "", -1, -1, -1}
			rows.Scan(&resp.ID, &resp.Name, &resp.Due, &resp.Desc, &resp.Complete, &resp.ClassID, &resp.UserID)
			homework = append(homework, resp)
		}
		return c.JSON(http.StatusOK, HomeworkResponse{"ok", homework})
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

	e.POST("/homework/delete", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		if c.FormValue("id") == "" {
			jsonResp := ErrorResponse{"error", "Missing ID parameter."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		// check if you are allowed to edit the given id
		idRows, err := DB.Query("SELECT id FROM homework WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting homework: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer idRows.Close()
		if !idRows.Next() {
			jsonResp := ErrorResponse{"error", "Invalid ID."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}

		deleteTx, err := DB.Begin()

		// delete the homework records
		_, err = deleteTx.Exec("DELETE FROM homework WHERE id = ?", c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting homework: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		// delete any associated calendar events
		_, err = deleteTx.Exec("DELETE FROM calendar_hwevents WHERE homeworkId = ?", c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting homework: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		err = deleteTx.Commit()
		if err != nil {
			log.Println("Error while deleting homework: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
