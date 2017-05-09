package api

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo"
)

// structs for data
type CalendarEvent struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Start int `json:"start"`
	End int `json:"end"`
	Desc string `json:"desc"`
	UserID int `json:"userId"`
}

// responses
type CalendarEventResponse struct {
	Status string `json:"status"`
	Events []CalendarEvent `json:"events"`
}
type SingleCalendarEventResponse struct {
	Status string `json:"status"`
	Event CalendarEvent `json:"event"`
}

func InitCalendarEventsAPI(e *echo.Echo) {
	e.GET("/calendar/events/getWeek/:monday", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		startDate, err := time.Parse("2006-01-02", c.Param("monday"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}
		endDate := startDate.Add(time.Hour * 24 * 7)

		rows, err := DB.Query("SELECT id, name, `start`, `end`, `desc`, userId FROM calendar_events WHERE userId = ? AND (`end` >= ? OR `start` <= ?)", GetSessionUserID(&c), startDate.Unix(), endDate.Unix())
		if err != nil {
			log.Println("Error while getting calendar events: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		events := []CalendarEvent{}
		for rows.Next() {
			event := CalendarEvent{-1, "", -1, -1, "", -1}
			rows.Scan(&event.ID, &event.Name, &event.Start, &event.End, &event.Desc, &event.UserID)
			events = append(events, event)
		}
		return c.JSON(http.StatusOK, CalendarEventResponse{"ok", events})
	})

	e.POST("/calendar/events/add", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("name") == "" || c.FormValue("start") == "" || c.FormValue("end") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		start, err := strconv.Atoi(c.FormValue("start"))
		end, err2 := strconv.Atoi(c.FormValue("end"))
		if err != nil || err2 != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		stmt, err := DB.Prepare("INSERT INTO calendar_events(name, `start`, `end`, `desc`, userId) VALUES(?, ?, ?, ?, ?)")
		if err != nil {
			log.Println("Error while adding calendar event: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("name"), start, end, c.FormValue("desc"), GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while adding calendar event: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
