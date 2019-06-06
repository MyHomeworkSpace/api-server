package api

import (
	"net/http"

	"github.com/MyHomeworkSpace/api-server/slack"

	"github.com/labstack/echo"
)

func routeFeedbackAdd(w http.ResponseWriter, r *http.Request, ec echo.Context, c RouteContext) {
	if GetSessionUserID(&ec) == -1 {
		ec.JSON(http.StatusUnauthorized, ErrorResponse{"error", "logged_out"})
		return
	}
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
	_, err = stmt.Exec(GetSessionUserID(&ec), ec.FormValue("type"), ec.FormValue("text"), ec.FormValue("screenshot"))
	if err != nil {
		ErrorLog_LogError("adding feedback", err)
		ec.JSON(http.StatusInternalServerError, ErrorResponse{"error", "internal_server_error"})
		return
	}

	if FeedbackSlackEnabled {
		user, err := Data_GetUserByID(GetSessionUserID(&ec))
		if err != nil {
			ErrorLog_LogError("posting feedback to Slack", err)
			ec.JSON(http.StatusOK, StatusResponse{"ok"})
			return
		}

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
