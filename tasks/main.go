package tasks

import (
	"database/sql"
	"log"
	"time"
)

type taskFunc func(lastCompletion *time.Time, param string, db *sql.DB) error

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

func taskWatcher(taskID string, taskName string, task taskFunc, source string, db *sql.DB) {
	log.Printf("Starting task '%s' (%s)...", taskName, taskID)

	// TODO: improve error handling

	lastCompletion, err := getLastCompletion(taskID, db)
	if err != nil {
		log.Println(err)
	}

	err = task(lastCompletion, source, db)
	if err != nil {
		log.Println(err)
	}

	if err == nil {
		err = updateLastCompletion(taskID, db)
		if err != nil {
			log.Println(err)
		}
	}

	log.Printf("Task completed.")
}
