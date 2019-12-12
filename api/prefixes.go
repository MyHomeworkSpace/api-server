package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/labstack/echo"
)

var DefaultPrefixes = []data.Prefix{
	data.Prefix{
		ID:         -1,
		Background: "4C6C9B",
		Color:      "FFFFFF",
		Words:      []string{"HW", "Read", "Reading"},
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "9ACD32",
		Color:      "FFFFFF",
		Words:      []string{"Project"},
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "C3A528",
		Color:      "FFFFFF",
		Words:      []string{"Report", "Essay", "Paper", "Write"},
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "FFA500",
		Color:      "FFFFFF",
		Words:      []string{"Quiz", "PopQuiz", "GradedHW", "GradedHomework"},
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "EE5D1E",
		Color:      "FFFFFF",
		Words:      []string{"Quest", "HalfTest"},
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "DC143C",
		Color:      "FFFFFF",
		Words:      []string{"Test", "Final", "Exam", "Midterm", "Ahh"},
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "2AC0F1",
		Color:      "FFFFFF",
		Words:      []string{"ICA", "FieldTrip", "Thingy"},
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "2AF15E",
		Color:      "FFFFFF",
		Words:      []string{"Study", "Memorize"},
		TimedEvent: true,
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "003DAD",
		Color:      "FFFFFF",
		Words:      []string{"DocID"},
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "000000",
		Color:      "00FF00",
		Words:      []string{"Trojun", "Hex"},
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "5000BC",
		Color:      "FFFFFF",
		Words:      []string{"OptionalHW", "Challenge"},
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "000099",
		Color:      "FFFFFF",
		Words:      []string{"Presentation", "Prez"},
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "123456",
		Color:      "FFFFFF",
		Words:      []string{"BuildSession", "Build"},
		TimedEvent: true,
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "5A1B87",
		Color:      "FFFFFF",
		Words:      []string{"Meeting", "Meet"},
		TimedEvent: true,
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "01B501",
		Color:      "FFFFFF",
		Words:      []string{"Begin", "Start", "Do"},
		TimedEvent: true,
		Default:    true,
	},
	data.Prefix{
		ID:         -1,
		Background: "E34000",
		Color:      "FFFFFF",
		Words:      []string{"Apply", "Application", "Deadline"},
		Default:    true,
	},
}

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

func routePrefixesGetDefaultList(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
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
	writeJSON(w, http.StatusOK, defaultPrefixesResponse{"ok", DefaultPrefixes, info, "FFD3BD", "000000"})
}

func routePrefixesGetList(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	prefixes := DefaultPrefixes

	// check for school prefixes we want to add
	for _, school := range c.User.Schools {
		if school.Enabled {
			prefixes = append(prefixes, school.School.Prefixes()...)
		}
	}

	// load user settings
	rows, err := DB.Query("SELECT id, background, color, words, isTimedEvent FROM prefixes WHERE userId = ?", c.User.ID)
	if err != nil {
		errorlog.LogError("getting custom prefixes", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		resp := data.Prefix{}

		timedEventInt := -1
		wordsListString := ""

		rows.Scan(&resp.ID, &resp.Background, &resp.Color, &wordsListString, &timedEventInt)

		err := json.Unmarshal([]byte(wordsListString), &resp.Words)
		if err != nil {
			errorlog.LogError("parsing custom prefix words", err)
		}

		resp.TimedEvent = (timedEventInt == 1)

		prefixes = append(prefixes, resp)
	}

	writeJSON(w, http.StatusOK, prefixesResponse{"ok", prefixes, "FFD3BD", "000000"})
}

func routePrefixesDelete(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
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

func routePrefixesAdd(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
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
