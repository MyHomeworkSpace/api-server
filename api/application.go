package api

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
)

type Application struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	AuthorName  string `json:"authorName"`
	ClientID    string `json:"clientId"`
	CallbackURL string `json:"callbackUrl"`
}
type ApplicationAuthorization struct {
	ID            int    `json:"id"`
	ApplicationID int    `json:"applicationId"`
	Name          string `json:"name"`
	AuthorName    string `json:"authorName"`
}

type ApplicationTokenResponse struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}
type ApplicationAuthorizationsResponse struct {
	Status         string                     `json:"status"`
	Authorizations []ApplicationAuthorization `json:"authorizations"`
}
type SingleApplicationResponse struct {
	Status      string      `json:"status"`
	Application Application `json:"application"`
}
type MultipleApplicationsResponse struct {
	Status       string        `json:"status"`
	Applications []Application `json:"applications"`
}

func Route_Application_CompleteAuth(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	// get the application
	applicationRows, err := DB.Query("SELECT id, name, authorName, callbackUrl FROM applications WHERE clientId = ?", ec.FormValue("clientId"))
	if err != nil {
		ErrorLog_LogError("completing application auth", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer applicationRows.Close()

	if !applicationRows.Next() {
		ec.JSON(http.StatusNotFound, ErrorResponse{"error", "not_found"})
		return
	}

	application := Application{-1, "", "", "", ""}
	applicationRows.Scan(&application.ID, &application.Name, &application.AuthorName, &application.CallbackURL)

	// check if we've already authorized this application
	tokenCheckRows, err := DB.Query("SELECT token FROM application_authorizations WHERE applicationId = ? AND userId = ?", application.ID, GetSessionUserID(&ec))
	if err != nil {
		ErrorLog_LogError("completing application auth", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer tokenCheckRows.Close()

	if tokenCheckRows.Next() {
		// if we have, just return that token
		token := ""
		tokenCheckRows.Scan(&token)
		ec.JSON(http.StatusOK, ApplicationTokenResponse{"ok", token})
		return
	}

	// add the new authorization
	token, err := Util_GenerateRandomString(56)
	if err != nil {
		ErrorLog_LogError("generating application token", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	stmt, err := DB.Prepare("INSERT INTO application_authorizations(applicationId, userId, token) VALUES(?, ?, ?)")
	if err != nil {
		ErrorLog_LogError("authorizing application", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	_, err = stmt.Exec(application.ID, GetSessionUserID(&ec), token)
	if err != nil {
		ErrorLog_LogError("authorizing application", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, ApplicationTokenResponse{"ok", token})
}

func Route_Application_Get(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	rows, err := DB.Query("SELECT id, name, authorName, callbackUrl FROM applications WHERE clientId = ?", ec.Param("id"))
	if err != nil {
		ErrorLog_LogError("getting application information", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		ec.JSON(http.StatusNotFound, ErrorResponse{"error", "not_found"})
		return
	}

	resp := Application{-1, "", "", "", ""}
	rows.Scan(&resp.ID, &resp.Name, &resp.AuthorName, &resp.CallbackURL)

	ec.JSON(http.StatusOK, SingleApplicationResponse{"ok", resp})
}

func Route_Application_GetAuthorizations(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}
	rows, err := DB.Query("SELECT application_authorizations.id, applications.id, applications.name, applications.authorName FROM application_authorizations INNER JOIN applications ON application_authorizations.applicationId = applications.id WHERE application_authorizations.userId = ?", GetSessionUserID(&ec))
	if err != nil {
		ErrorLog_LogError("getting authorizations", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	authorizations := []ApplicationAuthorization{}
	for rows.Next() {
		resp := ApplicationAuthorization{-1, -1, "", ""}
		rows.Scan(&resp.ID, &resp.ApplicationID, &resp.Name, &resp.AuthorName)
		authorizations = append(authorizations, resp)
	}
	ec.JSON(http.StatusOK, ApplicationAuthorizationsResponse{"ok", authorizations})
}

func Route_Application_RequestAuth(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	state := ec.FormValue("state")
	if state == "" {
		ec.Redirect(http.StatusFound, AuthURLBase+"?id="+ec.Param("id"))
	} else {
		ec.Redirect(http.StatusFound, AuthURLBase+"?id="+ec.Param("id")+"&state="+ec.FormValue("state"))
	}
}

func Route_Application_RevokeAuth(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	// find the authorization
	rows, err := DB.Query("SELECT userId FROM application_authorizations WHERE id = ?", ec.FormValue("id"))
	if err != nil {
		ErrorLog_LogError("revoking authorization", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	userId := -1
	rows.Scan(&userId)

	if GetSessionUserID(&ec) != userId {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	// delete the authorization
	deleteStmt, err := DB.Prepare("DELETE FROM application_authorizations WHERE id = ?")
	if err != nil {
		ErrorLog_LogError("revoking authorization", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer deleteStmt.Close()
	_, err = deleteStmt.Exec(ec.FormValue("id"))
	if err != nil {
		ErrorLog_LogError("revoking authorization", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func Route_Application_RevokeSelf(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	if !HasAuthToken(&ec) {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "bad_request"})
		return
	}

	// delete the authorization
	deleteStmt, err := DB.Prepare("DELETE FROM application_authorizations WHERE token = ?")
	if err != nil {
		ErrorLog_LogError("revoking authorization", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer deleteStmt.Close()
	_, err = deleteStmt.Exec(GetAuthToken(&ec))
	if err != nil {
		ErrorLog_LogError("revoking authorization", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func Route_Application_Manage_Create(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	// generate client id
	clientId, err := Util_GenerateRandomString(42)
	if err != nil {
		log.Println("Error while creating application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// get author name
	rows, err := DB.Query("SELECT name FROM users WHERE id = ?", GetSessionUserID(&ec))
	if err != nil {
		log.Println("Error while creating application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		log.Println("Error while creating application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	authorName := ""
	rows.Scan(&authorName)

	// actually create the application
	stmt, err := DB.Prepare("INSERT INTO applications(name, userId, authorName, clientId, callbackUrl) VALUES('New application', ?, ?, ?, '')")
	if err != nil {
		log.Println("Error while creating application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	_, err = stmt.Exec(GetSessionUserID(&ec), authorName, clientId)
	if err != nil {
		log.Println("Error while creating application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func Route_Application_Manage_GetAll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	rows, err := DB.Query("SELECT id, name, authorName, clientId, callbackUrl FROM applications WHERE userId = ?", GetSessionUserID(&ec))
	if err != nil {
		log.Println("Error while getting user applications: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	apps := []Application{}
	for rows.Next() {
		resp := Application{-1, "", "", "", ""}
		rows.Scan(&resp.ID, &resp.Name, &resp.AuthorName, &resp.ClientID, &resp.CallbackURL)
		apps = append(apps, resp)
	}

	ec.JSON(http.StatusOK, MultipleApplicationsResponse{"ok", apps})
}

func Route_Application_Manage_Update(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	if ec.FormValue("id") == "" || ec.FormValue("name") == "" || ec.FormValue("callbackUrl") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	// check that you can actually edit the application
	rows, err := DB.Query("SELECT id FROM applications WHERE userId = ? AND id = ?", GetSessionUserID(&ec), ec.FormValue("id"))
	if err != nil {
		log.Println("Error while updating application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	// update the application
	stmt, err := DB.Prepare("UPDATE applications SET name = ?, callbackUrl = ? WHERE id = ?")
	if err != nil {
		log.Println("Error while updating application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	_, err = stmt.Exec(ec.FormValue("name"), ec.FormValue("callbackUrl"), ec.FormValue("id"))
	if err != nil {
		log.Println("Error while updating application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func Route_Application_Manage_Delete(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	if ec.FormValue("id") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	// check that you can actually edit the application
	rows, err := DB.Query("SELECT id FROM applications WHERE userId = ? AND id = ?", GetSessionUserID(&ec), ec.FormValue("id"))
	if err != nil {
		log.Println("Error while deleting application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		ec.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		return
	}

	tx, err := DB.Begin()

	// delete authorizations
	authStmt, err := tx.Prepare("DELETE FROM application_authorizations WHERE applicationId = ?")
	if err != nil {
		log.Println("Error while deleting application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer authStmt.Close()
	_, err = authStmt.Exec(ec.FormValue("id"))
	if err != nil {
		log.Println("Error while deleting application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// delete applications
	appStmt, err := tx.Prepare("DELETE FROM applications WHERE id = ?")
	if err != nil {
		log.Println("Error while deleting application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer appStmt.Close()
	_, err = appStmt.Exec(ec.FormValue("id"))
	if err != nil {
		log.Println("Error while deleting application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	// go!
	err = tx.Commit()
	if err != nil {
		log.Println("Error while deleting application: ", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func InitApplicationAPI(e *echo.Echo) {
	e.POST("/application/completeAuth", Route(Route_Application_CompleteAuth))
	e.GET("/application/get/:id", Route(Route_Application_Get))
	e.GET("/application/getAuthorizations", Route(Route_Application_GetAuthorizations))
	e.GET("/application/requestAuth/:id", Route(Route_Application_RequestAuth))
	e.POST("/application/revokeAuth", Route(Route_Application_RevokeAuth))
	e.POST("/application/revokeSelf", Route(Route_Application_RevokeSelf))

	e.POST("/application/manage/create", Route(Route_Application_Manage_Create))
	e.GET("/application/manage/getAll", Route(Route_Application_Manage_GetAll))
	e.POST("/application/manage/update", Route(Route_Application_Manage_Update))
	e.POST("/application/manage/delete", Route(Route_Application_Manage_Delete))
}
