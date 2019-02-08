package slack

import (
	"encoding/json"
	"net/http"
	"strings"
)

// A WebhookField is a key-value field for a Slack webhook attachment.
type WebhookField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// A WebhookAttachment is an attachment for a Slack webhook message.
type WebhookAttachment struct {
	Fallback   string         `json:"fallback"`
	Color      string         `json:"color"`
	Title      string         `json:"title"`
	Text       string         `json:"text"`
	Fields     []WebhookField `json:"fields"`
	MarkdownIn []string       `json:"mrkdwn_in"`
	ImageURL   string         `json:"image_url"`
}

// A WebhookMessage is a message sent to a Slack webhook URL.
type WebhookMessage struct {
	Attachments []WebhookAttachment `json:"attachments"`
}

// Post posts the given WebhookMessage to the given webhook URL.
func Post(url string, msg WebhookMessage) error {
	slackMessageJSON, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "text/json", strings.NewReader(string(slackMessageJSON)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
