package columbia

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

func clearUserData(tx *sql.Tx, user *data.User) error {
	// clear away anything that is in the db
	_, err := tx.Exec("DELETE FROM columbia_classes WHERE userID = ?", user.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM columbia_meetings WHERE userID = ?", user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *school) Enroll(tx *sql.Tx, user *data.User, params map[string]interface{}) (map[string]interface{}, error) {
	// a student can only be enrolled in either columbia or barnard, but not both
	for _, schoolInfo := range user.Schools {
		if schoolInfo.SchoolID == "columbia" || schoolInfo.SchoolID == "barnard" {
			return nil, data.SchoolError{Code: "already_enrolled"}
		}
	}

	usernameRaw, ok := params["username"]
	passwordRaw, ok2 := params["password"]

	if !ok || !ok2 {
		return nil, data.SchoolError{Code: "missing_params"}
	}

	username, ok := usernameRaw.(string)
	password, ok2 := passwordRaw.(string)

	if !ok || !ok2 || username == "" || password == "" {
		return nil, data.SchoolError{Code: "invalid_params"}
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// request the main page so we get a login form
	loginDoc, err := s.ssolRequest(http.MethodGet, "", nil, jar)
	if err != nil {
		return nil, err
	}

	// get the login form fields
	loginURL, loginData, err := s.parseSSOLLoginForm(loginDoc)
	if err != nil {
		return nil, err
	}

	loginData["jsen"] = []string{"Y"}
	loginData["u_id"] = []string{username}
	loginData["u_pw"] = []string{password}

	// make the login request
	loginResultDoc, err := s.ssolRequest(http.MethodPost, loginURL, loginData, jar)
	if err != nil {
		return nil, err
	}

	// check if it worked
	loginSuccess, loginError := s.parseSSOLLoginResult(loginResultDoc)
	if !loginSuccess {
		if loginError == "You entered an incorrect user identifier or password." {
			return nil, data.SchoolError{Code: "creds_incorrect"}
		} else {
			return nil, fmt.Errorf("unknown SSOL login error: '%s", loginError)
		}
	}

	// at this point, we're logged in
	// we need to now go to the schedule page
	scheduleURL, err := s.findSSOLScheduleURL(loginResultDoc)
	if err != nil {
		return nil, err
	}

	scheduleDoc, err := s.ssolRequest(http.MethodGet, scheduleURL, nil, jar)
	if err != nil {
		return nil, err
	}

	// now we have the schedule page
	// we need to select the "show my name and personal data" option
	viewOptionsURL, viewOptionsParams, err := s.parseSSOLViewOptionsForm(scheduleDoc)
	if err != nil {
		return nil, err
	}

	viewOptionsParams["tran[1]_stdban"] = []string{"1"}

	ssolScheduleInfoDoc, err := s.ssolRequest(http.MethodPost, viewOptionsURL, viewOptionsParams, jar)
	if err != nil {
		return nil, err
	}

	// now we can parse it
	studentName, studentUNI, err := s.parseSSOLSchedulePage(tx, user, ssolScheduleInfoDoc)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status": schools.ImportStatusOK,

		"name":     studentName,
		"username": studentUNI,
	}, nil
}

func (s *school) Unenroll(tx *sql.Tx, user *data.User) error {
	return clearUserData(tx, user)
}

func (s *school) NeedsUpdate(db *sql.DB) (bool, error) {
	return (s.importStatus != schools.ImportStatusOK), nil
}
