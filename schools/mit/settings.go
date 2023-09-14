package mit

import (
	"database/sql"
	"strings"

	"github.com/MyHomeworkSpace/api-server/data"
)

func (s *school) GetSettings(db *sql.DB, user *data.User) (map[string]interface{}, error) {
	// get user's registration
	classRows, err := db.Query("SELECT subjectID, sectionID, title, units, sections, custom FROM mit_classes WHERE userID = ? ORDER BY custom, subjectID ASC", user.ID)
	if err != nil {
		return nil, err
	}
	defer classRows.Close()

	registeredClasses := []map[string]interface{}{}
	for classRows.Next() {
		subjectID, sectionID, title, units, selectedSections, custom := "", "", "", -1, "", false

		err = classRows.Scan(&subjectID, &sectionID, &title, &units, &selectedSections, &custom)
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
			"custom":           custom,
		}
		registeredClasses = append(registeredClasses, registeredClass)
	}

	return map[string]interface{}{
		"registration": registeredClasses,
		"peInfo":       s.peInfo,
		"showPE":       s.showPE,
	}, nil
}

func (s *school) CallSettingsMethod(db *sql.DB, user *data.User, methodName string, methodParams map[string]interface{}) (map[string]interface{}, error) {
	if methodName == "addCustomClass" {
		subjectNumberInterface, ok := methodParams["subjectNumber"]
		if !ok {
			return map[string]interface{}{
				"status": "error",
				"error":  "missing_params",
			}, nil
		}

		subjectNumber, ok := subjectNumberInterface.(string)
		if !ok {
			return map[string]interface{}{
				"status": "error",
				"error":  "invalid_params",
			}, nil
		}

		subjectNumber = strings.TrimSpace(subjectNumber)
		subjectNumber = strings.ToUpper(subjectNumber)

		// find the class (should have a non-zero number of offerings)
		rows, err := db.Query("SELECT title FROM mit_offerings WHERE id = ? LIMIT 1", subjectNumber)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if !rows.Next() {
			return map[string]interface{}{
				"status": "error",
				"error":  "invalid_params",
				"hint":   "Subject number has no offerings this term.",
			}, nil
		}

		name := ""
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}

		// the class exists this term
		// add it to the user's registration
		_, err = db.Exec(
			"INSERT INTO mit_classes(subjectID, sectionID, title, units, sections, custom, userID) VALUES(?, ?, ?, ?, ?, ?, ?)",
			subjectNumber,
			"",
			name,
			0,
			"",
			1,
			user.ID,
		)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"status": "ok",
		}, nil
	} else if methodName == "removeCustomClass" {
		subjectNumberInterface, ok := methodParams["subjectNumber"]
		if !ok {
			return map[string]interface{}{
				"status": "error",
				"error":  "missing_params",
			}, nil
		}

		subjectNumber, ok := subjectNumberInterface.(string)
		if !ok {
			return map[string]interface{}{
				"status": "error",
				"error":  "invalid_params",
			}, nil
		}

		subjectNumber = strings.TrimSpace(subjectNumber)
		subjectNumber = strings.ToUpper(subjectNumber)

		// find the user's custom registration of the class
		rows, err := db.Query("SELECT subjectID FROM mit_classes WHERE subjectID = ? AND custom = 1 AND userID = ?", subjectNumber, user.ID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if !rows.Next() {
			return map[string]interface{}{
				"status": "error",
				"error":  "invalid_params",
				"hint":   "Class does not exist for user.",
			}, nil
		}

		// it exists, we can delete it
		_, err = db.Exec("DELETE FROM mit_classes WHERE subjectID = ? AND custom = 1 AND userID = ?", subjectNumber, user.ID)
		if err != nil {
			return nil, err
		}

		return map[string]interface{}{
			"status": "ok",
		}, nil
	}

	return map[string]interface{}{
		"status": "error",
		"error":  "invalid_params",
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
