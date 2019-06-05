package data

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

type Tab struct {
	ID     int    `json:"id"`
	Slug   string `json:"slug"`
	Icon   string `json:"icon"`
	Label  string `json:"label"`
	Target string `json:"target"`
}
