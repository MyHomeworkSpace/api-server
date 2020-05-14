package data

import (
	"encoding/json"
	"strconv"
	"time"

	"gopkg.in/redis.v5"
)

// An EmailTokenType describes different types of email token
type EmailTokenType int

// Define the default EmailTokenType.
const (
	EmailTokenNone EmailTokenType = iota
	EmailTokenResetPassword
	EmailTokenChangeEmail
	EmailTokenVerifyEmail
)

// An EmailToken is used for situations like an email change or a password reset, where a confirmation email must be sent.
type EmailToken struct {
	Token    string         `json:"token"`
	Type     EmailTokenType `json:"type"`
	Metadata string         `json:"metadata"`
	UserID   int            `json:"userID"`
}

type User struct {
	ID                 int          `json:"id"`
	Name               string       `json:"name"`
	Email              string       `json:"email"`
	PasswordHash       string       `json:"-"`
	Type               string       `json:"type"`
	Features           string       `json:"features"`
	Level              int          `json:"level"`
	EmailVerified      bool         `json:"emailVerified"`
	ShowMigrateMessage int          `json:"showMigrateMessage"`
	CreatedAt          int          `json:"createdAt"`
	LastLoginAt        int          `json:"lastLoginAt"`
	Schools            []SchoolInfo `json:"schools"`
}

type Tab struct {
	ID     int    `json:"id"`
	Slug   string `json:"slug"`
	Icon   string `json:"icon"`
	Label  string `json:"label"`
	Target string `json:"target"`
}

// GetEmailToken fetches the given email token.
func GetEmailToken(token string) (EmailToken, error) {
	storedToken, err := RedisClient.HGetAll("token:" + token).Result()
	if err == redis.Nil {
		return EmailToken{}, ErrNotFound
	} else if err != nil {
		return EmailToken{}, err
	}

	if storedToken["type"] == "" {
		return EmailToken{}, ErrNotFound
	}

	tokenType, err := strconv.Atoi(storedToken["type"])
	if err != nil {
		return EmailToken{}, err
	}

	userID, err := strconv.Atoi(storedToken["userID"])
	if err != nil {
		return EmailToken{}, err
	}

	response := EmailToken{
		Token:    token,
		Type:     EmailTokenType(tokenType),
		Metadata: storedToken["metadata"],
		UserID:   userID,
	}

	return response, nil
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

// UserExistsWithEmail checks for a user with the given email address. If one exists, it returns the ID of that user.
func UserExistsWithEmail(email string) (bool, int, error) {
	rows, err := DB.Query("SELECT id FROM users WHERE email = ?", email)
	if err != nil {
		return false, -1, err
	}

	defer rows.Close()

	if rows.Next() {
		id := -1
		err = rows.Scan(&id)
		if err != nil {
			return false, -1, err
		}
		return true, id, nil
	}

	return false, -1, nil
}

// GetUserByID fetches data for the given user ID.
func GetUserByID(id int) (User, error) {
	rows, err := DB.Query("SELECT id, name, email, password, type, features, emailVerified, level, showMigrateMessage, createdAt, lastLoginAt FROM users WHERE id = ?", id)
	if err != nil {
		return User{}, err
	}

	if rows.Next() {
		user := User{}
		emailVerified := 0

		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Type, &user.Features, &emailVerified, &user.Level, &user.ShowMigrateMessage, &user.CreatedAt, &user.LastLoginAt)
		if err != nil {
			return User{}, err
		}

		if emailVerified == 1 {
			user.EmailVerified = true
		}

		rows.Close()

		user.Schools = []SchoolInfo{}

		schoolRows, err := DB.Query("SELECT id, schoolId, enabled, data, userId FROM schools WHERE userId = ?", user.ID)
		if err != nil {
			return User{}, err
		}

		for schoolRows.Next() {
			info := SchoolInfo{}
			dataString := ""

			schoolRows.Scan(&info.EnrollmentID, &info.SchoolID, &info.Enabled, &dataString, &info.UserID)

			// parse JSON data
			data := map[string]interface{}{}

			err := json.Unmarshal([]byte(dataString), &data)
			if err != nil {
				return User{}, err
			}

			// get school and hydrate it
			school, err := MainRegistry.GetSchoolByID(info.SchoolID)
			if err != nil {
				return User{}, err
			}

			school.Hydrate(data)

			// set SchoolInfo
			info.DisplayName = school.Name()
			info.ShortName = school.ShortName()
			info.UserDetails = school.UserDetails()
			info.EmailAddress = school.EmailAddress()
			info.School = school

			user.Schools = append(user.Schools, info)
		}

		schoolRows.Close()

		return user, nil
	} else {
		rows.Close()
		return User{}, ErrNotFound
	}
}

// DeleteEmailToken deletes the given email token.
func DeleteEmailToken(token EmailToken) error {
	return RedisClient.Del("token:" + token.Token).Err()
}

// SaveEmailToken saves the given email token with a default expiry.
func SaveEmailToken(token EmailToken) error {
	err := RedisClient.HMSet("token:"+token.Token, map[string]string{
		"type":     strconv.Itoa(int(token.Type)),
		"metadata": token.Metadata,
		"userID":   strconv.Itoa(token.UserID),
	}).Err()

	if err != nil {
		return err
	}

	return RedisClient.Expire("token:"+token.Token, 1*time.Hour).Err()
}
