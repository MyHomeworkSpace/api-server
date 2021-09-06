package dalton

import (
	"database/sql"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/MyHomeworkSpace/api-server/data"
	"github.com/MyHomeworkSpace/api-server/schools"
)

type term struct {
	ID     int    `json:"id"`
	TermID int    `json:"termId"`
	Name   string `json:"name"`
	UserID int    `json:"userId"`
}

type provider struct {
	schools.Provider
}

func getOffBlocksStartingBefore(db *sql.DB, before string, groupSQL string) ([]data.OffBlock, error) {
	// find the starts
	offBlockRows, err := db.Query("SELECT id, date, text, grade FROM dalton_announcements WHERE ("+groupSQL+") AND `type` = 2 AND `date` < ?", before)
	if err != nil {
		return nil, err
	}
	defer offBlockRows.Close()
	blocks := []data.OffBlock{}
	for offBlockRows.Next() {
		block := data.OffBlock{}
		err = offBlockRows.Scan(&block.StartID, &block.StartText, &block.Name, &block.Grade)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}

	// find the matching ends
	for i, block := range blocks {
		offBlockEndRows, err := db.Query("SELECT date FROM dalton_announcements WHERE ("+groupSQL+") AND `type` = 3 AND `text` = ? AND `date` > ?", block.Name, block.StartText)
		if err != nil {
			return nil, err
		}
		defer offBlockEndRows.Close()
		if offBlockEndRows.Next() {
			err = offBlockEndRows.Scan(&blocks[i].EndText)
			if err != nil {
				return nil, err
			}
		}
	}

	// parse dates
	for i, block := range blocks {
		blocks[i].Start, err = time.Parse("2006-01-02", block.StartText)
		if err != nil {
			return nil, err
		}
		blocks[i].End, err = time.Parse("2006-01-02", block.EndText)
		if err != nil {
			return nil, err
		}
	}

	return blocks, err
}

