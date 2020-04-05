package external

import (
	"bufio"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Update redownloads all events on the calendar from the source URL.
func (p *Provider) Update(tx *sql.Tx) error {
	// get the data
	resp, err := http.Get(p.ExternalCalendarURL)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)

	// now we parse the data - this is somewhat brittle
	timezone := time.UTC
	externalEvents := []externalEvent{}
	currentEvent := externalEvent{}
	currentState := ""
	for scanner.Scan() {
		text := scanner.Text()
		parts := strings.Split(text, ":")
		command := strings.Split(parts[0], ";")[0]

		if command == "BEGIN" {
			if parts[1] == "VCALENDAR" && currentState == "" {
				currentState = parts[1]
			} else if parts[1] == "VTIMEZONE" && currentState == "VCALENDAR" {
				currentState = parts[1]
			} else if parts[1] == "VEVENT" && currentState == "VCALENDAR" {
				currentState = parts[1]
				currentEvent = externalEvent{}
			} else if parts[1] == "STANDARD" || parts[1] == "DAYLIGHT" {
				// we don't care, skip it
				lookFor := "END:" + parts[1]
				for scanner.Scan() {
					if scanner.Text() == lookFor {
						break
					}
				}
			} else {
				return fmt.Errorf("external: unexpected BEGIN '%s'", parts[1])
			}
		} else if command == "END" {
			if parts[1] == "VCALENDAR" && currentState == "VCALENDAR" {
				// we are done
				break
			} else if parts[1] == currentState && (parts[1] == "VTIMEZONE" || parts[1] == "VEVENT") {
				if parts[1] == "VEVENT" {
					// save the current event
					externalEvents = append(externalEvents, currentEvent)
				}
				currentState = "VCALENDAR"
			} else {
				return fmt.Errorf("external: unexpected END '%s'", parts[1])
			}
		} else if command == "CALSCALE" {
			if parts[1] != "GREGORIAN" {
				tx.Rollback()
				return fmt.Errorf("external: unexpected CALSCALE '%s'", parts[1])
			}
		} else if command == "TZID" {
			timezone, err = time.LoadLocation(parts[1])
			if err != nil {
				tx.Rollback()
				return err
			}
		} else if command == "UID" {
			currentEvent.UID = parts[1]
		} else if command == "DTSTART" || command == "DTEND" {
			if strings.Contains(text, "VALUE=DATE") {
				// it's a full day event
				// skip it
				// TODO: handle these
				for scanner.Scan() {
					if scanner.Text() == "END:VEVENT" {
						break
					}
				}
				currentState = "VCALENDAR"
			} else {
				// TODO: we currently ignore a provided TZID in favor of the overall VTIMEZONE
				resultTime, err := time.ParseInLocation("20060102T150405", parts[1], timezone)
				if err != nil {
					tx.Rollback()
					return err
				}

				resultUnix := resultTime.Unix()

				if command == "DTSTART" {
					currentEvent.Start = resultUnix
				} else if command == "DTEND" {
					currentEvent.End = resultUnix
				}
			}
		} else if command == "SUMMARY" {
			currentEvent.Name = strings.Replace(text, "SUMMARY:", "", 1)
		} else if command == "VERSION" || command == "METHOD" || command == "PRODID" || command == "X-LIC-LOCATION" || command == "DTSTAMP" || command == "STATUS" || command == "CLASS" || command == "PRIORITY" || command == "CATEGORIES" {
			// don't really care
		} else {
			tx.Rollback()
			return fmt.Errorf("external: unhandled line '%s'", text)
		}
	}

	// wipe anything that's currently there
	_, err = tx.Exec("DELETE FROM calendar_external_events WHERE calendarID = ?", p.ExternalCalendarID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// insert the new data
	stmt, err := tx.Prepare("INSERT INTO calendar_external_events(uid, name, start, end, calendarID) VALUES(?, ?, ?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, externalEvent := range externalEvents {
		_, err = stmt.Exec(
			externalEvent.UID,
			externalEvent.Name,
			externalEvent.Start,
			externalEvent.End,
			p.ExternalCalendarID,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// set the last updated date
	_, err = tx.Exec("UPDATE calendar_external SET lastUpdated = ? WHERE id = ?", time.Now().Unix(), p.ExternalCalendarID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
