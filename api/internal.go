package api

import (
	"net/http"
	"strings"

	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/MyHomeworkSpace/api-server/tasks"

	"github.com/julienschmidt/httprouter"
)

func routeInternalStartTask(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	task := r.FormValue("task")

	if task == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	if task != "calendar:sync" && task != "mit:fetch:catalog" && task != "mit:fetch:coursews" && task != "mit:fetch:offerings" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	if task == "calendar:sync" {
		err := tasks.StartCalendarSync(DB)
		if err != nil {
			errorlog.LogError("starting task", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}
	} else {
		source := strings.Replace(task, "mit:fetch:", "", -1)

		err := tasks.StartImportFromMIT(source, DB)
		if err != nil {
			errorlog.LogError("starting task", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}
