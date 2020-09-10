package tasks

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/MyHomeworkSpace/api-server/config"
	"github.com/MyHomeworkSpace/api-server/errorlog"
	"github.com/MyHomeworkSpace/api-server/slack"
)

type taskResponse struct {
	RowsAffected int64
}

type taskFunc func(lastCompletion *time.Time, param string, db *sql.DB) (taskResponse, error)

func getLastCompletion(taskID string, db *sql.DB) (*time.Time, error) {
	rows, err := db.Query("SELECT lastCompletion FROM internal_tasks WHERE taskID = ?", taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		lastCompletion := ""
		err = rows.Scan(&lastCompletion)
		if err != nil {
			return nil, err
		}

		time, err := time.Parse("2006-01-02", lastCompletion)
		if err != nil {
			return nil, err
		}

		return &time, nil
	}

	return nil, nil
}

func updateLastCompletion(taskID string, db *sql.DB) error {
	existingDate, err := getLastCompletion(taskID, db)
	if err != nil {
		return err
	}

	newDate := time.Now().Format("2006-01-02")

	if existingDate == nil {
		_, err = db.Exec("INSERT INTO internal_tasks(taskID, lastCompletion) VALUES(?, ?)", taskID, newDate)
		if err != nil {
			return err
		}
	} else {
		_, err = db.Exec("UPDATE internal_tasks SET lastCompletion = ? WHERE taskID = ?", newDate, taskID)
		if err != nil {
			return err
		}
	}

	return nil
}

func taskResult(taskID string, taskName string, ok bool, err error, response taskResponse) {
	taskSlackConfig := config.GetCurrent().Tasks.Slack

	// log messages are harmless notifications that a task completed
	// non-log messages are for errors that require attention
	isLogMessage := ok

	if (!isLogMessage && taskSlackConfig.SlackEnabled) || (isLogMessage && taskSlackConfig.SlackLogEnabled) {
		message := fmt.Sprintf("'%s' (%s) failed!", taskName, taskID)
		color := "danger"
		text := "The error has been logged."

		reportURL := taskSlackConfig.SlackURL
		if isLogMessage {
			reportURL = taskSlackConfig.SlackLogURL
		}

		if ok {
			message = fmt.Sprintf("'%s' (%s) completed", taskName, taskID)
			color = "good"

			if response.RowsAffected == 0 {
				// don't bother reporting it
				return
			}

			text = fmt.Sprintf("%d row(s) affected", response.RowsAffected)
		}

		err := slack.Post(reportURL, slack.WebhookMessage{
			Attachments: []slack.WebhookAttachment{
				{
					Fallback: message,
					Color:    color,
					Title:    message,
					Text:     text,
					Fields: []slack.WebhookField{
						{
							Title: "Host",
							Value: config.GetCurrent().Server.HostName,
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
			errorlog.LogError("posting task result to Slack", err)
		}
	}

	if err != nil {
		errorlog.LogError(fmt.Sprintf("running task '%s' (%s)", taskName, taskID), err)
	}
}

func taskWatcher(taskID string, taskName string, task taskFunc, source string, db *sql.DB) {
	log.Printf("Starting task '%s' (%s)...", taskName, taskID)

	lastCompletion, err := getLastCompletion(taskID, db)
	if err != nil {
		taskResult(taskID, taskName, false, err, taskResponse{})
		return
	}

	response, err := task(lastCompletion, source, db)
	if err != nil {
		taskResult(taskID, taskName, false, err, taskResponse{})
		return
	}

	if err == nil {
		err = updateLastCompletion(taskID, db)
		if err != nil {
			taskResult(taskID, taskName, false, err, taskResponse{})
			return
		}
	}

	taskResult(taskID, taskName, true, nil, response)

	log.Printf("Task completed.")
}
