package tasks

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/mit"
	"github.com/MyHomeworkSpace/api-server/util"

	"github.com/MyHomeworkSpace/api-server/config"
)

// some classes have weird times and aren't on the catalog, so we just give up on them
var skipClasses = map[string][]string{
	"2021SP": {"20.051", "21M.442"},
	"2022FA": {"15.S24", "15.830", "21M.138", "21M.460"},
	"2022SP": {"8.962"},
}

type catalogListing struct {
	ID         string `json:"id"`
	ShortTitle string `json:"short"`
	Title      string `json:"title"`

	OfferedFall   bool `bool:"fall"`
	OfferedIAP    bool `bool:"iap"`
	OfferedSpring bool `bool:"spring"`

	FallInstructors   string `json:"fallI"`
	SpringInstructors string `json:"springI"`
}

type subjectOffering struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Section string `json:"section"`
	Term    string `json:"term"`

	Time  string `json:"time"`
	Place string `json:"place"`

	FacultyID   string `json:"facultyID"`
	FacultyName string `json:"facultyName"`

	IsFake   bool `json:"fake"`
	IsMaster bool `json:"master"`

	IsDesign     bool `json:"design"`
	IsLab        bool `json:"lab"`
	IsLecture    bool `json:"lecture"`
	IsRecitation bool `json:"recitation"`
}

// StartImportFromMIT begins an import of the given data from the MIT Data Warehouse.
func StartImportFromMIT(source string, db *sql.DB) error {
	if source != "catalog" && source != "coursews" && source != "offerings" {
		return errors.New("tasks: invalid parameter")
	}

	taskID := "mit_" + source
	taskName := "MIT Import - " + source

	go taskWatcher(taskID, taskName, importFromMIT, source, db)
	return nil
}

