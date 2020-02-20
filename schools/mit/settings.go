package mit

import (
	"database/sql"

	"github.com/MyHomeworkSpace/api-server/data"
)

func (s *school) GetSettings(db *sql.DB, user *data.User) (map[string]interface{}, error) {
	// get user's registration
	classRows, err := db.Query("SELECT subjectID, sectionID, title, units, sections FROM mit_classes WHERE userID = ?", user.ID)
	if err != nil {
		return nil, err
	}
	defer classRows.Close()

	registeredClasses := []map[string]interface{}{}
	for classRows.Next() {
		subjectID, sectionID, title, units, selectedSections := "", "", "", -1, ""

		err = classRows.Scan(&subjectID, &sectionID, &title, &units, &selectedSections)
		if err != nil {
			return nil, err
		}

		sectionRows, err := db.Query("SELECT section, time, place, facultyID, facultyName FROM mit_offerings WHERE id = ?", subjectID)
		if err != nil {
			return nil, err
		}

		sections := []map[string]interface{}{}

		for sectionRows.Next() {
			sectionCode, time, place, facultyID, facultyName := "", "", "", "", ""
			err = sectionRows.Scan(&sectionCode, &time, &place, &facultyID, &facultyName)
			if err != nil {
				sectionRows.Close()

				return nil, err
			}

			section := map[string]interface{}{
				"sectionCode": sectionCode,
				"time":        time,
				"place":       place,
				"facultyID":   facultyID,
				"facultyName": facultyName,
			}

			sections = append(sections, section)
		}

		sectionRows.Close()

		registeredClass := map[string]interface{}{
			"subjectID":        subjectID,
			"sectionID":        sectionID,
			"title":            title,
			"units":            units,
			"selectedSections": selectedSections,
			"sections":         sections,
		}
		registeredClasses = append(registeredClasses, registeredClass)
	}

	return map[string]interface{}{
		"registration": registeredClasses,
		"peInfo":       s.peInfo,
		"showPE":       s.showPE,
	}, nil
}

func (s *school) SetSettings(db *sql.DB, user *data.User, settings map[string]interface{}) (*sql.Tx, map[string]interface{}, error) {
	var err error

	showPE := s.showPE
	showPEInterface, ok := settings["showPE"]

	if ok {
		showPE, ok = showPEInterface.(bool)
		if !ok {
			return nil, nil, data.SchoolError{Code: "invalid_params"}
		}
	}

	sections, ok := settings["sections"]

	if !ok {
		return nil, nil, data.SchoolError{Code: "missing_params"}
	}

	sectionMap, ok := sections.(map[string]interface{})

	if !ok {
		return nil, nil, data.SchoolError{Code: "invalid_params"}
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, nil, err
	}

	for subjectID, subjectSections := range sectionMap {
		_, err = tx.Exec(
			"UPDATE mit_classes SET sections = ? WHERE subjectID = ? AND userID = ?",
			subjectSections,
			subjectID,
			user.ID,
		)
		if err != nil {
			tx.Rollback()
			return nil, nil, err
		}
	}

	return tx, map[string]interface{}{
		"showPE": showPE,
	}, nil
}
