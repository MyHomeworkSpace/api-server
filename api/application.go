package api

import (
	"encoding/base64"
	"net/http"

	"github.com/MyHomeworkSpace/api-server/config"
	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/MyHomeworkSpace/api-server/util"
	"github.com/labstack/echo"
)

type applicationTokenResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}
type applicationAuthorizationsResponse struct {
	Status         string                          `json:"status"`
	Authorizations []data.ApplicationAuthorization `json:"authorizations"`
}
type singleApplicationResponse struct {
	Status      string           `json:"status"`
	Application data.Application `json:"application"`
}
type multipleApplicationsResponse struct {
	Status       string             `json:"status"`
	Applications []data.Application `json:"applications"`
}

func routeApplicationCompleteAuth(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	// get the application
	applicationRows, err := DB.Query("SELECT id, name, authorName, callbackUrl FROM applications WHERE clientId = ?", r.FormValue("clientId"))
	if err != nil {
		errorlog.LogError("completing application auth", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer applicationRows.Close()

	if !applicationRows.Next() {
		writeJSON(w, http.StatusNotFound, errorResponse{"error", "not_found"})
		return
	}

	application := data.Application{}
	applicationRows.Scan(&application.ID, &application.Name, &application.AuthorName, &application.CallbackURL)

	// check if we've already authorized this application
	tokenCheckRows, err := DB.Query("SELECT token FROM application_authorizations WHERE applicationId = ? AND userId = ?", application.ID, c.User.ID)
	if err != nil {
		errorlog.LogError("completing application auth", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer tokenCheckRows.Close()

	if tokenCheckRows.Next() {
		// if we have, just return that token
		token := ""
		tokenCheckRows.Scan(&token)
		writeJSON(w, http.StatusOK, applicationTokenResponse{"ok", token})
		return
	}

	// add the new authorization
	token, err := util.GenerateRandomString(56)
	if err != nil {
		errorlog.LogError("generating application token", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	_, err = DB.Exec("INSERT INTO application_authorizations(applicationId, userId, token) VALUES(?, ?, ?)", application.ID, c.User.ID, token)
	if err != nil {
		errorlog.LogError("authorizing application", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, applicationTokenResponse{"ok", token})
}

func routeApplicationGet(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, authorName, callbackUrl FROM applications WHERE clientId = ?", ec.Param("id"))
	if err != nil {
		errorlog.LogError("getting application information", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		writeJSON(w, http.StatusNotFound, errorResponse{"error", "not_found"})
		return
	}

	resp := data.Application{}
	rows.Scan(&resp.ID, &resp.Name, &resp.AuthorName, &resp.CallbackURL)

	writeJSON(w, http.StatusOK, singleApplicationResponse{"ok", resp})
}

func routeApplicationGetAuthorizations(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT application_authorizations.id, applications.id, applications.name, applications.authorName FROM application_authorizations INNER JOIN applications ON application_authorizations.applicationId = applications.id WHERE application_authorizations.userId = ?", c.User.ID)
	if err != nil {
		errorlog.LogError("getting authorizations", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	authorizations := []data.ApplicationAuthorization{}
	for rows.Next() {
		resp := data.ApplicationAuthorization{}
		rows.Scan(&resp.ID, &resp.ApplicationID, &resp.Name, &resp.AuthorName)
		authorizations = append(authorizations, resp)
	}
	writeJSON(w, http.StatusOK, applicationAuthorizationsResponse{"ok", authorizations})
}

func routeApplicationRequestAuth(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	state := r.FormValue("state")
	if state == "" {
		http.Redirect(w, r, config.GetCurrent().Server.AppURLBase+"applicationAuth:"+ec.Param("id"), http.StatusFound)
	} else {
		http.Redirect(w, r, config.GetCurrent().Server.AppURLBase+"applicationAuth:"+ec.Param("id")+":"+base64.URLEncoding.EncodeToString([]byte(r.FormValue("state"))), http.StatusFound)
	}
}

func routeApplicationRevokeAuth(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	// find the authorization
	rows, err := DB.Query("SELECT userId FROM application_authorizations WHERE id = ?", r.FormValue("id"))
	if err != nil {
		errorlog.LogError("revoking authorization", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	userID := -1
	rows.Scan(&userID)

	if c.User.ID != userID {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	// delete the authorization
	_, err = DB.Exec("DELETE FROM application_authorizations WHERE id = ?", r.FormValue("id"))
	if err != nil {
		errorlog.LogError("revoking authorization", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeApplicationRevokeSelf(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if !HasAuthToken(&ec) {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "bad_request"})
		return
	}

	// delete the authorization
	_, err := DB.Exec("DELETE FROM application_authorizations WHERE token = ?", GetAuthToken(&ec))
	if err != nil {
		errorlog.LogError("revoking authorization", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeApplicationManageCreate(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	// generate client id
	clientID, err := util.GenerateRandomString(42)
	if err != nil {
		errorlog.LogError("creating application", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// get author name
	rows, err := DB.Query("SELECT name FROM users WHERE id = ?", c.User.ID)
	if err != nil {
		errorlog.LogError("creating application", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		errorlog.LogError("creating application", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	authorName := ""
	rows.Scan(&authorName)

	// actually create the application
	_, err = DB.Exec("INSERT INTO applications(name, userId, authorName, clientId, callbackUrl) VALUES('New application', ?, ?, ?, '')", c.User.ID, authorName, clientID)
	if err != nil {
		errorlog.LogError("creating application", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeApplicationManageGetAll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, authorName, clientId, callbackUrl FROM applications WHERE userId = ?", c.User.ID)
	if err != nil {
		errorlog.LogError("getting user applications", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	apps := []data.Application{}
	for rows.Next() {
		resp := data.Application{}
		rows.Scan(&resp.ID, &resp.Name, &resp.AuthorName, &resp.ClientID, &resp.CallbackURL)
		apps = append(apps, resp)
	}

	writeJSON(w, http.StatusOK, multipleApplicationsResponse{"ok", apps})
}

func routeApplicationManageUpdate(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if r.FormValue("id") == "" || r.FormValue("name") == "" || r.FormValue("callbackUrl") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	// check that you can actually edit the application
	rows, err := DB.Query("SELECT id FROM applications WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id"))
	if err != nil {
		errorlog.LogError("updating application", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	// update the application
	_, err = DB.Exec("UPDATE applications SET name = ?, callbackUrl = ? WHERE id = ?", r.FormValue("name"), r.FormValue("callbackUrl"), r.FormValue("id"))
	if err != nil {
		errorlog.LogError("updating application", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routeApplicationManageDelete(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if r.FormValue("id") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	// check that you can actually edit the application
	rows, err := DB.Query("SELECT id FROM applications WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting application", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	tx, err := DB.Begin()

	// delete authorizations
	_, err = tx.Exec("DELETE FROM application_authorizations WHERE applicationId = ?", r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting application", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// delete applications
	_, err = tx.Exec("DELETE FROM applications WHERE id = ?", r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting application", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// go!
	err = tx.Commit()
	if err != nil {
		errorlog.LogError("deleting application", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}
