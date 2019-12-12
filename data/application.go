package data

// An Application describes a third-party application designed to integrate with MyHomeworkSpace.
type Application struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	AuthorName  string `json:"authorName"`
	ClientID    string `json:"clientId"`
	CallbackURL string `json:"callbackUrl"`
}

// An ApplicationAuthorization describes a user's authorization of an application's access to their account.
type ApplicationAuthorization struct {
	ID            int    `json:"id"`
	ApplicationID int    `json:"applicationId"`
	Name          string `json:"name"`
	AuthorName    string `json:"authorName"`
}
