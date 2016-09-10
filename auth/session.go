package auth

import (
    "crypto/rand"
    "encoding/base64"
	"log"
)

type SessionInfo struct {
	UserId int
    Username string
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    // Note that err == nil only if we read len(b) bytes.
    if err != nil {
        return nil, err
    }

    return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
func GenerateRandomString(s int) (string, error) {
    b, err := GenerateRandomBytes(s)
    return base64.URLEncoding.EncodeToString(b), err
}

// GenerateUID creates a random session ID, for use with the session store.
func GenerateUID() (string, error) {
	return GenerateRandomString(26)
}

// SetSession stores the given value under the given name in the database.
// If the given name is already used, its value is overwritten.
func SetSession(name string, value SessionInfo) {
	stmt, err := DB.Prepare("INSERT INTO sessions(id, userId, username) VALUES(?, ?, ?)")
	if err != nil {
		log.Println("Error while setting session: ")
		log.Println(err)
		return
	}
	_, err = stmt.Exec(name, value.UserId, value.Username)
	if err != nil {
		// must be because it already exists
		// so, try to UPDATE instead
		stmt, err := DB.Prepare("UPDATE sessions SET userId=?, username=? WHERE id=?")
		if err != nil {
			log.Println("Error while setting session: ")
			log.Println(err)
			return
		}
		_, err = stmt.Exec(value.UserId, value.Username, name)
		if err != nil {
			log.Println("Error while setting session: ")
			log.Println(err)
			return
		}
	}
}

// GetSession retrieves the session information for the given name
func GetSession(name string) (SessionInfo) {
	rows, err := DB.Query("SELECT userId, username from sessions where id = ?", name)
	if err != nil {
		log.Println("Error while getting session: ")
		log.Println(err)
		return SessionInfo{-1, ""}
	}
	defer rows.Close()
	rows.Next()
	retval := SessionInfo{-1, ""}
	err = rows.Scan(&retval.UserId, &retval.Username)

	return retval
}
