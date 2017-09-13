package api

import (
	"errors"
	"strconv"
	"time"
)

var (
	ErrDataBadUsername = errors.New("data: bad username")
	ErrDataNotFound    = errors.New("data: not found")
)

type FacultyInfo struct {
	BlackbaudUserID     int    `json:"bbUserId"`
	FirstName           string `json:"firstName"`
	LastName            string `json:"lastName"`
	LargeFileName       string `json:"largeFileName"`
	DepartmentDisplay   string `json:"departmentDisplay"`
	GradeNumericDisplay string `json:"gradeNumericDisplay"`
}

type OffBlock struct {
	StartID   int       `json:"startId"`
	EndID     int       `json:"endId"`
	Start     time.Time `json:"-"`
	End       time.Time `json:"-"`
	StartText string    `json:"start"`
	EndText   string    `json:"end"`
	Name      string    `json:"name"`
	Grade     int       `json:"grade"`
}

type User struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Username           string `json:"username"`
	Email              string `json:"email"`
	Type               string `json:"type"`
	Features           string `json:"features"`
	Level              int    `json:"level"`
	ShowMigrateMessage int    `json:"showMigrateMessage"`
}

func Data_GetOffBlocksStartingBefore(before string, groups []int) ([]OffBlock, error) {
	// find the starts
	offBlockRows, err := DB.Query("SELECT id, date, text, grade FROM announcements WHERE ("+Data_GetAnnouncementGroupSQL(groups)+") AND `type` = 2 AND `date` < ?", before)
	if err != nil {
		return nil, err
	}
	defer offBlockRows.Close()
	blocks := []OffBlock{}
	for offBlockRows.Next() {
		block := OffBlock{}
		offBlockRows.Scan(&block.StartID, &block.StartText, &block.Name, &block.Grade)
		blocks = append(blocks, block)
	}

	// find the matching ends
	for i, block := range blocks {
		offBlockEndRows, err := DB.Query("SELECT date FROM announcements WHERE ("+Data_GetAnnouncementGroupSQL(groups)+") AND `type` = 3 AND `text` = ?", block.Name)
		if err != nil {
			return nil, err
		}
		defer offBlockEndRows.Close()
		if offBlockEndRows.Next() {
			offBlockEndRows.Scan(&blocks[i].EndText)
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

func Data_GetAnnouncementGroupSQL(groups []int) string {
	sql := ""
	first := true
	for _, group := range groups {
		if !first {
			sql += " OR "
		} else {
			first = false
		}
		// this is trusted input, and limited to integers, and so it is not vulnerable to SQL injection
		sql += "announcements.grade = "
		sql += strconv.Itoa(group)
	}
	return sql
}

func Data_GetGradeAnnouncementGroups(grade int) []int {
	groups := []int{AnnouncementGrade_All, grade}
	if grade < 9 {
		groups = append(groups, AnnouncementGrade_MiddleSchool)
	}
	if grade >= 4 && grade <= 6 {
		groups = append(groups, AnnouncementGrade_MiddleSchool_456)
	}
	if grade >= 7 && grade <= 8 {
		groups = append(groups, AnnouncementGrade_MiddleSchool_78)
	}
	if grade >= 9 {
		groups = append(groups, AnnouncementGrade_HighSchool)
	}
	return groups
}

func Data_GetUserByID(id int) (User, error) {
	rows, err := DB.Query("SELECT id, name, username, email, type, features, level, showMigrateMessage FROM users WHERE id = ?", id)
	if err != nil {
		return User{}, err
	}
	defer rows.Close()
	if rows.Next() {
		user := User{}
		err := rows.Scan(&user.ID, &user.Name, &user.Username, &user.Email, &user.Type, &user.Features, &user.Level, &user.ShowMigrateMessage)
		if err != nil {
			return User{}, err
		}
		return user, nil
	} else {
		return User{}, ErrDataNotFound
	}
}

func Data_GetUserGrade(user User) (int, error) {
	if len(user.Username) < 4 {
		// the username is not in the cXXyy format
		// this is probably a faculty member
		return AnnouncementGrade_Faculty, nil
	}
	yearInfoString := user.Username[1:3]
	yearInfo, err := strconv.Atoi(yearInfoString)
	if err != nil {
		// the username is not in the cXXyy format
		// this is probably a faculty member
		return AnnouncementGrade_Faculty, nil
	}

	differenceFromBase := (yearInfo - 19) * -1
	grade := Grade_ClassOf2019 + differenceFromBase

	return grade, nil
}
