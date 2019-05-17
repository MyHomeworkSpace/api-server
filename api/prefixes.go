package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo"
)

type Prefix struct {
	ID         int      `json:"id"`
	Background string   `json:"background"`
	Color      string   `json:"color"`
	Words      []string `json:"words"`
	TimedEvent bool     `json:"timedEvent"`
	Default    bool     `json:"default"`
}

var DefaultPrefixes = []Prefix{
	Prefix{
		ID:         -1,
		Background: "4C6C9B",
		Color:      "FFFFFF",
		Words:      []string{"HW", "Read", "Reading"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "9ACD32",
		Color:      "FFFFFF",
		Words:      []string{"Project"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "C3A528",
		Color:      "FFFFFF",
		Words:      []string{"Report", "Essay", "Paper", "Write"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "FFA500",
		Color:      "FFFFFF",
		Words:      []string{"Quiz", "PopQuiz", "GradedHW", "GradedHomework"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "DC143C",
		Color:      "FFFFFF",
		Words:      []string{"Test", "Final", "Exam", "Midterm", "Ahh"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "2AC0F1",
		Color:      "FFFFFF",
		Words:      []string{"ICA", "FieldTrip", "Thingy"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "2AF15E",
		Color:      "FFFFFF",
		Words:      []string{"Lab", "BookALab", "BookLab", "Study", "Memorize"},
		TimedEvent: true,
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "003DAD",
		Color:      "FFFFFF",
		Words:      []string{"DocID"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "000000",
		Color:      "00FF00",
		Words:      []string{"Trojun", "Hex"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "5000BC",
		Color:      "FFFFFF",
		Words:      []string{"OptionalHW", "Challenge"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "000099",
		Color:      "FFFFFF",
		Words:      []string{"Presentation", "Prez"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "123456",
		Color:      "FFFFFF",
		Words:      []string{"BuildSession", "Build"},
		TimedEvent: true,
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "5A1B87",
		Color:      "FFFFFF",
		Words:      []string{"Meeting", "Meet"},
		TimedEvent: true,
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "01B501",
		Color:      "FFFFFF",
		Words:      []string{"Begin", "Start", "Do"},
		TimedEvent: true,
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "E34000",
		Color:      "FFFFFF",
		Words:      []string{"Apply", "Application", "Deadline"},
		Default:    true,
	},
}

type PrefixesResponse struct {
	Status             string   `json:"status"`
	Prefixes           []Prefix `json:"prefixes"`
	FallbackBackground string   `json:"fallbackBackground"`
	FallbackColor      string   `json:"fallbackColor"`
}

func InitPrefixesAPI(e *echo.Echo) {
	e.GET("/prefixes/getDefaultList", func(c echo.Context) error {
		return c.JSON(http.StatusOK, PrefixesResponse{"ok", DefaultPrefixes, "FFD3BD", "000000"})
	})

	e.GET("/prefixes/getList", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		rows, err := DB.Query("SELECT id, background, color, words, isTimedEvent FROM prefixes WHERE userId = ?", GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("getting custom prefixes", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		prefixes := DefaultPrefixes
		for rows.Next() {
			resp := Prefix{-1, "", "", []string{}, false, false}

			timedEventInt := -1
			wordsListString := ""

			rows.Scan(&resp.ID, &resp.Background, &resp.Color, &wordsListString, &timedEventInt)

			err := json.Unmarshal([]byte(wordsListString), &resp.Words)
			if err != nil {
				ErrorLog_LogError("parsing custom prefix words", err)
			}

			resp.TimedEvent = (timedEventInt == 1)

			prefixes = append(prefixes, resp)
		}
		return c.JSON(http.StatusOK, PrefixesResponse{"ok", prefixes, "FFD3BD", "000000"})
	})

	e.POST("/prefixes/delete", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if c.FormValue("id") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		rows, err := DB.Query("SELECT id FROM prefixes WHERE userId = ? AND id = ?", GetSessionUserID(&c), c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("deleting prefixes", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		defer rows.Close()
		if !rows.Next() {
			return c.JSON(http.StatusForbidden, ErrorResponse{"error", "forbidden"})
		}

		_, err = DB.Exec("DELETE FROM prefixes WHERE id = ?", c.FormValue("id"))
		if err != nil {
			ErrorLog_LogError("deleting prefixes", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})

	e.POST("/prefixes/add", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		if c.FormValue("color") == "" || c.FormValue("background") == "" || c.FormValue("words") == "" || c.FormValue("timedEvent") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		timedEvent, err := strconv.ParseBool(c.FormValue("timedEvent"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		timedEventInt := 0
		if timedEvent {
			timedEventInt = 1
		}

		wordsInputString := c.FormValue("words")
		wordsList := []string{}
		cleanedWordsList := []string{}

		err = json.Unmarshal([]byte(wordsInputString), &wordsList)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		for _, word := range wordsList {
			if strings.TrimSpace(word) != "" {
				cleanedWordsList = append(cleanedWordsList, strings.TrimSpace(word))
			}
		}

		if len(cleanedWordsList) == 0 {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		wordsFormatted, err := json.Marshal(cleanedWordsList)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "invalid_params"})
		}

		_, err = DB.Exec("INSERT INTO prefixes(words, color, background, isTimedEvent, userId) VALUES (?, ?, ?, ?, ?)", string(wordsFormatted), c.FormValue("color"), c.FormValue("background"), timedEventInt, GetSessionUserID(&c))
		if err != nil {
			ErrorLog_LogError("adding prefix", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
