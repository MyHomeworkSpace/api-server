package api

import (
	"errors"
)

var (
	ErrDataNotFound = errors.New("data: not found")
)

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

func Data_GetUserByID(id int) (User, error) {
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
		return User{}, ErrDataNotFound
	}
}

func Data_GetUserGrade(user User) (int, error) {
	return 0, nil
}
