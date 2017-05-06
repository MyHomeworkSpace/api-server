package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func DaltonLogin(username string, password string) (map[string]interface{}, string, error) {
	hc := http.Client{}

	form := url.Values{}
    form.Add("username", username)
    form.Add("password", password)

	req, err := http.NewRequest("POST", "https://hsregistration.dalton.org/src/server/index.phplogin", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, "internal_server_error", err
	}
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Add("Referer", "https://hsregistration.dalton.org/")
    req.Header.Add("Origin", "https://hsregistration.dalton.org")

    resp, err := hc.Do(req)
	if err != nil {
		return nil, "internal_server_error", err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return nil, "dalton_login_malformed_response", err
    }
	if data["logged_in"] == false {
		return nil, "dalton_creds_incorrect", nil
	}

	return data, "", nil
}
