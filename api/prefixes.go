package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"

	"github.com/julienschmidt/httprouter"
)

type prefixesResponse struct {
	Status             string        `json:"status"`
	Prefixes           []data.Prefix `json:"prefixes"`
	FallbackBackground string        `json:"fallbackBackground"`
	FallbackColor      string        `json:"fallbackColor"`
}

type schoolPrefixInfo struct {
	School   data.SchoolResult `json:"school"`
	Prefixes []data.Prefix     `json:"prefixes"`
}

type defaultPrefixesResponse struct {
	Status             string             `json:"status"`
	Prefixes           []data.Prefix      `json:"prefixes"`
	SchoolPrefixes     []schoolPrefixInfo `json:"schoolPrefixes"`
	FallbackBackground string             `json:"fallbackBackground"`
	FallbackColor      string             `json:"fallbackColor"`
}

func routePrefixesGetDefaultList(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	info := []schoolPrefixInfo{}
	schools := MainRegistry.GetAllSchools()
	for _, school := range schools {
		info = append(info, schoolPrefixInfo{
			School: data.SchoolResult{
				SchoolID:    school.ID(),
				DisplayName: school.Name(),
				ShortName:   school.ShortName(),
			},
			Prefixes: school.Prefixes(),
		})
	}
	writeJSON(w, http.StatusOK, defaultPrefixesResponse{"ok", data.DefaultPrefixes, info, "FFD3BD", "000000"})
}

func routePrefixesGetList(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	prefixes, err := data.GetPrefixesForUser(c.User)
	if err != nil {
		errorlog.LogError("getting prefixes for user", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, prefixesResponse{"ok", prefixes, data.FallbackBackground, data.FallbackColor})
}

func routePrefixesDelete(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("id") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	rows, err := DB.Query("SELECT id FROM prefixes WHERE userId = ? AND id = ?", c.User.ID, r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting prefixes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()
	if !rows.Next() {
		writeJSON(w, http.StatusForbidden, errorResponse{"error", "forbidden"})
		return
	}

	_, err = DB.Exec("DELETE FROM prefixes WHERE id = ?", r.FormValue("id"))
	if err != nil {
		errorlog.LogError("deleting prefixes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}

func routePrefixesAdd(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("color") == "" || r.FormValue("background") == "" || r.FormValue("words") == "" || r.FormValue("timedEvent") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	timedEvent, err := strconv.ParseBool(r.FormValue("timedEvent"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	timedEventInt := 0
	if timedEvent {
		timedEventInt = 1
	}

	wordsInputString := r.FormValue("words")
	wordsList := []string{}
	cleanedWordsList := []string{}

	err = json.Unmarshal([]byte(wordsInputString), &wordsList)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	for _, word := range wordsList {
		if strings.TrimSpace(word) != "" {
			cleanedWordsList = append(cleanedWordsList, strings.TrimSpace(word))
		}
	}

	if len(cleanedWordsList) == 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	wordsFormatted, err := json.Marshal(cleanedWordsList)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "invalid_params"})
		return
	}

	_, err = DB.Exec("INSERT INTO prefixes(words, color, background, isTimedEvent, userId) VALUES (?, ?, ?, ?, ?)", string(wordsFormatted), r.FormValue("color"), r.FormValue("background"), timedEventInt, c.User.ID)
	if err != nil {
		errorlog.LogError("adding prefix", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}
