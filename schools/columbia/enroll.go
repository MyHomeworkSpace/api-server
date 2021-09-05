package columbia

import (
	"database/sql"
	"os"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
	"github.com/PuerkitoBio/goquery"
)

func clearUserData(tx *sql.Tx, user *data.User) error {
	// clear away anything that is in the db
	_, err := tx.Exec("DELETE FROM columbia_classes WHERE userID = ?", user.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM columbia_meetings WHERE userID = ?", user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *school) Enroll(tx *sql.Tx, user *data.User, params map[string]interface{}) (map[string]interface{}, error) {
	// TODO: login, navigate to schedule page, show user info

	// TODO: right now you need to download the html file yourself
	ssolFile, err := os.Open("ssol.html")
	if err != nil {
		return nil, err
	}
	defer ssolFile.Close()

	ssolScheduleDoc, err := goquery.NewDocumentFromReader(ssolFile)
	if err != nil {
		return nil, err
	}

	studentName, err := s.parseSSOLSchedulePage(tx, user, ssolScheduleDoc)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status": schools.ImportStatusOK,

		"name":     studentName,
		"username": "tst1234", // TODO: fix
	}, nil
}

func (s *school) Unenroll(tx *sql.Tx, user *data.User) error {
	return clearUserData(tx, user)
}

func (s *school) NeedsUpdate(db *sql.DB) (bool, error) {
	return (s.importStatus != schools.ImportStatusOK), nil
}
