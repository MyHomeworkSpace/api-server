package data

type Homework struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Due      string `json:"due"`
	Desc     string `json:"desc"`
	Complete int    `json:"complete"`
	ClassID  int    `json:"classId"`
	UserID   int    `json:"userId"`
}
