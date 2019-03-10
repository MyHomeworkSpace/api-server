package api

import (
	"errors"
	"strconv"
)

var (
	ErrDataBadUsername = errors.New("data: bad username")
	ErrDataNotFound    = errors.New("data: not found")
)

type Feedback struct {
	ID            int    `json:"id"`
	UserID        int    `json:"userid"`
	Type          string `json:"type"`
	Text          string `json:"text"`
	Timestamp     string `json:"timestamp"`
	UserName      string `json:"userName"`
	UserEmail     string `json:"userEmail"`
	HasScreenshot bool   `json:"hasScreenshot"`
}

type Tab struct {
	ID     int    `json:"id"`
	Slug   string `json:"slug"`
	Icon   string `json:"icon"`
	Label  string `json:"label"`
	Target string `json:"target"`
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
	TwoFactorVerified  int    `json:"twoFactorVerified"`
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

func Data_GetPrefForUser(key string, userId int) (Pref, error) {
	rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", userId, key)
	if err != nil {
		return Pref{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return Pref{}, ErrDataNotFound
	}

	pref := Pref{}
	err = rows.Scan(&pref.ID, &pref.Key, &pref.Value)
	if err != nil {
		return Pref{}, err
	}

	return pref, nil
}

func Data_GetTabsByUserID(userId int) ([]Tab, error) {
	rows, err := DB.Query("SELECT tabs.id, tabs.slug, tabs.icon, tabs.label, tabs.target FROM tabs INNER JOIN tab_permissions ON tab_permissions.tabId = tabs.id WHERE tab_permissions.userId = ?", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tabs := []Tab{}
	for rows.Next() {
		tab := Tab{}
		err := rows.Scan(&tab.ID, &tab.Slug, &tab.Icon, &tab.Label, &tab.Target)
		if err != nil {
			return nil, err
		}
		tabs = append(tabs, tab)
	}
	return tabs, nil
}

func Data_GetUserByID(id int) (User, error) {
	rows, err := DB.Query("SELECT id, name, username, email, type, features, level, showMigrateMessage, twoFactorVerified FROM users WHERE id = ?", id)
	if err != nil {
		return User{}, err
	}
	defer rows.Close()
	if rows.Next() {
		user := User{}
		err := rows.Scan(&user.ID, &user.Name, &user.Username, &user.Email, &user.Type, &user.Features, &user.Level, &user.ShowMigrateMessage, &user.TwoFactorVerified)
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
