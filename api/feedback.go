package api

import (
	"net/http"

	"github.com/MyHomeworkSpace/api-server/config"
	"github.com/MyHomeworkSpace/api-server/slack"

	"github.com/labstack/echo"
)

func routeFeedbackAdd(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("type") == "" || ec.FormValue("text") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	_, err := DB.Exec(
		"INSERT INTO feedback(userId, type, text, screenshot) VALUES(?, ?, ?, ?)",
		c.User.ID, ec.FormValue("type"), ec.FormValue("text"), ec.FormValue("screenshot"),
	)
	if err != nil {
		ErrorLog_LogError("adding feedback", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	if config.GetCurrent().Feedback.SlackEnabled {
		screenshotStatement := "No screenshot included."
		if ec.FormValue("screenshot") != "" {
			screenshotStatement = "View screenshot on admin console."
		}

		err = slack.Post(config.GetCurrent().Feedback.SlackURL, slack.WebhookMessage{
			Attachments: []slack.WebhookAttachment{
				slack.WebhookAttachment{
					Fallback: "New feedback submission",
					Color:    "good",
					Title:    "New feedback submission",
					Text:     ec.FormValue("text"),
					Fields: []slack.WebhookField{
						slack.WebhookField{
							Title: "Feedback type",
							Value: ec.FormValue("type"),
							Short: true,
						},
						slack.WebhookField{
							Title: "Host",
							Value: config.GetCurrent().Feedback.SlackHostName,
							Short: true,
						},
						slack.WebhookField{
							Title: "User (name)",
							Value: c.User.Name,
							Short: true,
						},
						slack.WebhookField{
							Title: "User (email)",
							Value: c.User.Email,
							Short: true,
						},
						slack.WebhookField{
							Title: "User (username)",
							Value: c.User.Username,
							Short: true,
						},
						slack.WebhookField{
							Title: "User (type)",
							Value: c.User.Type,
							Short: true,
						},
						slack.WebhookField{
							Title: "Screenshot",
							Value: screenshotStatement,
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

	ec.JSON(http.StatusOK, StatusResponse{"ok"})
}
