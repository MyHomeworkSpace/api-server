package api

import (
	"log"
	"net/http"
	"time"

	//"github.com/MyHomeworkSpace/api-server/auth"

	"github.com/labstack/echo"
)

// structs for data
type PlannerAnnouncement struct {
	ID int `json:"id"`
	Date string `json:"date"`
	Text string `json:"text"`
}

type PlannerEvent struct {
	EventID int `json:"eventId"`
	SectionIndex int `json:"sectionIndex"`
	UserID int `json:"userId"`
	SubID int `json:"subId"`
	Date string `json:"date"`
	Text string `json:"text"`
	Done int `json:"done"`
}

type PlannerFriday struct {
	EntryID int `json:"entryId"`
	Date string `json:"date"`
	Index int `json:"index"`
}

type PlannerSection struct {
	SectionGID int `json:"sectionGid"`
	SectionIndex int `json:"sectionIndex"`
	UserID int `json:"userId"`
	Name string `json:"name"`
}

// responses
type PlannerAnnouncementResponse struct {
	Status string `json:"status"`
	Announcement PlannerAnnouncement `json:"announcement"`
}

type MultiPlannerAnnouncementResponse struct {
	Status string `json:"status"`
	Announcements []PlannerAnnouncement `json:"announcements"`
}

type PlannerFridayResponse struct {
	Status string `json:"status"`
	Friday PlannerFriday `json:"friday"`
}

type MultiPlannerSectionResponse struct {
	Status string `json:"status"`
	Sections []PlannerSection `json:"sections"`
}

type PlannerWeekResponse struct {
	Status string `json:"status"`
	StartDate string `json:"startDate"`
	Announcements []PlannerAnnouncement `json:"announcements"`
	Events []PlannerEvent `json:"events"`
}

func InitPlannerAPI(e *echo.Echo) {
	e.GET("/planner/announcements/get/:date", func(c echo.Context) error {
		rows, err := DB.Query("SELECT id, date, text FROM planner_announcements WHERE date = ?", c.Param("date"))
		if err != nil {
			log.Println("Error while getting announcement information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()
		if rows.Next() {
			resp := PlannerAnnouncement{-1, "", ""}
			rows.Scan(&resp.ID, &resp.Date, &resp.Text)
			jsonResp := PlannerAnnouncementResponse{"ok", resp}
			return c.JSON(http.StatusOK, jsonResp)
		} else {
			jsonResp := StatusResponse{"ok"}
			return c.JSON(http.StatusOK, jsonResp)
		}
	})

	e.GET("/planner/announcements/getWeek/:date", func(c echo.Context) error {
		startDate, err := time.Parse("2006-01-02", c.Param("date"))
		if err != nil {
			jsonResp := ErrorResponse{"error", "Invalid date."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}
		endDate := startDate.Add(time.Hour * 24 * 7)
		rows, err := DB.Query("SELECT id, date, text FROM planner_announcements WHERE date >= ? AND date <= ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		if err != nil {
			log.Println("Error while getting announcement information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()
		announcements := []PlannerAnnouncement{}
		for rows.Next() {
			resp := PlannerAnnouncement{-1, "", ""}
			rows.Scan(&resp.ID, &resp.Date, &resp.Text)
			announcements = append(announcements, resp)
		}
		jsonResp := MultiPlannerAnnouncementResponse{"ok", announcements}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.GET("/planner/fridays/get/:date", func(c echo.Context) error {
		rows, err := DB.Query("SELECT * FROM planner_fridays WHERE date = ?", c.Param("date"))
		if err != nil {
			log.Println("Error while getting friday information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()
		if rows.Next() {
			resp := PlannerFriday{-1, "", -1}
			rows.Scan(&resp.EntryID, &resp.Date, &resp.Index)
			jsonResp := PlannerFridayResponse{"ok", resp}
			return c.JSON(http.StatusOK, jsonResp)
		} else {
			jsonResp := StatusResponse{"ok"}
			return c.JSON(http.StatusOK, jsonResp)
		}
	})

	// TODO: other /events/get* things? not sure if they're needed really...

	e.GET("/planner/events/getWholeWeek/:date", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		startDate, err := time.Parse("2006-01-02", c.Param("date"))
		if err != nil {
			jsonResp := ErrorResponse{"error", "Invalid date."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}
		endDate := startDate.Add(time.Hour * 24 * 7)

		announcements := []PlannerAnnouncement{}
		events := []PlannerEvent{}

		announcementRows, err := DB.Query("SELECT * FROM planner_announcements WHERE date >= ? AND date <= ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		if err != nil {
			log.Println("Error while getting week information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer announcementRows.Close()
		for announcementRows.Next() {
			resp := PlannerAnnouncement{-1, "", ""}
			announcementRows.Scan(&resp.ID, &resp.Date, &resp.Text)
			announcements = append(announcements, resp)
		}

		eventRows, err := DB.Query("SELECT * FROM planner_events WHERE userId = ? AND date >= ? AND date <= ?", GetSessionUserID(&c), startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		if err != nil {
			log.Println("Error while getting week information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer eventRows.Close()
		for eventRows.Next() {
			resp := PlannerEvent{-1, -1, -1, -1, "", "", -1}
			eventRows.Scan(&resp.EventID, &resp.SectionIndex, &resp.UserID, &resp.SubID, &resp.Date, &resp.Text, &resp.Done)
			events = append(events, resp)
		}

		jsonResp := PlannerWeekResponse{"ok", startDate.Format("2006-01-02"), announcements, events}
		return c.JSON(http.StatusOK, jsonResp)
	})

	// /events/post

	// /events/purgeLine

	e.GET("/planner/sections/get", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		rows, err := DB.Query("SELECT sectionGid, sectionIndex, userId, name FROM planner_sections WHERE userId = ?", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting section information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()
		sections := []PlannerSection{}
		for rows.Next() {
			resp := PlannerSection{-1, -1, -1, ""}
			rows.Scan(&resp.SectionGID, &resp.SectionIndex, &resp.UserID, &resp.Name)
			sections = append(sections, resp)
		}
		jsonResp := MultiPlannerSectionResponse{"ok", sections}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.POST("/planner/sections/add", func(c echo.Context) error {
		if c.FormValue("name") == "" {
			jsonResp := ErrorResponse{"error", "Missing required parameters."}
			return c.JSON(http.StatusUnprocessableEntity, jsonResp)
		}
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		rows, err := DB.Query("SELECT sectionIndex FROM planner_sections WHERE userId = ? ORDER BY sectionIndex DESC", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while adding section: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()
		rows.Next()
		newIndex := -1
		rows.Scan(&newIndex)
		newIndex++

		// add the new section
		stmt, err := DB.Prepare("INSERT INTO planner_sections(sectionIndex, userId, name) VALUES(?, ?, ?)")
		if err != nil {
			log.Println("Error while adding section: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		_, err = stmt.Exec(newIndex, GetSessionUserID(&c), c.FormValue("name"))
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
