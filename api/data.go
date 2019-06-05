package api

import (
	"github.com/MyHomeworkSpace/api-server/data"
)

func Data_GetPrefForUser(key string, userId int) (data.Pref, error) {
	rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", userId, key)
	if err != nil {
		return data.Pref{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return data.Pref{}, data.ErrNotFound
	}

	pref := data.Pref{}
	err = rows.Scan(&pref.ID, &pref.Key, &pref.Value)
	if err != nil {
		return data.Pref{}, err
	}

	return pref, nil
}

func Data_GetTabsByUserID(userId int) ([]data.Tab, error) {
	rows, err := DB.Query("SELECT tabs.id, tabs.slug, tabs.icon, tabs.label, tabs.target FROM tabs INNER JOIN tab_permissions ON tab_permissions.tabId = tabs.id WHERE tab_permissions.userId = ?", userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tabs := []data.Tab{}
	for rows.Next() {
		tab := data.Tab{}
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
		return data.User{}, data.ErrNotFound
	}
}
