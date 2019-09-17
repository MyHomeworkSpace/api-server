package data

// A Prefix defines a group of words that get automatically recognized (for example: HW, Test, Quiz)
type Prefix struct {
	ID         int      `json:"id"`
	Background string   `json:"background"`
	Color      string   `json:"color"`
	Words      []string `json:"words"`
	TimedEvent bool     `json:"timedEvent"`
	Default    bool     `json:"default"`
}
