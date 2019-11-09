package tasks

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/MyHomeworkSpace/api-server/mit"
	"github.com/MyHomeworkSpace/api-server/util"

	"github.com/MyHomeworkSpace/api-server/config"
)

// some classes have weird times and aren't on the catalog, so we just give up on them
var skipClasses = []string{"15.830", "24.A03"}

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
	if source != "catalog" && source != "offerings" {
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

	// TODO: don't need this to be manually set
	params.Add("termCode", mitConfig.CurrentTermCode)
	params.Add("academicYear", mitConfig.CurrentTermCode[:4])

	// TODO: remove
	params.Add("lastUpdateDate", "2018-01-01")

	requestURL := mitConfig.DataProxyURL + "fetch?" + params.Encode()

	client := &http.Client{}
	request, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return taskResponse{}, err
	}

	request.Header.Add("X-MHS-Auth", mitConfig.DataProxyToken)

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
				rowsAffected += 1
			}
		}
	} else if source == "offerings" {
		offerings := []subjectOffering{}
		err = json.NewDecoder(response.Body).Decode(&offerings)
		if err != nil {
			return taskResponse{}, err
		}

		for _, offering := range offerings {
			if offering.Time == "" {
				continue
			}

			if util.StringSliceContains(skipClasses, offering.ID) {
				continue
			}

			termInfo, err := mit.GetTermByCode(offering.Term)
			if err != nil {
				return taskResponse{}, err
			}

			// check that we can parse the time info
			_, err = mit.ParseTimeInfo(offering.Time, termInfo)
			if err != nil {
				return taskResponse{}, err
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
				rowsAffected += 1
			}
		}
	}

	return taskResponse{
		RowsAffected: rowsAffected,
	}, tx.Commit()
}
