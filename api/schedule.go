package api

import (
	"net/http"

	"github.com/labstack/echo"
)

type FacultyListResponse struct {
	Status      string        `json:"status"`
	FacultyList []FacultyInfo `json:"faculty"`
}

type FacultyPeriodResponse struct {
	Status  string          `json:"status"`
	Periods []FacultyPeriod `json:"periods"`
}

type FacultyPeriod struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	SectionID int    `json:"sectionId"`
	Room      string `json:"room"`
	Block     string `json:"block"`
	DayNumber int    `json:"dayNumber"`
	Grade     int    `json:"grade"`
	Term      int    `json:"term"`
	Start     int    `json:"start"`
	End       int    `json:"end"`
	FacultyID int    `json:"facultyId"`
}

func InitScheduleAPI(e *echo.Echo) {
	e.GET("/schedule/getFacultyList", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		rows, err := DB.Query("SELECT bbId, firstName, lastName, largeFileName, department, grades FROM faculty ORDER BY lastName ASC")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		users := []FacultyInfo{}
		for rows.Next() {
			user := FacultyInfo{}
			rows.Scan(&user.BlackbaudUserID, &user.FirstName, &user.LastName, &user.LargeFileName, &user.DepartmentDisplay, &user.GradeNumericDisplay)
			users = append(users, user)
		}

		return c.JSON(http.StatusOK, FacultyListResponse{"ok", users})
	})

	e.GET("/schedule/:sectionId/periods", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		sectionId := c.Param("sectionId")

		rows, err := DB.Query("SELECT id, name, sectionId, room, block, dayNumber, grade, term, start, end, facultyId FROM faculty_periods WHERE sectionId = ?", sectionId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		periods := []FacultyPeriod{}
		for rows.Next() {
			period := FacultyPeriod{}
			rows.Scan(&period.ID, &period.Name, &period.SectionID, &period.Room, &period.Block, &period.DayNumber, &period.Grade, &period.Term, &period.Start, &period.End, &period.FacultyID)
			periods = append(periods, period)
		}

		return c.JSON(http.StatusOK, FacultyPeriodResponse{"ok", periods})
	})
}
