package api

import (
	"encoding/json"
	"net/http"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/MyHomeworkSpace/api-server/schools"

	"github.com/julienschmidt/httprouter"
)

type schoolSettingsResponse struct {
	Status   string                 `json:"status"`
	Settings map[string]interface{} `json:"settings"`
}

func routeSchoolsSettingsGet(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("school") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	// find school
	registrySchool, err := MainRegistry.GetSchoolByID(r.FormValue("school"))
	if err == data.ErrNotFound {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	} else if err != nil {
		errorlog.LogError("getting settings for school - "+r.FormValue("school"), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	foundSchool := false
	var userSchool data.School

	// check we're already enrolled
	for _, checkSchool := range c.User.Schools {
		if checkSchool.SchoolID == registrySchool.ID() {
			// we are
			foundSchool = true
			userSchool = checkSchool.School
			break
		}
	}

	if !foundSchool {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "not_enrolled"})
		return
	}

	settings, err := userSchool.GetSettings(DB, c.User)
	if err != nil {
		detailedSchoolError, ok := err.(data.DetailedSchoolError)
		if ok {
			// it wants to report something
			writeJSON(w, http.StatusOK, detailedSchoolErrorResponse{"error", detailedSchoolError.Code, detailedSchoolError.Details})
			return
		}

		schoolError, ok := err.(data.SchoolError)
		if ok {
			// it wants to report an error code
			writeJSON(w, http.StatusOK, errorResponse{"error", schoolError.Code})
			return
		}

		// server error
		errorlog.LogError("getting settings for school - "+registrySchool.ID(), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, schoolSettingsResponse{"ok", settings})
}

func routeSchoolsSettingsSet(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("school") == "" || r.FormValue("settings") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	// find school
	registrySchool, err := MainRegistry.GetSchoolByID(r.FormValue("school"))
	if err == data.ErrNotFound {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	} else if err != nil {
		errorlog.LogError("setting settings for school - "+r.FormValue("school"), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	foundSchool := false
	var userSchool data.School

	// check we're already enrolled
	for _, checkSchool := range c.User.Schools {
		if checkSchool.SchoolID == registrySchool.ID() {
			// we are
			foundSchool = true
			userSchool = checkSchool.School
			break
		}
	}

	if !foundSchool {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "not_enrolled"})
		return
	}

	// try parsing the new settings
	var settings map[string]interface{}
	err = json.Unmarshal([]byte(r.FormValue("settings")), &settings)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	// it worked, so pass them to the school
	tx, updates, err := userSchool.SetSettings(DB, c.User, settings)
	if err != nil {
		if err == schools.ErrUnsupportedOperation {
			// you can't do that
			writeJSON(w, http.StatusOK, errorResponse{"error", "unsupported_operation"})
			return
		}

		detailedSchoolError, ok := err.(data.DetailedSchoolError)
		if ok {
			// it wants to report something
			writeJSON(w, http.StatusOK, detailedSchoolErrorResponse{"error", detailedSchoolError.Code, detailedSchoolError.Details})
			return
		}

		schoolError, ok := err.(data.SchoolError)
		if ok {
			// it wants to report an error code
			writeJSON(w, http.StatusOK, errorResponse{"error", schoolError.Code})
			return
		}

		// server error
		errorlog.LogError("setting settings for school - "+registrySchool.ID(), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// get the school's current data
	schoolData, err := data.GetDataForSchool(&userSchool, c.User)
	if err != nil {
		tx.Rollback()
		errorlog.LogError("setting settings for school - "+registrySchool.ID(), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// update it
	for key, value := range updates {
		schoolData[key] = value
	}

	schoolDataBytes, err := json.Marshal(schoolData)
	if err != nil {
		tx.Rollback()
		errorlog.LogError("setting settings for school - "+registrySchool.ID(), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// save
	_, err = tx.Exec("UPDATE schools SET data = ? WHERE schoolId = ? AND userId = ?", string(schoolDataBytes), userSchool.ID(), c.User.ID)
	if err != nil {
		tx.Rollback()
		errorlog.LogError("setting settings for school - "+registrySchool.ID(), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// commit
	err = tx.Commit()
	if err != nil {
		errorlog.LogError("setting settings for school - "+registrySchool.ID(), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, schoolSettingsResponse{"ok", settings})
}
