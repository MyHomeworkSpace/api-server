package api

import (
	"net/http"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/labstack/echo"
)

// responses
type PrefResponse struct {
	Status string    `json:"status"`
	Pref   data.Pref `json:"pref"`
}

type PrefsResponse struct {
	Status string      `json:"status"`
	Prefs  []data.Pref `json:"prefs"`
}

func routePrefsGet(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", c.User.ID, ec.Param("key"))
	if err != nil {
		ErrorLog_LogError("getting pref", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		ec.JSON(http.StatusNotFound, ErrorResponse{"error", "not_found"})
		return
	}

	resp := data.Pref{-1, "", ""}
	rows.Scan(&resp.ID, &resp.Key, &resp.Value)

	ec.JSON(http.StatusOK, PrefResponse{"ok", resp})
}

func routePrefsGetAll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ?", c.User.ID)
	if err != nil {
		ErrorLog_LogError("getting prefs", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	prefs := []data.Pref{}

	for rows.Next() {
		pref := data.Pref{}
		rows.Scan(&pref.ID, &pref.Key, &pref.Value)
		prefs = append(prefs, pref)
	}

	ec.JSON(http.StatusOK, PrefsResponse{"ok", prefs})
}

func routePrefsSet(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("key") == "" || ec.FormValue("value") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", c.User.ID, ec.FormValue("key"))
	if err != nil {
		ErrorLog_LogError("setting pref", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		// doesn't exist, add it
		_, err = DB.Exec(
			"INSERT INTO prefs(userId, `key`, `value`) VALUES(?, ?, ?)",
			c.User.ID, ec.FormValue("key"), ec.FormValue("value"),
		)
		if err != nil {
			ErrorLog_LogError("inserting pref", err)
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}
	} else {
		// exists already, update it
		_, err = DB.Exec(
			"UPDATE prefs SET `key` = ?, `value` = ? WHERE userId = ? AND `key` = ?",
			ec.FormValue("key"), ec.FormValue("value"), c.User.ID, ec.FormValue("key"),
		)
		if err != nil {
			ErrorLog_LogError("updating pref", err)
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}
