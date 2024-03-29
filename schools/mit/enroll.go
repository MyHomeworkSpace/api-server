package mit

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/MyHomeworkSpace/api-server/config"

	"github.com/thatoddmailbox/touchstone-client/touchstone"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

func clearUserData(tx *sql.Tx, user *data.User) error {
	// clear away anything that is in the db
	_, err := tx.Exec("DELETE FROM mit_classes WHERE userId = ?", user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *school) Enroll(tx *sql.Tx, user *data.User, params map[string]interface{}) (map[string]interface{}, error) {
	/*
	 * mit enrollment stages:
	 * 0 = enter kerberos username/password
	 * 1 = duo
	 */

	stageInterface, hasStage := params["stage"]
	if !hasStage {
		return nil, data.SchoolError{Code: "missing_params"}
	}

	stageFloat, ok := stageInterface.(float64)
	if !ok {
		return nil, data.SchoolError{Code: "invalid_params"}
	}

	stage := int(stageFloat)

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

	if stage < 0 || stage > 2 {
		// ???
		return nil, data.SchoolError{Code: "invalid_params"}
	}

	cookieJar, _ := cookiejar.New(nil)
	httpClient := &http.Client{
		Jar: cookieJar,
	}

	if config.GetCurrent().MIT.AuthProxyURL != "" {
		authProxyURL, err := url.Parse(config.GetCurrent().MIT.AuthProxyURL)
		if err != nil {
			return nil, err
		}

		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(authProxyURL),
			ProxyConnectHeader: http.Header{
				"X-MHS-Auth": []string{config.GetCurrent().MIT.ProxyToken},
			},
		}
	}

	tsClient := touchstone.NewClientWithHTTPClient(httpClient)

	challenge, err := tsClient.BeginUsernamePasswordAuth(username, password)
	if err != nil {
		if err == touchstone.ErrBadCreds {
			return nil, data.SchoolError{Code: "creds_incorrect"}
		}

		return nil, err
	}

	if stage == 0 {
		return nil, data.DetailedSchoolError{
			Code: "more_info",
			Details: map[string]interface{}{
				"duo": map[string]interface{}{
					"devices": challenge.Devices,
					"methods": challenge.Methods,
				},
			},
		}
	}

	duoMethodIndexInterface, hasStage := params["duoMethodIndex"]
	if !hasStage {
		return nil, data.SchoolError{Code: "missing_params"}
	}

	duoMethodIndexString, ok := duoMethodIndexInterface.(float64)
	if !ok {
		return nil, data.SchoolError{Code: "invalid_params"}
	}

	duoMethodIndex := int(duoMethodIndexString)

	if duoMethodIndex < 0 {
		return nil, data.SchoolError{Code: "invalid_params"}
	}

	if duoMethodIndex > len(challenge.Methods)-1 {
		return nil, data.SchoolError{Code: "invalid_params"}
	}

	duoMethod := challenge.Methods[duoMethodIndex]

	result, err := challenge.StartMethod(&duoMethod)
	if err != nil {
		return nil, err
	}

	if result.StatusCode != "pushed" {
		return nil, data.SchoolError{Code: "duo_denied"}
	}

	final, response, err := challenge.WaitForCompletion()
	if err != nil {
		return nil, err
	}

	if response.StatusCode != "allow" {
		return nil, data.SchoolError{Code: "duo_denied"}
	}

	// we made it!
	err = tsClient.CompleteAuthWithDuo(final)
	if err != nil {
		return nil, err
	}

	academicProfile, registration, peInfo, err := fetchDataWithClient(tsClient, username)
	if err != nil {
		return nil, err
	}

	// now we save the user's classes
	// wipe whatever's there
	err = clearUserData(tx, user)
	if err != nil {
		return nil, err
	}

	classInsertStmt, err := tx.Prepare("INSERT INTO mit_classes(subjectID, sectionID, title, units, sections, custom, userID) VALUES(?, ?, ?, ?, '', 0, ?)")
	if err != nil {
		return nil, err
	}
	defer classInsertStmt.Close()
	for _, subject := range registration.StatusOfRegistration.Subjects {
		selection := subject.Selection
		_, err = classInsertStmt.Exec(
			strings.TrimSpace(selection.SubjectID),
			strings.TrimSpace(selection.SectionID),
			strings.TrimSpace(selection.Title),
			selection.Units,
			user.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	var peInfoStringPointer *string

	if peInfo != nil {
		peInfoBytes, err := json.Marshal(peInfo)
		if err != nil {
			return nil, err
		}

		peInfoString := string(peInfoBytes)
		peInfoStringPointer = &peInfoString
	}

	return map[string]interface{}{
		"status":          1,
		"name":            academicProfile.Name,
		"username":        username,
		"year":            academicProfile.Year,
		"mitID":           academicProfile.MITID,
		"load":            registration.StatusOfRegistration.RegistrationLoad,
		"termCode":        registration.StatusOfRegistration.TermCode,
		"termDescription": registration.StatusOfRegistration.TermDescription,
		"peInfo":          peInfoStringPointer,
	}, nil
}

func (s *school) Unenroll(tx *sql.Tx, user *data.User) error {
	return clearUserData(tx, user)
}

func (s *school) NeedsUpdate(db *sql.DB) (bool, error) {
	return (s.importStatus != schools.ImportStatusOK), nil
}
