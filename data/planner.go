package data

type PlannerAnnouncement struct {
	ID    int    `json:"id"`
	Date  string `json:"date"`
	Text  string `json:"text"`
	Grade int    `json:"grade"`
	Type  int    `json:"type"`
}

type PlannerFriday struct {
	ID    int    `json:"id"`
	Date  string `json:"date"`
	Index int    `json:"index"`
}
