package api

import (
	"net/http"
	"strings"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/labstack/echo"
)

var MainRegistry data.SchoolRegistry

type SchoolResultResponse struct {
	Status string             `json:"status"`
	School *data.SchoolResult `json:"school"`
}

func routeSchoolsLookup(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("email") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	email := ec.FormValue("email")

	emailParts := strings.Split(email, "@")

	if len(emailParts) != 2 {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		return
	}

	domain := strings.ToLower(strings.TrimSpace(emailParts[1]))

	school, err := MainRegistry.GetSchoolByEmailDomain(domain)

	if err == data.ErrNotFound {
		ec.JSON(http.StatusOK, SchoolResultResponse{"ok", nil})
		return
	} else if err != nil {
		ErrorLog_LogError("looking up school by email domain", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	schoolResult := data.SchoolResult{
		SchoolID:    school.ID(),
		DisplayName: school.Name(),
	}

	ec.JSON(http.StatusOK, SchoolResultResponse{"ok", &schoolResult})
}