func (p *provider) GetData(db *sql.DB, user *data.User, location *time.Location, startTime time.Time, endTime time.Time, dataType data.ProviderDataType) (data.ProviderData, error) {
	result := data.ProviderData{
		Announcements: nil,
		Events:        nil,
	}

	dayCount := int((endTime.Sub(startTime).Hours() / 24) + 0.5)

	school := (p.Provider.School).(*school)

	// get user's grade
	grade, err := getUserGrade(school.username)
	if err != nil {
		return data.ProviderData{}, err
	}

	// get user's announcement groups
	announcementGroups := getGradeAnnouncementGroups(grade)
	announcementGroupsSQL := getAnnouncementGroupSQL(announcementGroups)

	// get all rotation information for time period
	rotationRows, err := db.Query("SELECT * FROM dalton_rotations WHERE date >= ? AND date <= ?", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
	if err != nil {
		return data.ProviderData{}, err
	}
	rotations := []data.PlannerFriday{}
	for rotationRows.Next() {
		rotation := data.PlannerFriday{}
		err = rotationRows.Scan(&rotation.ID, &rotation.Date, &rotation.Index)
		if err != nil {
			return data.ProviderData{}, err
		}
		rotations = append(rotations, rotation)
	}
	rotationRows.Close()

	// get announcements for time period
	announcementRows, err := db.Query("SELECT id, date, text, grade, `type` FROM dalton_announcements WHERE date >= ? AND date <= ? AND ("+announcementGroupsSQL+") AND type < 2", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
	if err != nil {
		return data.ProviderData{}, err
	}
	defer announcementRows.Close()
	announcements := []data.PlannerAnnouncement{}
	for announcementRows.Next() {
		resp := data.PlannerAnnouncement{}
		err = announcementRows.Scan(&resp.ID, &resp.Date, &resp.Text, &resp.Grade, &resp.Type)
		if err != nil {
			return data.ProviderData{}, err
		}
		announcements = append(announcements, resp)
	}

	// get off blocks for time period
	offBlocks, err := getOffBlocksStartingBefore(db, endTime.Format("2006-01-02"), announcementGroupsSQL)
	if err != nil {
		return data.ProviderData{}, err
	}

	// generate list of all off days in time period
	offDays := []string{}

	for _, announcement := range announcements {
		if announcement.Type == data.AnnouncementTypeFullOff {
			offDays = append(offDays, announcement.Date)
		}
	}

	if dataType&data.ProviderDataAnnouncements != 0 {
		result.Announcements = []data.PlannerAnnouncement{}

		if CurrentMode != SchoolModeVirtualModified {
			// add rotations as announcements
			for _, rotation := range rotations {
				date, err := time.Parse("2006-01-02", rotation.Date)
				if err != nil {
					return data.ProviderData{}, err
				}

				rotationAnnouncement := data.PlannerAnnouncement{
					ID:    -1,
					Date:  rotation.Date,
					Text:  date.Weekday().String() + " " + strconv.Itoa(rotation.Index),
					Grade: -1,
					Type:  0,
				}
				result.Announcements = append(result.Announcements, rotationAnnouncement)
			}
		}

		// add exception day announcements
		currentDay := startTime
		for i := 0; i < dayCount; i++ {
			if i != 0 {
				currentDay = currentDay.Add(24 * time.Hour)
			}

			dayString := currentDay.Format("2006-01-02")

			overrideDay, isException := ExceptionDays[dayString]
			if isException {
				result.Announcements = append(result.Announcements, data.PlannerAnnouncement{
					ID:    -1,
					Date:  dayString,
					Text:  overrideDay.String() + " Class Schedule",
					Grade: -1,
					Type:  0,
				})
			}
		}

		// add standard announcements
		result.Announcements = append(result.Announcements, announcements...)

		for _, offBlock := range offBlocks {
			offDayCount := int(math.Ceil(offBlock.End.Sub(offBlock.Start).Hours() / 24))
			offDay := offBlock.Start
			result.Announcements = append(result.Announcements, data.PlannerAnnouncement{
				ID:    offBlock.StartID,
				Date:  offBlock.StartText,
				Text:  "Start of " + offBlock.Name,
				Grade: offBlock.Grade,
				Type:  data.AnnouncementTypeBreakStart,
			})
			for i := 0; i < offDayCount; i++ {
				if i != 0 {
					result.Announcements = append(result.Announcements, data.PlannerAnnouncement{
						ID:    offBlock.StartID,
						Date:  offDay.Format("2006-01-02"),
						Text:  offBlock.Name,
						Grade: offBlock.Grade,
						Type:  data.AnnouncementTypeBreakStart,
					})
				}
				offDays = append(offDays, offDay.Format("2006-01-02"))
				offDay = offDay.Add(24 * time.Hour)
			}
			result.Announcements = append(result.Announcements, data.PlannerAnnouncement{
				ID:    offBlock.EndID,
				Date:  offBlock.EndText,
				Text:  "End of " + offBlock.Name,
				Grade: offBlock.Grade,
				Type:  data.AnnouncementTypeBreakEnd,
			})
		}
	}

	if dataType&data.ProviderDataEvents != 0 {
		result.Events = []data.Event{}

		// get terms for user
		termRows, err := db.Query("SELECT id, termId, name, userId FROM dalton_terms WHERE userId = ? ORDER BY name ASC", user.ID)
		if err != nil {
			return data.ProviderData{}, err
		}
		defer termRows.Close()
		availableTerms := []term{}
		for termRows.Next() {
			term := term{}
			err = termRows.Scan(&term.ID, &term.TermID, &term.Name, &term.UserID)
			if err != nil {
				return data.ProviderData{}, err
			}
			availableTerms = append(availableTerms, term)
		}

		// get schedule events
		currentDay := startTime
		for i := 0; i < dayCount; i++ {
			if i != 0 {
				currentDay = currentDay.Add(24 * time.Hour)
			}

			dayString := currentDay.Format("2006-01-02")

			var currentTerm *term

			for _, termToCheck := range availableTerms {
				dates := TermMap[termToCheck.Name]
				if dates == nil {
					// this is an old schedule, just ignore it
					continue
				}

				startDate, endDate := dates[0], dates[1]
				if currentDay.Add(time.Second).After(startDate) && currentDay.Before(endDate) {
					// special exception: in normal mode, seniors stop having classes at a certain point
					if CurrentMode == SchoolModeVirtual || CurrentMode == SchoolModeVirtualModified {
						// if user is a senior, their classes end earlier
						if grade == 12 && currentDay.After(Day_SeniorEnd) {
							currentTerm = nil
							break
						}
					}

					currentTerm = &termToCheck
					break
				}
			}

			if currentTerm != nil {
				dayTime, _ := time.ParseInLocation("2006-01-02", dayString, location)
				dayOffset := int(dayTime.Unix())

				// check if it's an off day
				isOff := false

				for _, offDay := range offDays {
					if dayString == offDay {
						isOff = true
						break
					}
				}

				if isOff {
					continue
				}

				// calculate day index (1 = monday, 8 = rotation 4)
				dayNumber := int(dayTime.Weekday())

				overrideDay, isException := ExceptionDays[dayString]
				if isException {
					dayNumber = int(overrideDay)
				}

				rotationDay := time.Friday
				if CurrentMode == SchoolModeVirtual {
					rotationDay = time.Wednesday
				}

				if !isException && dayTime.Weekday() == rotationDay && CurrentMode != SchoolModeVirtualModified {
					rotationNumber := -1
					for _, rotation := range rotations {
						if dayString == rotation.Date {
							rotationNumber = rotation.Index
							break
						}
					}

					if rotationNumber == 1 {
						// the first rotation gets the normal day number
						// this is by default for friday, but needs to explicitly stated for virtual school, with wednesday rotations
						dayNumber = int(rotationDay)
					} else if rotationNumber != -1 {
						dayNumber = 4 + rotationNumber
					} else {
						continue
					}
				}

				if !isException && dayTime.Weekday() == time.Saturday || dayTime.Weekday() == time.Sunday {
					continue
				}

				rows, err := db.Query("SELECT dalton_periods.id, dalton_classes.termId, dalton_classes.sectionId, dalton_classes.`name`, dalton_classes.ownerId, dalton_classes.ownerName, dalton_periods.dayNumber, dalton_periods.block, dalton_periods.buildingName, dalton_periods.roomNumber, dalton_periods.`start`, dalton_periods.`end`, dalton_periods.userId FROM dalton_periods INNER JOIN dalton_classes ON dalton_periods.classId = dalton_classes.sectionId WHERE dalton_periods.userId = ? AND dalton_classes.userId = ? AND (dalton_classes.termId = ? OR dalton_classes.termId = -1) AND dalton_periods.dayNumber = ? GROUP BY dalton_periods.id, dalton_classes.termId, dalton_classes.name, dalton_classes.ownerId, dalton_classes.ownerName", user.ID, user.ID, currentTerm.TermID, dayNumber)
				if err != nil {
					return data.ProviderData{}, err
				}
				defer rows.Close()
				for rows.Next() {
					event := data.Event{
						Tags: map[data.EventTagType]interface{}{},
					}

					termID, classID, ownerID, ownerName, dayNumber, block, buildingName, roomNumber := -1, -1, -1, "", -1, "", "", ""

					err = rows.Scan(&event.ID, &termID, &classID, &event.Name, &ownerID, &ownerName, &dayNumber, &block, &buildingName, &roomNumber, &event.Start, &event.End, &event.UserID)
					if err != nil {
						return data.ProviderData{}, err
					}

					event.UniqueID = strconv.Itoa(event.ID) + "-" + strconv.Itoa(classID) + "-" + dayString

					event.Tags[data.EventTagShortName] = strings.TrimSpace(strings.SplitN(event.Name, " - ", 2)[0])
					event.Tags[data.EventTagReadOnly] = true
					event.Tags[data.EventTagCancelable] = true
					event.Tags[data.EventTagTermID] = termID
					event.Tags[data.EventTagClassID] = classID
					event.Tags[data.EventTagOwnerID] = ownerID
					event.Tags[data.EventTagOwnerName] = ownerName
					event.Tags[data.EventTagDayNumber] = dayNumber
					event.Tags[data.EventTagBlock] = block
					event.Tags[data.EventTagBuildingName] = buildingName
					event.Tags[data.EventTagRoomNumber] = roomNumber
					event.Tags[data.EventTagHideBuildingName] = true

					event.Start += dayOffset
					event.End += dayOffset

					result.Events = append(result.Events, event)
				}

				if dayTime.Weekday() == time.Thursday {
					// special case: assembly
					for eventIndex, event := range result.Events {
						// check for an "HS House" event
						// starting 11:50, ending 12:50
						if strings.HasPrefix(event.Name, "HS House") && event.Start == int(dayTime.Unix())+42600 && event.End == int(dayTime.Unix())+46200 {
							// found it
							// now look up what type of assembly period it is this week
							assemblyType, foundType := AssemblyTypeList[dayTime.Format("2006-01-02")]

							if !foundType || assemblyType == AssemblyTypeAssembly {
								// set name to assembly and room to Theater
								result.Events[eventIndex].Name = "Assembly"
								result.Events[eventIndex].Tags[data.EventTagRoomNumber] = "Theater"
							} else if assemblyType == AssemblyTypeLongHouse {
								// set name to long house
								result.Events[eventIndex].Name = "Long House"
							} else if assemblyType == AssemblyTypeLab {
								// just remove it
								result.Events = append(result.Events[:eventIndex], result.Events[eventIndex+1:]...)
							}
						}
					}
				}
			}
		}
	}

	return result, nil
}
