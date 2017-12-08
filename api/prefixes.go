package api

import (
	"net/http"

	"github.com/labstack/echo"
)

type Prefix struct {
	Background string   `json:"background"`
	Color      string   `json:"color"`
	Words      []string `json:"words"`
	TimedEvent bool     `json:"timedEvent"`
	Default    bool     `json:"default"`
}

var DefaultPrefixes = []Prefix{
	Prefix{
		Background: "4C6C9B",
		Color:      "FFFFFF",
		Words:      []string{"HW", "Read", "Reading"},
		Default:    true,
	},
	Prefix{
		Background: "9ACD32",
		Color:      "FFFFFF",
		Words:      []string{"Project"},
		Default:    true,
	},
	Prefix{
		Background: "C3A528",
		Color:      "FFFFFF",
		Words:      []string{"Report", "Essay", "Paper", "Write"},
		Default:    true,
	},
	Prefix{
		Background: "FFA500",
		Color:      "FFFFFF",
		Words:      []string{"Quiz", "PopQuiz"},
		Default:    true,
	},
	Prefix{
		Background: "DC143C",
		Color:      "FFFFFF",
		Words:      []string{"Test", "Final", "Exam", "Midterm"},
		Default:    true,
	},
	Prefix{
		Background: "2AC0F1",
		Color:      "FFFFFF",
		Words:      []string{"ICA"},
		Default:    true,
	},
	Prefix{
		Background: "2AF15E",
		Color:      "FFFFFF",
		Words:      []string{"Lab", "Study", "Memorize"},
		TimedEvent: true,
		Default:    true,
	},
	Prefix{
		Background: "003DAD",
		Color:      "FFFFFF",
		Words:      []string{"DocID"},
		Default:    true,
	},
	Prefix{
		Background: "000000",
		Color:      "00FF00",
		Words:      []string{"Trojun", "Hex"},
		Default:    true,
	},
	Prefix{
		Background: "5000BC",
		Color:      "FFFFFF",
		Words:      []string{"OptionalHW", "Challenge"},
		Default:    true,
	},
	Prefix{
		Background: "000099",
		Color:      "FFFFFF",
		Words:      []string{"Presentation", "Prez"},
		Default:    true,
	},
	Prefix{
		Background: "123456",
		Color:      "FFFFFF",
		Words:      []string{"BuildSession", "Build"},
		TimedEvent: true,
		Default:    true,
	},
	Prefix{
		Background: "5A1B87",
		Color:      "FFFFFF",
		Words:      []string{"Meeting", "Meet"},
		TimedEvent: true,
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
	e.GET("/prefixes/getList", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}

		return c.JSON(http.StatusOK, PrefixesResponse{"ok", DefaultPrefixes, "FFD3BD", "000000"})
	})
}
