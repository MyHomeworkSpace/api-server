package api

import (
	"net/http"

	"github.com/MyHomeworkSpace/api-server/config"
	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/MyHomeworkSpace/api-server/slack"

	"github.com/julienschmidt/httprouter"
)

func routeFeedbackAdd(w http.ResponseWriter, r *http.Request, p httprouter.Params, c RouteContext) {
	if r.FormValue("type") == "" || r.FormValue("text") == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse{"error", "missing_params"})
		return
	}

	_, err := DB.Exec(
		"INSERT INTO feedback(userId, type, text, screenshot) VALUES(?, ?, ?, ?)",
		c.User.ID, r.FormValue("type"), r.FormValue("text"), r.FormValue("screenshot"),
	)
	if err != nil {
		errorlog.LogError("adding feedback", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"error", "internal_server_error"})
		return
	}

	if config.GetCurrent().Feedback.SlackEnabled {
		screenshotStatement := "No screenshot included."
		if r.FormValue("screenshot") != "" {
			screenshotStatement = "View screenshot on admin console."
		}

		err = slack.Post(config.GetCurrent().Feedback.SlackURL, slack.WebhookMessage{
			Attachments: []slack.WebhookAttachment{
				slack.WebhookAttachment{
					Fallback: "New feedback submission",
					Color:    "good",
					Title:    "New feedback submission",
					Text:     r.FormValue("text"),
					Fields: []slack.WebhookField{
						slack.WebhookField{
							Title: "Feedback type",
							Value: r.FormValue("type"),
							Short: true,
						},
						slack.WebhookField{
							Title: "Host",
							Value: config.GetCurrent().Server.HostName,
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
			errorlog.LogError("posting feedback to Slack", err)
		}
	}

	writeJSON(w, http.StatusOK, statusResponse{"ok"})
}
