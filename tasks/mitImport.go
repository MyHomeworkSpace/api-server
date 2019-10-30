package tasks

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/MyHomeworkSpace/api-server/mit"
	"github.com/MyHomeworkSpace/api-server/util"

	"github.com/MyHomeworkSpace/api-server/config"
)

// some classes have weird times and aren't on the catalog, so we just give up on them
var skipClasses = []string{"15.830", "24.A03"}

type subjectOffering struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Section string `json:"section"`
	Term    string `json:"term"`

	Time  string `json:"time"`
	Place string `json:"place"`

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

func importFromMIT(lastCompletion *time.Time, source string, db *sql.DB) error {
	warehouseConfig := config.GetCurrent().MIT.Warehouse
	params := url.Values{}

	params.Add("source", source)

	// TODO: don't need this to be manually set
	params.Add("termCode", warehouseConfig.CurrentTermCode)

	requestURL := warehouseConfig.DataProxyURL + "fetch?" + params.Encode()

	response, err := http.Get(requestURL)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if source == "catalog" {
		return errors.New("tasks: catalog not supported yet")
	} else if source == "offerings" {
		offerings := []subjectOffering{}
		err = json.NewDecoder(response.Body).Decode(&offerings)
		if err != nil {
			return err
		}

		for _, offering := range offerings {
			if offering.Time == "" {
				continue
			}

			if util.StringSliceContains(skipClasses, offering.ID) {
				continue
			}

			// check that we can parse the time info
			_, err := mit.ParseAllTimeInfo(offering.Time)
			if err != nil {
				return err
			}

			_, err = tx.Exec(
				"REPLACE INTO mit_offerings(id, title, section, term, time, place, isFake, isMaster, isDesign, isLab, isLecture, isRecitation) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
				offering.ID,
				offering.Title,
				offering.Section,
				offering.Term,
				offering.Time,
				offering.Place,
				offering.IsFake,
				offering.IsMaster,
				offering.IsDesign,
				offering.IsLab,
				offering.IsLecture,
				offering.IsRecitation,
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}
