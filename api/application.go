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

func InitApplicationAPI(e *echo.Echo) {
	e.POST("/application/completeAuth", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		// get the application
		applicationRows, err := DB.Query("SELECT id, name, authorName, callbackUrl FROM applications WHERE clientId = ?", c.FormValue("clientId"))
		if err != nil {
			log.Println("Error while completing application auth: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer applicationRows.Close()

		if !applicationRows.Next() {
			return c.JSON(http.StatusNotFound, ErrorResponse{"error", "not_found"})
		}

		application := Application{-1, "", "", "", ""}
		applicationRows.Scan(&application.ID, &application.Name, &application.AuthorName, &application.CallbackURL)

		// check if we've already authorized this application
		tokenCheckRows, err := DB.Query("SELECT token FROM application_authorizations WHERE applicationId = ? AND userId = ?", application.ID, GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while completing application auth: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer tokenCheckRows.Close()

		if tokenCheckRows.Next() {
			// if we have, just return that token
			token := ""
			tokenCheckRows.Scan(&token)
			return c.JSON(http.StatusOK, ApplicationTokenResponse{"ok", token})
		}

		// add the new authorization
		token, err := Util_GenerateRandomString(56)
		if err != nil {
			log.Println("Error while generating application token:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		stmt, err := DB.Prepare("INSERT INTO application_authorizations(applicationId, userId, token) VALUES(?, ?, ?)")
		if err != nil {
			log.Println("Error while authorizing application:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(application.ID, GetSessionUserID(&c), token)
		if err != nil {
			log.Println("Error while authorizing application:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, ApplicationTokenResponse{"ok", token})
	})
	e.GET("/application/get/:id", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		rows, err := DB.Query("SELECT id, name, authorName, callbackUrl FROM applications WHERE clientId = ?", c.Param("id"))
		if err != nil {
			log.Println("Error while getting application information: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		if !rows.Next() {
			return c.JSON(http.StatusNotFound, ErrorResponse{"error", "not_found"})
		}

		resp := Application{-1, "", "", "", ""}
		rows.Scan(&resp.ID, &resp.Name, &resp.AuthorName, &resp.CallbackURL)

		return c.JSON(http.StatusOK, SingleApplicationResponse{"ok", resp})
	})
	e.GET("/application/getAuthorizations", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		rows, err := DB.Query("SELECT application_authorizations.id, applications.id, applications.name, applications.authorName FROM application_authorizations INNER JOIN applications ON application_authorizations.applicationId = applications.id WHERE application_authorizations.userId = ?", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting authorizations: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		authorizations := []ApplicationAuthorization{}
		for rows.Next() {
			resp := ApplicationAuthorization{-1, -1, "", ""}
			rows.Scan(&resp.ID, &resp.ApplicationID, &resp.Name, &resp.AuthorName)
			authorizations = append(authorizations, resp)
		}
		return c.JSON(http.StatusOK, ApplicationAuthorizationsResponse{"ok", authorizations})
	})
	e.GET("/application/requestAuth/:id", func(c echo.Context) error {
		state := c.FormValue("state")
		if state == "" {
			return c.Redirect(http.StatusFound, AuthURLBase+"?id="+c.Param("id"))
		} else {
			return c.Redirect(http.StatusFound, AuthURLBase+"?id="+c.Param("id")+"&state="+c.FormValue("state"))
		}
	})
	e.POST("/application/revokeAuth", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		// find the authorization
		rows, err := DB.Query("SELECT userId FROM application_authorizations WHERE id = ?", c.FormValue("id"))
		if err != nil {
			log.Println("Error while revoking authorization:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		if !rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		userId := -1
		rows.Scan(&userId)

		if GetSessionUserID(&c) != userId {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		// delete the authorization
		deleteStmt, err := DB.Prepare("DELETE FROM application_authorizations WHERE id = ?")
		if err != nil {
			log.Println("Error while revoking authorization:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer deleteStmt.Close()
		_, err = deleteStmt.Exec(c.FormValue("id"))
		if err != nil {
			log.Println("Error while revoking authorization:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
	e.POST("/application/revokeSelf", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if !HasAuthToken(&c) {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "bad_request"})
		}

		// delete the authorization
		deleteStmt, err := DB.Prepare("DELETE FROM application_authorizations WHERE token = ?")
		if err != nil {
			log.Println("Error while revoking authorization:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer deleteStmt.Close()
		_, err = deleteStmt.Exec(GetAuthToken(&c))
		if err != nil {
			log.Println("Error while revoking authorization:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/application/manage/create", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		// generate client id
		clientId, err := Util_GenerateRandomString(42)
		if err != nil {
			log.Println("Error while creating application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		// get author name
		rows, err := DB.Query("SELECT name FROM users WHERE id = ?", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while creating application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		if !rows.Next() {
			log.Println("Error while creating application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		authorName := ""
		rows.Scan(&authorName)

		// actually create the application
		stmt, err := DB.Prepare("INSERT INTO applications(name, userId, authorName, clientId, callbackUrl) VALUES('New application', ?, ?, ?, '')")
		if err != nil {
			log.Println("Error while creating application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(GetSessionUserID(&c), authorName, clientId)
		if err != nil {
			log.Println("Error while creating application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
	e.GET("/application/manage/getAll", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		rows, err := DB.Query("SELECT id, name, authorName, clientId, callbackUrl FROM applications WHERE userId = ?", GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while getting user applications: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()

		apps := []Application{}
		for rows.Next() {
			resp := Application{-1, "", "", "", ""}
			rows.Scan(&resp.ID, &resp.Name, &resp.AuthorName, &resp.ClientID, &resp.CallbackURL)
			apps = append(apps, resp)
		}

		return c.JSON(http.StatusOK, MultipleApplicationsResponse{"ok", apps})
	})
	e.POST("/application/manage/update", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if c.FormValue("id") == "" || c.FormValue("name") == "" || c.FormValue("callbackUrl") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		// check that you can actually edit the application
		rows, err := DB.Query("SELECT id FROM applications WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			log.Println("Error while updating application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		if !rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		// update the application
		stmt, err := DB.Prepare("UPDATE applications SET name = ?, callbackUrl = ? WHERE id = ?")
		if err != nil {
			log.Println("Error while updating application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(c.FormValue("name"), c.FormValue("callbackUrl"), c.FormValue("id"))
		if err != nil {
			log.Println("Error while updating application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
	e.POST("/application/manage/delete", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if c.FormValue("id") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		// check that you can actually edit the application
		rows, err := DB.Query("SELECT id FROM applications WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		if !rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		tx, err := DB.Begin()

		// delete authorizations
		authStmt, err := tx.Prepare("DELETE FROM application_authorizations WHERE applicationId = ?")
		if err != nil {
			log.Println("Error while deleting application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer authStmt.Close()
		_, err = authStmt.Exec(c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		// delete applications
		appStmt, err := tx.Prepare("DELETE FROM applications WHERE id = ?")
		if err != nil {
			log.Println("Error while deleting application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer appStmt.Close()
		_, err = appStmt.Exec(c.FormValue("id"))
		if err != nil {
			log.Println("Error while deleting application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		// go!
		err = tx.Commit()
		if err != nil {
			log.Println("Error while deleting application: ", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
