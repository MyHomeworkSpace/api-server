package data

type User struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Username           string `json:"-"`
	Email              string `json:"email"`
	Type               string `json:"type"`
	Features           string `json:"features"`
	Level              int    `json:"level"`
	ShowMigrateMessage int    `json:"showMigrateMessage"`
}

type Tab struct {
	ID     int    `json:"id"`
	Slug   string `json:"slug"`
	Icon   string `json:"icon"`
	Label  string `json:"label"`
	Target string `json:"target"`
}

// GetTabsByUserID fetches all tabs that the given user has access to.
func GetTabsByUserID(userID int) ([]Tab, error) {
	rows, err := DB.Query("SELECT tabs.id, tabs.slug, tabs.icon, tabs.label, tabs.target FROM tabs INNER JOIN tab_permissions ON tab_permissions.tabId = tabs.id WHERE tab_permissions.userId = ?", userID)
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

// GetUserByID fetches data for the given user ID.
func GetUserByID(id int) (User, error) {
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
		return User{}, ErrNotFound
	}
}
