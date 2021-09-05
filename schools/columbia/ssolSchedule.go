package columbia

import (
	"database/sql"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/PuerkitoBio/goquery"
)

func (s *school) parseSSOLSchedulePage(tx *sql.Tx, user *data.User, doc *goquery.Document) (string, error) {
	dataGrids := doc.Find(".DataGrid")

	// first data grid is user info (we just need the name)
	// second data grid is schedule by class
	// third data grid is schedule by day

	// get name from first data grid
	studentName := dataGrids.Eq(0).Find(".clsDataGridTitle td").Text()
	studentName = strings.TrimSpace(studentName)
	studentName = strings.ReplaceAll(studentName, "   ", " ")

	// get class list from second data grid
	classRows := dataGrids.Eq(1).Find("tr")

	classInsertStmt, err := tx.Prepare("INSERT INTO columbia_classes(department, number, section, name, instructorName, instructorEmail, userID) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return "", err
	}
	defer classInsertStmt.Close()

	classInfoMap := map[string][]string{}

	// first three rows and last rows are to be discarded
	for i := 3; i < classRows.Length()-1; i++ {
		classRowColumns := classRows.Eq(i).Find("td")

		// class, grading, instructor, day, time/location, start/end
		// we just need class and instructor

		classCell := classRowColumns.Eq(0)

		// the raw text tells us department and number
		classCellText := strings.ReplaceAll(strings.TrimSpace(classCell.Text()), "\u00A0", " ")
		classParts := strings.Split(classCellText, " ")
		classDepartment := classParts[0]
		classNumber := classParts[1]
		classSection := "" // TODO

		// the text in the font tag tells us the name
		className := strings.TrimSpace(classCell.Find("font").Text())

		instructorCell := classRowColumns.Eq(2)

		// the text in the a tag tells us the instructor email
		classInstructorEmail := strings.TrimSpace(instructorCell.Find("a").Text())

		// the raw text tells us instructor name
		// this is a terrible hack and i'm sorry
		classInstructorName := strings.ReplaceAll(strings.TrimSpace(instructorCell.Text()), "\u00A0", " ")
		classInstructorName = strings.ReplaceAll(classInstructorName, classInstructorEmail, "")

		// set a mapping so that the meeting code can insert this info directly into the row
		classInfoMap[className] = []string{classDepartment, classNumber, classSection}

		// add to db
		_, err = classInsertStmt.Exec(
			classDepartment,
			classNumber,
			classSection,
			className,
			classInstructorName,
			classInstructorEmail,
			user.ID,
		)
		if err != nil {
			return "", err
		}
	}

	// get meeting times from third data grid
	scheduleRows := dataGrids.Eq(2).Find("tr")

	meetingInsertStmt, err := tx.Prepare("INSERT INTO columbia_meetings(department, number, section, name, building, room, dow, start, end, beginDate, endDate, userID) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return "", err
	}
	defer meetingInsertStmt.Close()

	// first two rows are to be discarded
	currentDow := time.Sunday
	for i := 2; i < scheduleRows.Length(); i++ {
		// each row has three possibilities:
		// 1. it's the first row of a day of the week -> 5 columns
		// 2. it's not the first row for that day of the week -> 4 columns
		// 3. it's a separator between days of the week -> 1 column
		scheduleRowColumns := scheduleRows.Eq(i).Find("td")
		switch scheduleRowColumns.Length() {
		case 5:
			// first class of the day of week
			dowMap := map[string]time.Weekday{
				"Sun": time.Sunday,
				"Mon": time.Monday,
				"Tue": time.Tuesday,
				"Wed": time.Wednesday,
				"Thr": time.Thursday,
				"Fri": time.Friday,
				"Sat": time.Saturday,
			}
			dowText := strings.TrimSpace(scheduleRowColumns.Eq(0).Text())
			currentDow = dowMap[dowText]

			fallthrough

		case 4:
			// class meeting time
			columnOffset := scheduleRowColumns.Length() - 4

			timeText := strings.TrimSpace(scheduleRowColumns.Eq(columnOffset).Text())
			locationText := strings.TrimSpace(scheduleRowColumns.Eq(columnOffset + 1).Text())
			classNameText := strings.TrimSpace(scheduleRowColumns.Eq(columnOffset + 2).Text())
			termText := strings.TrimSpace(scheduleRowColumns.Eq(columnOffset + 3).Text())

			// parse location into room and building
			locationParts := strings.Split(locationText, "\u00A0")
			roomText := strings.TrimSpace(locationParts[0])
			buildingText := ""
			if len(locationParts) > 1 {
				buildingText = strings.TrimSpace(locationParts[1])
			}

			// parse start and end times
			timeParts := strings.Split(timeText, "-")
			startTime, err := time.Parse("3:04pm", timeParts[0])
			if err != nil {
				return "", err
			}
			endTime, err := time.Parse("3:04pm", timeParts[1])
			if err != nil {
				return "", err
			}

			startTime = startTime.AddDate(1970, 0, 0)
			endTime = endTime.AddDate(1970, 0, 0)

			// parse begin and end dates
			termParts := strings.Split(termText, "-")
			beginDate, err := time.Parse("01/02/06", termParts[0])
			if err != nil {
				return "", err
			}
			endDate, err := time.Parse("01/02/06", termParts[1])
			if err != nil {
				return "", err
			}

			// get info from map
			classInfo := classInfoMap[classNameText]

			// add to db
			_, err = meetingInsertStmt.Exec(
				classInfo[0],
				classInfo[1],
				classInfo[2],
				classNameText,
				buildingText,
				roomText,
				currentDow,
				startTime.Unix(),
				endTime.Unix(),
				beginDate.Format("2006-01-02"),
				endDate.Format("2006-01-02"),
				user.ID,
			)
			if err != nil {
				return "", err
			}

		case 1:
			// separator
		}
	}

	return studentName, nil
}
