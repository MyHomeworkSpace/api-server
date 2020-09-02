package dalton

import (
	"strconv"
)

func getUserGrade(username string) (AnnouncementGrade, error) {
	if len(username) < 4 {
		// the username is not in the cXXyy format
		// this is probably a faculty member
		return AnnouncementGradeFaculty, nil
	}
	yearInfoString := username[1:3]
	yearInfo, err := strconv.Atoi(yearInfoString)
	if err != nil {
		// the username is not in the cXXyy format
		// this is probably a faculty member
		return AnnouncementGradeFaculty, nil
	}

	differenceFromBase := (yearInfo - 19) * -1
	grade := Grade_ClassOf2019 + AnnouncementGrade(differenceFromBase)

	return grade, nil
}

func getAnnouncementGroupSQL(groups []AnnouncementGrade) string {
	sql := ""
	first := true
	for _, group := range groups {
		if !first {
			sql += " OR "
		} else {
			first = false
		}
		// this is trusted input, and limited to integers, and so it is not vulnerable to SQL injection
		sql += "dalton_announcements.grade = "
		sql += strconv.Itoa(int(group))
	}
	return sql
}

func getGradeAnnouncementGroups(grade AnnouncementGrade) []AnnouncementGrade {
	groups := []AnnouncementGrade{AnnouncementGradeAll, grade}
	if grade < 9 {
		groups = append(groups, AnnouncementGradeMiddleSchool)
	}
	if grade >= 4 && grade <= 6 {
		groups = append(groups, AnnouncementGradeMiddleSchool456)
	}
	if grade >= 7 && grade <= 8 {
		groups = append(groups, AnnouncementGradeMiddleSchool78)
	}
	if grade >= 9 {
		groups = append(groups, AnnouncementGradeHighSchool)
	}
	return groups
}
