package api

import (
	"net/http"

	"github.com/labstack/echo"
)

type FacultyListResponse struct {
	Status      string        `json:"status"`
	FacultyList []FacultyInfo `json:"faculty"`
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
}
