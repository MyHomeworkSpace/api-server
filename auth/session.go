package auth

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"strconv"
	"time"
)

type SessionInfo struct {
	UserID int
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
	result := RedisClient.HMSet("session:"+name, map[string]string{
		"userId": strconv.Itoa(value.UserID),
	})

	if result.Err() != nil {
		log.Println("Error while setting session: ")
		log.Println(result.Err())
		return
	}

	expireResult := RedisClient.Expire("session:"+name, 7*24*time.Hour)
	if expireResult.Err() != nil {
		log.Println("Error while setting session: ")
		log.Println(expireResult.Err())
		return
	}
}

// GetSession retrieves the session information for the given name
func GetSession(name string) SessionInfo {
	result := RedisClient.HGetAll("session:" + name)
	if result.Err() != nil {
		log.Println("Error while getting session: ")
		log.Println(result.Err())
		return SessionInfo{-1}
	}

	resultMap, err := result.Result()
	if err != nil {
		log.Println("Error while getting session: ")
		log.Println(err)
		return SessionInfo{-1}
	}

	retval := SessionInfo{-1}

	retval.UserID, err = strconv.Atoi(resultMap["userId"])

	if err != nil {
		return SessionInfo{-1}
	}

	return retval
}

func GetSessionFromAuthToken(authToken string) SessionInfo {
	rows, err := DB.Query("SELECT users.id FROM application_authorizations INNER JOIN users ON application_authorizations.userId = users.id WHERE application_authorizations.token = ?", authToken)
	if err != nil {
		log.Println("Error while getting session from auth token:")
		log.Println(err)
		return SessionInfo{-1}
	}
	defer rows.Close()
	rows.Next()

	retval := SessionInfo{-1}
	rows.Scan(&retval.UserID)

	return retval
}
