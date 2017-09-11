package api

import (
	"errors"
	"strconv"
)

var (
	ErrDataBadUsername = errors.New("data: bad username")
	ErrDataNotFound    = errors.New("data: not found")
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
	if len(user.Username) < 4 {
		return -1, ErrDataBadUsername
	}
	yearInfoString := user.Username[1:3]
	yearInfo, err := strconv.Atoi(yearInfoString)
	if err != nil {
		return -1, ErrDataBadUsername
	}

	differenceFromBase := (yearInfo - 19) * -1
	grade := Grade_ClassOf2019 + differenceFromBase

	return grade, nil
}
