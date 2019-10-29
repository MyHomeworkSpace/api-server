package errorlog

import (
	"fmt"
	"log"
	"runtime"

	"github.com/MyHomeworkSpace/api-server/config"
	"github.com/MyHomeworkSpace/api-server/slack"
)

// LogError logs the given error to the default logger and any configured services.
func LogError(desc string, err error) {
	buf := make([]byte, 1<<16)
	stackSize := runtime.Stack(buf, false)
	stackTrace := string(buf[0:stackSize])

	log.Println("======================================")

	log.Printf("Error occurred while '%s'!", desc)
	errDesc := ""
	if err != nil {
		errDesc = err.Error()
	} else {
		errDesc = "(err == nil)"
	}
	log.Println(errDesc)
	log.Println(stackTrace)

	log.Println("======================================")

	if config.GetCurrent().ErrorLog.SlackEnabled {
		title := fmt.Sprintf("An error occurred - %s", desc)
		err = slack.Post(config.GetCurrent().ErrorLog.SlackURL, slack.WebhookMessage{
			Attachments: []slack.WebhookAttachment{
				slack.WebhookAttachment{
					Fallback: title,
					Color:    "danger",
					Title:    title,
					Text:     "```" + stackTrace + "```",
					Fields: []slack.WebhookField{
						slack.WebhookField{
							Title: "Host",
							Value: config.GetCurrent().Server.HostName,
							Short: true,
						},
						slack.WebhookField{
							Title: "Message",
							Value: err.Error(),
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
			log.Println("An error occurred while reporting the previous error to Slack")
			log.Println(err)
		}
	}
}
