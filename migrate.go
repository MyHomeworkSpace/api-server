package main

import (
	"database/sql"
	"log"
	"strings"
)

// don't change these!
var defaultColors = []string{
	"ff4d40",
	"ffa540",
	"40ff73",
	"4071ff",
	"ff4086",
	"40ccff",
	"5940ff",
	"ff40f5",
	"a940ff",
	"e6ab68",
	"4d4d4d",
}

func handleMigrateError(err error) {
	log.Fatalln(err)
}

func migrateClasses(tx *sql.Tx) error {
	classCountByUser := map[int]int{}

	classes, err := DB.Query("SELECT id, name, COALESCE(teacher, ''), COALESCE(color, ''), userId FROM classes")
	if err != nil {
		return err
	}

	for classes.Next() {
		id, name, teacher, color, userID := -1, "", "", "", -1
		err = classes.Scan(&id, &name, &teacher, &color, &userID)
		if err != nil {
			return err
		}

		sortIndex, ok := classCountByUser[userID]
		if !ok {
			classCountByUser[userID] = 0
		}
		classCountByUser[userID]++

		if color == "" {
			// set an explicit color instead of inferring it
			color = defaultColors[id%len(defaultColors)]
		}

		// update teacher, color, and sort index
		_, err = tx.Exec(
			"UPDATE classes SET teacher = ?, color = ?, sortIndex = ? WHERE id = ?",
			teacher, strings.ToLower(color), sortIndex, id,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func migrate(name string) {
	log.Printf("Starting migration '%s'...", name)

	if name == "classes" {
		tx, err := DB.Begin()
		if err != nil {
			handleMigrateError(err)
		}

		err = migrateClasses(tx)
		if err != nil {
			tx.Rollback()
			handleMigrateError(err)
		}

		err = tx.Commit()
		if err != nil {
			handleMigrateError(err)
		}
	} else {
		log.Fatalf("Unknown migration name!")
	}

	log.Println("Done!")
}
