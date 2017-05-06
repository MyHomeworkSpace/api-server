package api

import (
	//"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/MyHomeworkSpace/api-server/auth"

	"github.com/labstack/echo"
)

// structs for data

// responses

func InitCalendarAPI(e *echo.Echo) {
	e.POST("/calendar/import", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("password") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		// test the credentials first so we don't run into blackbaud's rate limiting
		_, resp, err := auth.DaltonLogin(GetSessionInfo(&c).Username, c.FormValue("password"))
		if resp != "" || err != nil {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", resp})
		}

		// set up ajax token and stuff
		ajaxToken, err := Blackbaud_GetAjaxToken()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "ajaxtoken_error"})
		}

		jar, err := cookiejar.New(nil)
	    if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		// sign in to blackbaud
		response, err := Blackbaud_Request("POST", "SignIn", url.Values{}, map[string]interface{}{
			"From": "",
			"InterfaceSource": "WebApp",
			"Password": c.FormValue("password"),
			"Username": GetSessionInfo(&c).Username,
			"remember": "false",
		}, jar, ajaxToken)

		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "bb_signin_error"})
		}

		result, worked := response["AuthenticationResult"].(float64)

		if worked && result == 5 {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "bb_signin_rate_limit"})
		}

		if !worked || result == 2 {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "bb_signin_error"})
		}

		// get list of classes

		// find all periods of classes

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
