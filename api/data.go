package api

import (
	"errors"

	"github.com/MyHomeworkSpace/api-server/data"
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

func Data_GetUserByID(id int) (data.User, error) {
	rows, err := DB.Query("SELECT id, name, username, email, type, features, level, showMigrateMessage FROM users WHERE id = ?", id)
	if err != nil {
		return data.User{}, err
	}
	defer rows.Close()
	if rows.Next() {
		user := data.User{}
		err := rows.Scan(&user.ID, &user.Name, &user.Username, &user.Email, &user.Type, &user.Features, &user.Level, &user.ShowMigrateMessage)
		if err != nil {
			return data.User{}, err
		}
		return user, nil
	} else {
		return data.User{}, ErrDataNotFound
	}
}
