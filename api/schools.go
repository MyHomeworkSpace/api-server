package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/labstack/echo"
)

var MainRegistry data.SchoolRegistry

type detailedSchoolErrorResponse struct {
	Status  string                 `json:"status"`
	Error   string                 `json:"error"`
	Details map[string]interface{} `json:"details"`
}

type schoolResultResponse struct {
	Status string             `json:"status"`
	School *data.SchoolResult `json:"school"`
}

func routeSchoolsEnroll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("school") == "" || ec.FormValue("data") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	var err error
	reenroll := false
	if ec.FormValue("reenroll") != "" {
		reenroll, err = strconv.ParseBool(ec.FormValue("reenroll"))
		if err != nil {
			ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
			return
		}
	}

	// find school
	school, err := MainRegistry.GetSchoolByID(ec.FormValue("school"))
	if err == data.ErrNotFound {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	} else if err != nil {
		errorlog.LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// check we're not already enrolled
	enrolled := false
	for _, userSchool := range c.User.Schools {
		if userSchool.SchoolID == school.ID() {
			// we are
			enrolled = true
		}
	}

	if reenroll {
		if !enrolled {
			ec.JSON(http.StatusBadRequest, errorResponse{"error", "not_enrolled"})
			return
		}
	} else {
		if enrolled {
			ec.JSON(http.StatusBadRequest, errorResponse{"error", "already_enrolled"})
			return
		}
	}

	// parse data
	enrollDataString := ec.FormValue("data")
	enrollData := map[string]interface{}{}

	err = json.Unmarshal([]byte(enrollDataString), &enrollData)
	if err != nil {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// actually do it

	// new transaction
	tx, err := DB.Begin()
	if err != nil {
		errorlog.LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// clear any existing school record
	_, err = tx.Exec("DELETE FROM schools WHERE schoolId = ? AND userId = ?", school.ID(), c.User.ID)
	if err != nil {
		tx.Rollback()

		errorlog.LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	if reenroll {
		err := school.Unenroll(tx, c.User)
		if err != nil {
			tx.Rollback()

			errorlog.LogError("enrolling in school - unenrolling before enroll", err)
			ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}
	}

	// enroll
	result, err := school.Enroll(tx, c.User, enrollData)
	if err != nil {
		tx.Rollback()

		detailedSchoolError, ok := err.(data.DetailedSchoolError)
		if ok {
			// it wants to report something
			ec.JSON(http.StatusOK, detailedSchoolErrorResponse{"error", detailedSchoolError.Code, detailedSchoolError.Details})
			return
		}

		schoolError, ok := err.(data.SchoolError)
		if ok {
			// it wants to report an error code
			ec.JSON(http.StatusOK, errorResponse{"error", schoolError.Code})
			return
		}

		// server error
		errorlog.LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	resultDataBytes, err := json.Marshal(result)
	if err != nil {
		tx.Rollback()

		errorlog.LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	resultDataString := string(resultDataBytes)

	// save the new data
	_, err = tx.Exec("INSERT INTO schools(schoolId, data, userId) VALUES(?, ?, ?)", school.ID(), resultDataString, c.User.ID)
	if err != nil {
		tx.Rollback()

		errorlog.LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// go!
	err = tx.Commit()
	if err != nil {
		errorlog.LogError("enrolling in school", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeSchoolsLookup(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("email") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	email := ec.FormValue("email")

	emailParts := strings.Split(email, "@")

	if len(emailParts) != 2 {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	domain := strings.ToLower(strings.TrimSpace(emailParts[1]))

	school, err := MainRegistry.GetSchoolByEmailDomain(domain)

	if err == data.ErrNotFound {
		ec.JSON(http.StatusOK, schoolResultResponse{"ok", nil})
		return
	} else if err != nil {
		errorlog.LogError("looking up school by email domain", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	schoolResult := data.SchoolResult{
		SchoolID:    school.ID(),
		DisplayName: school.Name(),
		ShortName:   school.ShortName(),
	}

	ec.JSON(http.StatusOK, schoolResultResponse{"ok", &schoolResult})
}

func routeSchoolsSetEnabled(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("school") == "" || ec.FormValue("enabled") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	enabled, err := strconv.ParseBool(ec.FormValue("enabled"))
	if err != nil {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// find school
	school, err := MainRegistry.GetSchoolByID(ec.FormValue("school"))
	if err == data.ErrNotFound {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	} else if err != nil {
		errorlog.LogError("set school's enabled status", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
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
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "not_enrolled"})
		return
	}

	// update its status
	_, err = DB.Exec(
		"UPDATE schools SET enabled = ? WHERE schoolId = ? AND userId = ?",
		enabled,
		school.ID(),
		c.User.ID,
	)
	if err != nil {
		errorlog.LogError("set school's enabled status", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeSchoolsUnenroll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("school") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	// find school
	school, err := MainRegistry.GetSchoolByID(ec.FormValue("school"))
	if err == data.ErrNotFound {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	} else if err != nil {
		errorlog.LogError("unenrolling from school", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
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
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "not_enrolled"})
		return
	}

	// remove it
	tx, err := DB.Begin()
	if err != nil {
		errorlog.LogError("unenrolling from school", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	err = school.Unenroll(tx, c.User)
	if err != nil {
		tx.Rollback()

		errorlog.LogError("unenrolling from school", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	_, err = tx.Exec("DELETE FROM schools WHERE schoolId = ? AND userId = ?", school.ID(), c.User.ID)
	if err != nil {
		tx.Rollback()

		errorlog.LogError("unenrolling from school", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	err = tx.Commit()
	if err != nil {
		errorlog.LogError("unenrolling from school", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}
