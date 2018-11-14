package api

import (
	"net/http"

	"github.com/MyHomeworkSpace/api-server/slack"

	"github.com/labstack/echo"
)

// structs for data
type Feedback struct {
	ID        int    `json:"id"`
	UserID    int    `json:"userid"`
	Type      string `json:"type"`
	Text      string `json:"text"`
	Timestamp string `json:"timestamp"`
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
			ErrorLog_LogError("adding feedback", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}
		_, err = stmt.Exec(GetSessionUserID(&c), c.FormValue("type"), c.FormValue("text"))
		if err != nil {
			ErrorLog_LogError("adding feedback", err)
			return c.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		}

		if FeedbackSlackEnabled {
			user, err := Data_GetUserByID(GetSessionUserID(&c))
			if err != nil {
				ErrorLog_LogError("posting feedback to Slack", err)
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
				ErrorLog_LogError("posting feedback to Slack", err)
			}
		}

		return c.JSON(http.StatusOK, StatusResponse{"ok"})
	})
}
