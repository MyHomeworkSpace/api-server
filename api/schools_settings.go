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

func handleSchoolSettingsLookup(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) (bool, data.School) {
	if r.FormValue("school") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return false, nil
	}

	// find school
	registrySchool, err := MainRegistry.GetSchoolByID(r.FormValue("school"))
	if err == data.ErrNotFound {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return false, nil
	} else if err != nil {
		errorlog.LogError("getting settings for school - "+r.FormValue("school"), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return false, nil
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
		return false, nil
	}

	return true, userSchool
}

func routeSchoolsSettingsCallMethod(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	foundSchool, userSchool := handleSchoolSettingsLookup(w, r, p, c)
	if !foundSchool {
		return
	}

	if r.FormValue("methodName") == "" || r.FormValue("methodParams") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	methodName := r.FormValue("methodName")
	methodParams := r.FormValue("methodParams")

	var methodParamsMap map[string]interface{}
	err := json.Unmarshal([]byte(methodParams), &methodParamsMap)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	response, err := userSchool.CallSettingsMethod(DB, c.User, methodName, methodParamsMap)
	if err != nil {
		errorlog.LogError("call settings method for school - "+userSchool.ID(), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func routeSchoolsSettingsGet(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	foundSchool, userSchool := handleSchoolSettingsLookup(w, r, p, c)
	if !foundSchool {
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
		errorlog.LogError("getting settings for school - "+userSchool.ID(), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, schoolSettingsResponse{"ok", settings})
}

func routeSchoolsSettingsSet(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	foundSchool, userSchool := handleSchoolSettingsLookup(w, r, p, c)
	if !foundSchool {
		return
	}

	// try parsing the new settings
	var settings map[string]interface{}
	err := json.Unmarshal([]byte(r.FormValue("settings")), &settings)
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
		errorlog.LogError("setting settings for school - "+userSchool.ID(), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// get the school's current data
	schoolData, err := data.GetDataForSchool(&userSchool, c.User)
	if err != nil {
		tx.Rollback()
		errorlog.LogError("setting settings for school - "+userSchool.ID(), err)
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
		errorlog.LogError("setting settings for school - "+userSchool.ID(), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// save
	_, err = tx.Exec("UPDATE schools SET data = ? WHERE schoolId = ? AND userId = ?", string(schoolDataBytes), userSchool.ID(), c.User.ID)
	if err != nil {
		tx.Rollback()
		errorlog.LogError("setting settings for school - "+userSchool.ID(), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// commit
	err = tx.Commit()
	if err != nil {
		errorlog.LogError("setting settings for school - "+userSchool.ID(), err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, schoolSettingsResponse{"ok", settings})
}
