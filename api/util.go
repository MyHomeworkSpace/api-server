package api

import (
	"crypto/rand"
	"encoding/base64"
)

func Util_GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}

func Util_GenerateRandomString(s int) (string, error) {
	b, err := Util_GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

func Util_StringSliceContains(slice []string, text string) bool {
	for _, v := range slice {
		if v == text {
			return true
		}
	}
	return false
}
