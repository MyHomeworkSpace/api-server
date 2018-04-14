package api

import (
	"log"
	"net/http"

	"github.com/MyHomeworkSpace/api-server/slack"

	"github.com/labstack/echo"
)

// structs for data
type Feedback struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Due      string `json:"due"`
	Desc     string `json:"desc"`
	Complete int    `json:"complete"`
	ClassID  int    `json:"classId"`
	UserID   int    `json:"userId"`
}

// responses

func InitFeedbackAPI(e *echo.Echo) {
	e.POST("/feedback/add", func(c echo.Context) error {
		if GetSessionUserID(&c) == -1 {
			return c.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		}
		if c.FormValue("type") == "" || c.FormValue("text") == "" {
			return c.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		}

		stmt, err := DB.Prepare("INSERT INTO feedback(userId, type, text) VALUES(?, ?, ?)")
		if err != nil {
			log.Println("Error while adding feedback: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(GetSessionUserID(&c), c.FormValue("type"), c.FormValue("text"))
		if err != nil {
			log.Println("Error while adding feedback: ")
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		if FeedbackSlackEnabled {
			user, err := Data_GetUserByID(GetSessionUserID(&c))
			if err != nil {
				log.Println("Error while posting feedback to Slack: ")
				log.Println(err)
				return c.JSON(http.StatusOK, StatusResponse{"ok"})
			}

			err = slack.Post(FeedbackSlackURL, slack.WebhookMessage{
				Attachments: []slack.WebhookAttachment{
					slack.WebhookAttachment{
						Fallback: "New feedback submission",
						Color:    "good",
						Title:    "New feedback submission",
						Text:     c.FormValue("text"),
						Fields: []slack.WebhookField{
							slack.WebhookField{
								Title: "Feedback type",
								Value: c.FormValue("type"),
								Short: true,
							},
							slack.WebhookField{
								Title: "Host",
								Value: FeedbackSlackHostName,
								Short: true,
							},
							slack.WebhookField{
								Title: "User (name)",
								Value: user.Name,
								Short: true,
							},
							slack.WebhookField{
								Title: "User (email)",
								Value: user.Email,
								Short: true,
							},
							slack.WebhookField{
								Title: "User (username)",
								Value: user.Username,
								Short: true,
							},
							slack.WebhookField{
								Title: "User (type)",
								Value: user.Type,
								Short: true,
							},
						},
						MarkdownIn: []string{
							"fields",
						},
					},
				},
			})
			if err != nil {
				log.Println("Error while posting feedback to Slack: ")
				log.Println(err)
			}
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
