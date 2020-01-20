package data

// A HomeworkClass is a class that can be associated with Homework items.
type HomeworkClass struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Teacher   string `json:"teacher"`
	Color     string `json:"color"`
	SortIndex int    `json:"sortIndex"`
	UserID    int    `json:"userId"`
}

// GetClassesForUser gets all HomeworkClasses for the given user.
func GetClassesForUser(user *User) ([]HomeworkClass, error) {
	rows, err := DB.Query("SELECT id, name, teacher, color, sortIndex, userId FROM classes WHERE userId = ? ORDER BY sortIndex ASC", user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	classes := []HomeworkClass{}
	for rows.Next() {
		resp := HomeworkClass{-1, "", "", "", -1, -1}
		rows.Scan(&resp.ID, &resp.Name, &resp.Teacher, &resp.Color, &resp.SortIndex, &resp.UserID)
		classes = append(classes, resp)
	}

	return classes, nil
}
