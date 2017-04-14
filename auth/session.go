package auth

import (
    "crypto/rand"
    "encoding/base64"
	"log"
	"strconv"
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
	result := RedisClient.HMSet("session:" + name, map[string]string{
		"userId": strconv.Itoa(value.UserId),
		"username": value.Username,
	})
	if result.Err() != nil {
		log.Println("Error while setting session: ")
		log.Println(result.Err())
		return
	}
}

// GetSession retrieves the session information for the given name
func GetSession(name string) (SessionInfo) {
	result := RedisClient.HGetAll("session:" + name)
	if result.Err() != nil {
		log.Println("Error while getting session: ")
		log.Println(result.Err())
		return SessionInfo{-1, ""}
	}

	resultMap, err := result.Result()
	if err != nil {
		log.Println("Error while getting session: ")
		log.Println(err)
		return SessionInfo{-1, ""}
	}

	retval := SessionInfo{-1, ""}

	retval.UserId, err = strconv.Atoi(resultMap["userId"])
	retval.Username = resultMap["username"]

	return retval
}
