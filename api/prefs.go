package api

import (
	"net/http"

	"github.com/labstack/echo"
)

// structs for data
type Pref struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// responses
type PrefResponse struct {
	Status string `json:"status"`
	Pref   Pref   `json:"pref"`
}

type PrefsResponse struct {
	Status string `json:"status"`
	Prefs  []Pref `json:"prefs"`
}

func routePrefsGet(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", GetSessionUserID(&ec), ec.Param("key"))
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

	resp := Pref{-1, "", ""}
	rows.Scan(&resp.ID, &resp.Key, &resp.Value)

	ec.JSON(http.StatusOK, PrefResponse{"ok", resp})
}

func routePrefsGetAll(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ?", GetSessionUserID(&ec))
	if err != nil {
		ErrorLog_LogError("getting prefs", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	prefs := []Pref{}

	for rows.Next() {
		pref := Pref{}
		rows.Scan(&pref.ID, &pref.Key, &pref.Value)
		prefs = append(prefs, pref)
	}

	ec.JSON(http.StatusOK, PrefsResponse{"ok", prefs})
}

func routePrefsSet(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}

	if ec.FormValue("key") == "" || ec.FormValue("value") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", GetSessionUserID(&ec), ec.FormValue("key"))
	if err != nil {
		ErrorLog_LogError("setting pref", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		// doesn't exist, add it
		stmt, err := DB.Prepare("INSERT INTO prefs(userId, `key`, `value`) VALUES(?, ?, ?)")
		if err != nil {
			ErrorLog_LogError("inserting pref", err)
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}
		_, err = stmt.Exec(GetSessionUserID(&ec), ec.FormValue("key"), ec.FormValue("value"))
		if err != nil {
			ErrorLog_LogError("inserting pref", err)
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}
	} else {
		// exists already, update it
		stmt, err := DB.Prepare("UPDATE prefs SET `key` = ?, `value` = ? WHERE userId = ? AND `key` = ?")
		if err != nil {
			ErrorLog_LogError("updating pref", err)
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}
		_, err = stmt.Exec(ec.FormValue("key"), ec.FormValue("value"), GetSessionUserID(&ec), ec.FormValue("key"))
		if err != nil {
			ErrorLog_LogError("updating pref", err)
			ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
			return
		}
	}

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}

func InitPrefsAPI(e *echo.Echo) {
	e.GET("/prefs/get/:key", Route(routePrefsGet))
	e.GET("/prefs/getAll", Route(routePrefsGetAll))
	e.POST("/prefs/set", Route(routePrefsSet))
}
