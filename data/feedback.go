package data

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
