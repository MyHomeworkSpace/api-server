package dalton

import (
	"strconv"

	"github.com/MyHomeworkSpace/api-server/data"
)

func getUserGrade(user data.User) (int, error) {
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

func getAnnouncementGroupSQL(groups []int) string {
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

func getGradeAnnouncementGroups(grade int) []int {
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
