package api

import (
	"net/http"
	"strings"

	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/MyHomeworkSpace/api-server/tasks"
	"github.com/labstack/echo"
)

func routeInternalStartTask(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	task := r.FormValue("task")

	if task == "" {
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "missing_params"})
		return
	}

	if task != "mit:fetch:catalog" && task != "mit:fetch:offerings" {
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "invalid_params"})
		return
	}

	source := strings.Replace(task, "mit:fetch:", "", -1)

	err := tasks.StartImportFromMIT(source, DB)
	if err != nil {
		errorlog.LogError("starting task", err)
		ec.JSON(http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	ec.JSON(http.StatusOK, statusResponse{"ok"})
}
