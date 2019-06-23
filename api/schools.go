package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/labstack/echo"
)

var MainRegistry data.SchoolRegistry

type SchoolResultResponse struct {
	Status string             `json:"status"`
	School *data.SchoolResult `json:"school"`
}

func routeSchoolsEnroll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("school") == "" || ec.FormValue("data") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	// find school
	school, err := MainRegistry.GetSchoolByID(ec.FormValue("school"))
	if err == data.ErrNotFound {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	} else if err != nil {
		ErrorLog_LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// check we're not already enrolled
	for _, userSchool := range c.User.Schools {
		if userSchool.SchoolID == school.ID() {
			// we are
			ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "already_enrolled"})
			return
		}
	}

	// parse data
	enrollDataString := ec.FormValue("data")
	enrollData := map[string]interface{}{}

	err = json.Unmarshal([]byte(enrollDataString), &enrollData)
	if err != nil {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}

	// actually do it

	// new transaction
	tx, err := DB.Begin()
	if err != nil {
		ErrorLog_LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// clear any existing school record
	_, err = tx.Exec("DELETE FROM schools WHERE schoolId = ? AND userId = ?", school.ID(), c.User.ID)
	if err != nil {
		tx.Rollback()

		ErrorLog_LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// enroll
	result, err := school.Enroll(tx, c.User, enrollData)
	if err != nil {
		tx.Rollback()

		schoolError, ok := err.(data.SchoolError)
		if ok {
			// it wants to report an error code
			ec.JSON(http.StatusOK, ErrorResponse{"error", schoolError.Code})
			return
		}

		// server error
		ErrorLog_LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	resultDataBytes, err := json.Marshal(result)
	if err != nil {
		tx.Rollback()

		ErrorLog_LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	resultDataString := string(resultDataBytes)

	// save the new data
	_, err = tx.Exec("INSERT INTO schools(schoolId, data, userId) VALUES(?, ?, ?)", school.ID(), resultDataString, c.User.ID)
	if err != nil {
		tx.Rollback()

		ErrorLog_LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// go!
	err = tx.Commit()
	if err != nil {
		ErrorLog_LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func routeSchoolsLookup(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("email") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	email := ec.FormValue("email")

	emailParts := strings.Split(email, "@")

	if len(emailParts) != 2 {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}

	domain := strings.ToLower(strings.TrimSpace(emailParts[1]))

	school, err := MainRegistry.GetSchoolByEmailDomain(domain)

	if err == data.ErrNotFound {
		ec.JSON(http.StatusOK, SchoolResultResponse{"ok", nil})
		return
	} else if err != nil {
		ErrorLog_LogError("looking up school by email domain", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	schoolResult := data.SchoolResult{
		SchoolID:    school.ID(),
		DisplayName: school.Name(),
	}

	ec.JSON(http.StatusOK, SchoolResultResponse{"ok", &schoolResult})
}

func routeSchoolsUnenroll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("school") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	// find school
	school, err := MainRegistry.GetSchoolByID(ec.FormValue("school"))
	if err == data.ErrNotFound {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	} else if err != nil {
		ErrorLog_LogError("unenrolling from school", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	foundSchool := false

	// check we're already enrolled
	for _, userSchool := range c.User.Schools {
		if userSchool.SchoolID == school.ID() {
			// we are
			foundSchool = true
			break
		}
	}

	if !foundSchool {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "not_enrolled"})
		return
	}

	// remove it
	tx, err := DB.Begin()
	if err != nil {
		ErrorLog_LogError("unenrolling from school", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	err = school.Unenroll(tx, c.User)
	if err != nil {
		tx.Rollback()

		ErrorLog_LogError("unenrolling from school", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	_, err = tx.Exec("DELETE FROM schools WHERE schoolId = ? AND userId = ?", school.ID(), c.User.ID)
	if err != nil {
		tx.Rollback()

		ErrorLog_LogError("unenrolling from school", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	err = tx.Commit()
	if err != nil {
		ErrorLog_LogError("unenrolling from school", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}
