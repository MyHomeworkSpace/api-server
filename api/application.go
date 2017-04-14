package api

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
)

type Application struct {
	ID int `json:"id"`
	Name string `json:"name"`
	AuthorName string `json:"authorName"`
	CallbackURL string `json:"callbackUrl"`
}

type ApplicationTokenResponse struct {
	Status string `json:"status"`
	Token string `json:"token"`
}
type SingleApplicationResponse struct {
	Status string `json:"status"`
	Application Application `json:"application"`
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
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		defer applicationRows.Close()

		if !applicationRows.Next() {
			return c.JSON(http.StatusNotFound, ErrorResponse{"error", "not_found"})
		}

		application := Application{-1, "", "", ""}
		applicationRows.Scan(&application.ID, &application.Name, &application.AuthorName, &application.CallbackURL)

		// check if we've already authorized this application
		tokenCheckRows, err := DB.Query("SELECT token FROM application_authorizations WHERE applicationId = ? AND userId = ?", application.ID, GetSessionUserID(&c))
		if err != nil {
			log.Println("Error while completing application auth: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
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
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}

		stmt, err := DB.Prepare("INSERT INTO application_authorizations(applicationId, userId, token) VALUES(?, ?, ?)")
		if err != nil {
			log.Println("Error while authorizing application:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		_, err = stmt.Exec(application.ID, GetSessionUserID(&c), token)
		if err != nil {
			log.Println("Error while authorizing application:")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, StatusResponse{"error"})
		}
		
		return c.JSON(http.StatusOK, ApplicationTokenResponse{"ok", token})
	})
	e.GET("/application/get/:id", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}

		rows, err := DB.Query("SELECT id, name, authorName, callbackUrl FROM applications WHERE clientId = ?", c.Param("id"))
		if err != nil {
			log.Println("Error while getting application information: ")
			log.Println(err)
			jsonResp := StatusResponse{"error"}
			return c.JSON(http.StatusInternalServerError, jsonResp)
		}
		defer rows.Close()

		if !rows.Next() {
			jsonResp := ErrorResponse{"error", "not_found"}
			return c.JSON(http.StatusNotFound, jsonResp)
		}

		resp := Application{-1, "", "", ""}
		rows.Scan(&resp.ID, &resp.Name, &resp.AuthorName, &resp.CallbackURL)

		jsonResp := SingleApplicationResponse{"ok", resp}
		return c.JSON(http.StatusOK, jsonResp)
	})
	e.GET("/application/requestAuth/:id", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			jsonResp := ErrorResponse{"error", "logged_out"}
			return c.JSON(http.StatusUnauthorized, jsonResp)
		}
		return c.Redirect(http.StatusFound, AuthURLBase + "?id=" + c.Param("id"))
	})
}