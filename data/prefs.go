package data

type Pref struct {
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// GetPrefForUser fetches the Pref with the given key for the given user.
func GetPrefForUser(key string, userID int) (Pref, error) {
	rows, err := DB.Query("SELECT `id`, `key`, `value` FROM prefs WHERE userId = ? AND `key` = ?", userID, key)
	if err != nil {
		return Pref{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return Pref{}, ErrNotFound
	}

	pref := Pref{}
	err = rows.Scan(&pref.ID, &pref.Key, &pref.Value)
	if err != nil {
		return Pref{}, err
	}

	return pref, nil
}
