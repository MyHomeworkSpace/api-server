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
	ID    int    `json:"id"`
	Date  string `json:"date"`
	Text  string `json:"text"`
	Grade int    `json:"grade"`
	Type  int    `json:"type"`
}

type PlannerFriday struct {
	ID    int    `json:"id"`
	Date  string `json:"date"`
	Index int    `json:"index"`
}

// responses
type MultiPlannerAnnouncementResponse struct {
	Status        string                `json:"status"`
	Announcements []PlannerAnnouncement `json:"announcements"`
}

type PlannerFridayResponse struct {
	Status string        `json:"status"`
	Friday PlannerFriday `json:"friday"`
}

type PlannerWeekInfoResponse struct {
	Status        string                `json:"status"`
	Announcements []PlannerAnnouncement `json:"announcements"`
	Friday        PlannerFriday         `json:"friday"`
}

func InitPlannerAPI(e *echo.Echo) {
	e.GET("/planner/getWeekInfo/:date", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		startDate, err := time.Parse("2006-01-02", c.Param("date"))
		if err != nil {
			jsonResp := ErrorResponse{"error", "Invalid date."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}
		endDate := startDate.Add(time.Hour * 24 * 7)

		user, err := Data_GetUserByID(GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting planner week information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		grade, err := Data_GetUserGrade(user)
		if err != nil {
			log.Println("Error while getting planner week information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		announcementsGroups := Data_GetGradeAnnouncementGroups(grade)
		announcementsGroupsSQL := Data_GetAnnouncementGroupSQL(announcementsGroups)

		announcementRows, err := DB.Query("SELECT id, date, text, grade, `type` FROM announcements WHERE date >= ? AND date < ? AND ("+announcementsGroupsSQL+")", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		if err != nil {
			log.Println("Error while getting announcement information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer announcementRows.Close()
		announcements := []PlannerAnnouncement{}
		for announcementRows.Next() {
			resp := PlannerAnnouncement{-1, "", "", -1, -1}
			announcementRows.Scan(&resp.ID, &resp.Date, &resp.Text, &resp.Grade, &resp.Type)
			announcements = append(announcements, resp)
		}

		fridayDate := startDate.Add(time.Hour * 24 * 4)

		fridayRows, err := DB.Query("SELECT * FROM fridays WHERE date = ?", fridayDate)
		if err != nil {
			log.Println("Error while getting friday information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer fridayRows.Close()
		friday := PlannerFriday{-1, "", -1}
		if fridayRows.Next() {
			fridayRows.Scan(&friday.ID, &friday.Date, &friday.Index)
		}

		return c.JSON(http.StatusOK, PlannerWeekInfoResponse{"ok", announcements, friday})
	})

	e.GET("/planner/announcements/getWeek/:date", func(c echo.Context) error {
		startDate, err := time.Parse("2006-01-02", c.Param("date"))
		if err != nil {
			jsonResp := ErrorResponse{"error", "Invalid date."}
			return c.JSON(http.StatusBadRequest, jsonResp)
		}
		endDate := startDate.Add(time.Hour * 24 * 7)
		rows, err := DB.Query("SELECT id, date, text, grade, `type` FROM announcements WHERE date >= ? AND date < ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
		if err != nil {
			log.Println("Error while getting announcement information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()
		announcements := []PlannerAnnouncement{}
		for rows.Next() {
			resp := PlannerAnnouncement{-1, "", "", -1, -1}
			rows.Scan(&resp.ID, &resp.Date, &resp.Text, &resp.Grade, &resp.Type)
			announcements = append(announcements, resp)
		}
		jsonResp := MultiPlannerAnnouncementResponse{"ok", announcements}
		return c.JSON(http.StatusOK, jsonResp)
	})

	e.GET("/planner/fridays/get/:date", func(c echo.Context) error {
		rows, err := DB.Query("SELECT * FROM fridays WHERE date = ?", c.Param("date"))
		if err != nil {
			log.Println("Error while getting friday information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()
		if rows.Next() {
			resp := PlannerFriday{-1, "", -1}
			rows.Scan(&resp.ID, &resp.Date, &resp.Index)
			jsonResp := PlannerFridayResponse{"ok", resp}
			return c.JSON(http.StatusOK, jsonResp)
		} else {
			jsonResp := StatusResponse{"ok"}
			return c.JSON(http.StatusOK, jsonResp)
		}
	})
}
