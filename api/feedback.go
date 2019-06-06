package api

import (
	"net/http"

	"github.com/MyHomeworkSpace/api-server/slack"

	"github.com/labstack/echo"
)

func routeFeedbackAdd(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if ec.FormValue("type") == "" || ec.FormValue("text") == "" {
		ec.JSON(http.StatusBadRequest, ErrorResponse{"error", "missing_params"})
		return
	}

	stmt, err := DB.Prepare("INSERT INTO feedback(userId, type, text, screenshot) VALUES(?, ?, ?, ?)")
	if err != nil {
		ErrorLog_LogError("adding feedback", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}
	_, err = stmt.Exec(c.User.ID, ec.FormValue("type"), ec.FormValue("text"), ec.FormValue("screenshot"))
	if err != nil {
		ErrorLog_LogError("adding feedback", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	if FeedbackSlackEnabled {
		screenshotStatement := "No screenshot included."
		if ec.FormValue("screenshot") != "" {
			screenshotStatement = "View screenshot on admin console."
		}

		err = slack.Post(FeedbackSlackURL, slack.WebhookMessage{
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
							Value: FeedbackSlackHostName,
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
