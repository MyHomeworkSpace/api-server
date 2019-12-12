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
	applicationRows, err := DB.Query("SELECT id, name, authorName, callbackUrl FROM applications WHERE clientId = ?", ec.FormValue("clientId"))
	if err != nil {
		errorlog.LogError("completing application auth", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer applicationRows.Close()

	if !applicationRows.Next() {
		ec.JSON(http.StatusNotFound, errorResponse{"error", "not_found"})
		return
	}

	application := data.Application{}
	applicationRows.Scan(&application.ID, &application.Name, &application.AuthorName, &application.CallbackURL)

	// check if we've already authorized this application
	tokenCheckRows, err := DB.Query("SELECT token FROM application_authorizations WHERE applicationId = ? AND userId = ?", application.ID, c.User.ID)
	if err != nil {
		errorlog.LogError("completing application auth", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer tokenCheckRows.Close()

	if tokenCheckRows.Next() {
		// if we have, just return that token
		token := ""
		tokenCheckRows.Scan(&token)
		ec.JSON(http.StatusOK, applicationTokenResponse{"ok", token})
		return
	}

	// add the new authorization
	token, err := util.GenerateRandomString(56)
	if err != nil {
		errorlog.LogError("generating application token", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	_, err = DB.Exec("INSERT INTO application_authorizations(applicationId, userId, token) VALUES(?, ?, ?)", application.ID, c.User.ID, token)
	if err != nil {
		errorlog.LogError("authorizing application", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, applicationTokenResponse{"ok", token})
}

func routeApplicationGet(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, authorName, callbackUrl FROM applications WHERE clientId = ?", ec.Param("id"))
	if err != nil {
		errorlog.LogError("getting application information", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		ec.JSON(http.StatusNotFound, errorResponse{"error", "not_found"})
		return
	}

	resp := data.Application{}
	rows.Scan(&resp.ID, &resp.Name, &resp.AuthorName, &resp.CallbackURL)

	ec.JSON(http.StatusOK, singleApplicationResponse{"ok", resp})
}

func routeApplicationGetAuthorizations(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT application_authorizations.id, applications.id, applications.name, applications.authorName FROM application_authorizations INNER JOIN applications ON application_authorizations.applicationId = applications.id WHERE application_authorizations.userId = ?", c.User.ID)
	if err != nil {
		errorlog.LogError("getting authorizations", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	authorizations := []data.ApplicationAuthorization{}
	for rows.Next() {
		resp := data.ApplicationAuthorization{}
		rows.Scan(&resp.ID, &resp.ApplicationID, &resp.Name, &resp.AuthorName)
		authorizations = append(authorizations, resp)
	}
	ec.JSON(http.StatusOK, applicationAuthorizationsResponse{"ok", authorizations})
}

func routeApplicationRequestAuth(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	state := ec.FormValue("state")
	if state == "" {
		ec.Redirect(http.StatusFound, config.GetCurrent().Server.AppURLBase+"applicationAuth:"+ec.Param("id"))
	} else {
		ec.Redirect(http.StatusFound, config.GetCurrent().Server.AppURLBase+"applicationAuth:"+ec.Param("id")+":"+base64.URLEncoding.EncodeToString([]byte(ec.FormValue("state"))))
	}
}

func routeApplicationRevokeAuth(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	// find the authorization
	rows, err := DB.Query("SELECT userId FROM application_authorizations WHERE id = ?", ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("revoking authorization", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		ec.JSON(http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	userID := -1
	rows.Scan(&userID)

	if c.User.ID != userID {
		ec.JSON(http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	// delete the authorization
	_, err = DB.Exec("DELETE FROM application_authorizations WHERE id = ?", ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("revoking authorization", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeApplicationRevokeSelf(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if !HasAuthToken(&ec) {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "bad_request"})
		return
	}

	// delete the authorization
	_, err := DB.Exec("DELETE FROM application_authorizations WHERE token = ?", GetAuthToken(&ec))
	if err != nil {
		errorlog.LogError("revoking authorization", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeApplicationManageCreate(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	// generate client id
	clientID, err := util.GenerateRandomString(42)
	if err != nil {
		errorlog.LogError("creating application", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// get author name
	rows, err := DB.Query("SELECT name FROM users WHERE id = ?", c.User.ID)
	if err != nil {
		errorlog.LogError("creating application", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		errorlog.LogError("creating application", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	authorName := ""
	rows.Scan(&authorName)

	// actually create the application
	_, err = DB.Exec("INSERT INTO applications(name, userId, authorName, clientId, callbackUrl) VALUES('New application', ?, ?, ?, '')", c.User.ID, authorName, clientID)
	if err != nil {
		errorlog.LogError("creating application", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeApplicationManageGetAll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT id, name, authorName, clientId, callbackUrl FROM applications WHERE userId = ?", c.User.ID)
	if err != nil {
		errorlog.LogError("getting user applications", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	apps := []data.Application{}
	for rows.Next() {
		resp := data.Application{}
		rows.Scan(&resp.ID, &resp.Name, &resp.AuthorName, &resp.ClientID, &resp.CallbackURL)
		apps = append(apps, resp)
	}

	ec.JSON(http.StatusOK, multipleApplicationsResponse{"ok", apps})
}

func routeApplicationManageUpdate(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("id") == "" || ec.FormValue("name") == "" || ec.FormValue("callbackUrl") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	// check that you can actually edit the application
	rows, err := DB.Query("SELECT id FROM applications WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("updating application", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		ec.JSON(http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	// update the application
	_, err = DB.Exec("UPDATE applications SET name = ?, callbackUrl = ? WHERE id = ?", ec.FormValue("name"), ec.FormValue("callbackUrl"), ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("updating application", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}

func routeApplicationManageDelete(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("id") == "" {
		ec.JSON(http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	// check that you can actually edit the application
	rows, err := DB.Query("SELECT id FROM applications WHERE userId = ? AND id = ?", c.User.ID, ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting application", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		ec.JSON(http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	tx, err := DB.Begin()

	// delete authorizations
	_, err = tx.Exec("DELETE FROM application_authorizations WHERE applicationId = ?", ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting application", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// delete applications
	_, err = tx.Exec("DELETE FROM applications WHERE id = ?", ec.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting application", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	// go!
	err = tx.Commit()
	if err != nil {
		errorlog.LogError("deleting application", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}
