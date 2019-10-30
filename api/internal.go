package api

import (
	"net/http"

	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/MyHomeworkSpace/api-server/tasks"
	"github.com/labstack/echo"
)

func routeInternalStartTask(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	source := r.FormValue("source")

	if source == "" {
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "missing_params"})
		return
	}

	if source != "catalog" && source != "offerings" {
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "invalid_params"})
		return
	}

	err := tasks.StartImportFromMIT(source, DB)
	if err != nil {
		errorlog.LogError("starting task", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}
