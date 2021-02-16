package api

import (
	"net/http"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/julienschmidt/httprouter"
)

// responses
type prefResponse struct {
	Status string    `json:"status"`
	Pref   data.Pref `json:"pref"`
}

type prefsResponse struct {
	Status string      `json:"status"`
	Prefs  []data.Pref `json:"prefs"`
}

func routePrefsGet(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", c.User.ID, p.ByName("key"))
	if err != nil {
		errorlog.LogError("getting pref", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		writeJSON(w, http.StatusNotFound, errorResponse{"error", "not_found"})
		return
	}

	resp := data.Pref{}
	rows.Scan(&resp.ID, &resp.Key, &resp.Value)

	writeJSON(w, http.StatusOK, prefResponse{"ok", resp})
}

func routePrefsGetAll(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ?", c.User.ID)
	if err != nil {
		errorlog.LogError("getting prefs", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	prefs := []data.Pref{}

	for rows.Next() {
		pref := data.Pref{}
		rows.Scan(&pref.ID, &pref.Key, &pref.Value)
		prefs = append(prefs, pref)
	}

	writeJSON(w, http.StatusOK, prefsResponse{"ok", prefs})
}

func routePrefsSet(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("key") == "" || r.FormValue("value") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", c.User.ID, r.FormValue("key"))
	if err != nil {
		errorlog.LogError("setting pref", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		// doesn't exist, add it
		_, err = DB.Exec(
			"INSERT INTO prefs(userId, `key`, `value`) VALUES(?, ?, ?)",
			c.User.ID, r.FormValue("key"), r.FormValue("value"),
		)
		if err != nil {
			errorlog.LogError("inserting pref", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}
	} else {
		// exists already, update it
		_, err = DB.Exec(
			"UPDATE prefs SET `key` = ?, `value` = ? WHERE userId = ? AND `key` = ?",
			r.FormValue("key"), r.FormValue("value"), c.User.ID, r.FormValue("key"),
		)
		if err != nil {
			errorlog.LogError("updating pref", err)
			writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
			return
		}
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}
