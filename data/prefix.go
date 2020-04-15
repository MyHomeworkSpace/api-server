package data

import (
	"encoding/json"
)

// A Prefix defines a group of words that get automatically recognized (for example: HW, Test, Quiz)
type Prefix struct {
	ID         int      `json:"id"`
	Background string   `json:"background"`
	Color      string   `json:"color"`
	Words      []string `json:"words"`
	TimedEvent bool     `json:"timedEvent"`
	Default    bool     `json:"default"`
}

// DefaultPrefixes is the list of prefixes that all users start out with.
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
		Background: "EE5D1E",
		Color:      "FFFFFF",
		Words:      []string{"Quest"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "DC143C",
		Color:      "FFFFFF",
		Words:      []string{"Test", "Final", "Exam", "Midterm"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "2AC0F1",
		Color:      "FFFFFF",
		Words:      []string{"ICA", "FieldTrip"},
		Default:    true,
	},
	Prefix{
		ID:         -1,
		Background: "2AF15E",
		Color:      "FFFFFF",
		Words:      []string{"Study", "Memorize"},
		TimedEvent: true,
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
	Prefix{
		ID:         -1,
		Background: "3F4146",
		Color:      "FFFFFF",
		Words:      []string{"Form", "File", "Submit"},
		Default:    true,
	},
}

// FallbackBackground is the background color of a word that does not have an associated prefix.
const FallbackBackground = "FFD3BD"

// FallbackColor is the text color of a word that does not have an associated prefix.
const FallbackColor = "000000"

// GetPrefixesForUser returns a list of all prefixes for the given user, factoring in schools and custom settings.
func GetPrefixesForUser(user *User) ([]Prefix, error) {
	prefixes := DefaultPrefixes

	// check for school prefixes we want to add
	for _, school := range user.Schools {
		if school.Enabled {
			prefixes = append(prefixes, school.School.Prefixes()...)
		}
	}

	// load user settings
	rows, err := DB.Query("SELECT id, background, color, words, isTimedEvent FROM prefixes WHERE userId = ?", user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		resp := Prefix{}

		timedEventInt := -1
		wordsListString := ""

		rows.Scan(&resp.ID, &resp.Background, &resp.Color, &wordsListString, &timedEventInt)

		err := json.Unmarshal([]byte(wordsListString), &resp.Words)
		if err != nil {
			return nil, err
		}

		resp.TimedEvent = (timedEventInt == 1)

		prefixes = append(prefixes, resp)
	}

	return prefixes, nil
}