func importFromMIT(lastCompletion *time.Time, source string, db *sql.DB) (taskResponse, error) {
	mitConfig := config.GetCurrent().MIT
	params := url.Values{}

	params.Add("source", source)

	currentTerm := mit.GetCurrentTerm()
	params.Add("termCode", currentTerm.Code)
	params.Add("academicYear", currentTerm.Code[:4])

	// TODO: remove
	params.Add("lastUpdateDate", "2019-01-01")

	requestURL := mitConfig.DataProxyURL + "fetch?" + params.Encode()
	if source == "coursews" {
		// actually use the secret coursews API
		requestURL = "https://coursews.mit.edu/coursews/"
	}

	client := &http.Client{}
	request, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return taskResponse{}, err
	}

	if source != "coursews" {
		request.Header.Add("X-MHS-Auth", mitConfig.ProxyToken)
	}

	response, err := client.Do(request)
	if err != nil {
		return taskResponse{}, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return taskResponse{}, err
		}

		return taskResponse{}, fmt.Errorf(
			"tasks: MIT data server returned status code %d, body: '%s'",
			response.StatusCode,
			string(bodyBytes),
		)
	}

	tx, err := db.Begin()
	if err != nil {
		return taskResponse{}, err
	}

	rowsAffected := int64(0)

	if source == "catalog" {
		listings := []catalogListing{}
		err = json.NewDecoder(response.Body).Decode(&listings)
		if err != nil {
			return taskResponse{}, err
		}

		for _, listing := range listings {
			result, err := tx.Exec(
				`INSERT INTO
					mit_listings(id, shortTitle, title, offeredFall, offeredIAP, offeredSpring, fallInstructors, springInstructors)
					VALUES(?, ?, ?, ?, ?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE
					id = VALUES(id),
					shortTitle = VALUES(shortTitle),
					title = VALUES(title),
					offeredFall = VALUES(offeredFall),
					offeredIAP = VALUES(offeredIAP),
					offeredSpring = VALUES(offeredSpring),
					fallInstructors = VALUES(fallInstructors),
					springInstructors = VALUES(springInstructors)
				`,
				listing.ID,
				listing.ShortTitle,
				listing.Title,
				listing.OfferedFall,
				listing.OfferedIAP,
				listing.OfferedSpring,
				listing.FallInstructors,
				listing.SpringInstructors,
			)
			if err != nil {
				return taskResponse{}, err
			}

			rowAffected, err := result.RowsAffected()
			if err != nil {
				return taskResponse{}, err
			}
			if rowAffected > 0 {
				rowsAffected++
			}
		}
	} else if source == "coursews" {
		// unfortunately, this API gives us data in a rather annoying format
		// so we cannot just parse them into a go struct
		// instead we have to use a hacky series of typecasts :(
		wsData := map[string]interface{}{}
		err = json.NewDecoder(response.Body).Decode(&wsData)
		if err != nil {
			return taskResponse{}, err
		}

		items := wsData["items"].([]interface{})

		// assume it's all the same term
		termInfo, err := mit.GetTermByCode(currentTerm.Code)
		if err != nil {
			return taskResponse{}, err
		}

		// first, clear out any data from a previous term
		_, err = tx.Exec("DELETE FROM mit_offerings WHERE term <> ?", currentTerm.Code)
		if err != nil {
			return taskResponse{}, err
		}

		for _, itemInterface := range items {
			item := itemInterface.(map[string]interface{})

			itemType := item["type"].(string)

			if itemType == "Class" {
				// we actually don't care about classes
				continue
			}

			if itemType != "LectureSession" && itemType != "LabSession" && itemType != "RecitationSession" {
				return taskResponse{}, fmt.Errorf("tasks: unknown coursews item type '%s'", itemType)
			}

			itemLabel := item["label"].(string)
			itemSectionOf := item["section-of"].(string)
			itemTimeAndPlace := item["timeAndPlace"].(string)

			if itemTimeAndPlace == "null null" {
				// this record tells us absolutely nothing
				// ignore it
				continue
			}

			// for some reason, the labels are the section + class number joined together
			// for example, section L01 of 6.003 has a label of "L016.003"
			// parse out the section ID
			sectionID := strings.Replace(itemLabel, itemSectionOf, "", -1)

			// in order to maximize suffering, the coursews api also combines the time and place fields
			// examples of this include "MW9.30-11 4-251" (easy)
			// or "TR9-11 (MEETS 4/7 TO 5/14) MEC-209" and "M EVE (6-8 PM) BOSTON PRE-REL" (why??)
			// the high quality algorithm to parse this is to break the string into spaces, and keep removing words until it works
			timeAndPlaceParts := strings.Split(itemTimeAndPlace, " ")
			currentTimeString := ""
			parsed := false
			for i := len(timeAndPlaceParts); i > 0; i-- {
				currentTimeString = ""
				for j := 0; j < i; j++ {
					if j != 0 {
						currentTimeString += " "
					}
					currentTimeString += timeAndPlaceParts[j]
				}

				// attempt
				_, err = mit.ParseTimeInfo(currentTimeString, termInfo)
				if err != nil {
					// oofie
					continue
				}

				// we survived!
				parsed = true
				break
			}

			if !parsed {
				return taskResponse{}, fmt.Errorf("tasks: failed to parse timeAndPlace '%s'", itemTimeAndPlace)
			}

			time := currentTimeString
			place := strings.TrimSpace(strings.Replace(itemTimeAndPlace, currentTimeString, "", -1))

			if place == "" {
				return taskResponse{}, fmt.Errorf("tasks: didn't get place from timeAndPlace '%s'", itemTimeAndPlace)
			}

			if place == "null" {
				place = ""
			}

			isDesign := false // design sections seem to not be included?
			isLab := (sectionID[0] == 'B')
			isLecture := (sectionID[0] == 'L')
			isRecitation := (sectionID[0] == 'R')

			// now, try to insert this new record
			// since we have very little info, we do NOT overwrite existing faculty/extra data if we have some
			result, err := tx.Exec(
				`INSERT INTO
					mit_offerings(id, title, section, term, time, place, facultyID, facultyName, isFake, isMaster, isDesign, isLab, isLecture, isRecitation)
					VALUES(?, '', ?, ?, ?, ?, '', '', 0, 0, ?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE
					id = VALUES(id),
					section = VALUES(section),
					term = VALUES(term),
					time = VALUES(time),
					place = VALUES(place),
					isDesign = VALUES(isDesign),
					isLab = VALUES(isLab),
					isLecture = VALUES(isLecture),
					isRecitation = VALUES(isRecitation)`,
				itemSectionOf,
				sectionID,
				currentTerm.Code,
				time,
				place,
				isDesign,
				isLab,
				isLecture,
				isRecitation,
			)
			if err != nil {
				return taskResponse{}, err
			}

			rowAffected, err := result.RowsAffected()
			if err != nil {
				return taskResponse{}, err
			}
			if rowAffected > 0 {
				rowsAffected++
			}
		}
	} else if source == "offerings" {
		offerings := []subjectOffering{}
		err = json.NewDecoder(response.Body).Decode(&offerings)
		if err != nil {
			return taskResponse{}, err
		}

		// first, clear out any data from a previous term
		_, err = tx.Exec("DELETE FROM mit_offerings WHERE term <> ?", currentTerm.Code)
		if err != nil {
			return taskResponse{}, err
		}

		for _, offering := range offerings {
			if offering.Time == "" {
				continue
			}

			if util.StringSliceContains(skipClasses[currentTerm.Code], offering.ID) {
				continue
			}

			termInfo, err := mit.GetTermByCode(offering.Term)
			if err != nil {
				return taskResponse{}, err
			}

			// check that we can parse the time info
			_, err = mit.ParseTimeInfo(offering.Time, termInfo)
			if err != nil {
				return taskResponse{}, fmt.Errorf("mit: failed to parse time of offering of %s (%s): %s", offering.ID, offering.Section, err.Error())
			}

			result, err := tx.Exec(
				`INSERT INTO
					mit_offerings(id, title, section, term, time, place, facultyID, facultyName, isFake, isMaster, isDesign, isLab, isLecture, isRecitation)
					VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE
					id = VALUES(id),
					title = VALUES(title),
					section = VALUES(section),
					term = VALUES(term),
					time = VALUES(time),
					place = VALUES(place),
					facultyID = VALUES(facultyID),
					facultyName = VALUES(facultyName),
					isFake = VALUES(isFake),
					isMaster = VALUES(isMaster),
					isDesign = VALUES(isDesign),
					isLab = VALUES(isLab),
					isLecture = VALUES(isLecture),
					isRecitation = VALUES(isRecitation)`,
				offering.ID,
				offering.Title,
				offering.Section,
				offering.Term,
				offering.Time,
				offering.Place,
				offering.FacultyID,
				offering.FacultyName,
				offering.IsFake,
				offering.IsMaster,
				offering.IsDesign,
				offering.IsLab,
				offering.IsLecture,
				offering.IsRecitation,
			)
			if err != nil {
				return taskResponse{}, err
			}

			rowAffected, err := result.RowsAffected()
			if err != nil {
				return taskResponse{}, err
			}
			if rowAffected > 0 {
				rowsAffected++
			}
		}
	}

	return taskResponse{
		RowsAffected: rowsAffected,
	}, tx.Commit()
}
