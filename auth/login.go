package auth

import (
	"net/http/cookiejar"
	"net/url"

	"github.com/MyHomeworkSpace/api-server/blackbaud"
)

func DaltonLogin(username string, password string) (map[string]interface{}, string, error) {
	schoolSlug := "dalton"

	// set up ajax token and stuff
	ajaxToken, err := blackbaud.GetAjaxToken(schoolSlug)
	if err != nil {
		return nil, "internal_server_error", err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, "internal_server_error", err
	}

	// sign in to blackbaud
	response, err := blackbaud.Request(schoolSlug, "POST", "SignIn", url.Values{}, map[string]interface{}{
		"From":            "",
		"InterfaceSource": "WebApp",
		"Password":        password,
		"Username":        username,
		"remember":        "false",
	}, jar, ajaxToken)

	if err != nil {
		return nil, "creds_incorrect", nil
	}

	result, worked := (response.(map[string]interface{}))["AuthenticationResult"].(float64)

	if worked && result == 5 {
		return nil, "bb_signin_rate_limit", nil
	}

	if !worked || result == 1 || result == 2 {
		return nil, "creds_incorrect", nil
	}

	// get user id
	response, err = blackbaud.Request(schoolSlug, "GET", "webapp/context", url.Values{}, map[string]interface{}{}, jar, ajaxToken)
	if err != nil {
		return nil, "internal_server_error", err
	}

	// bbUserId := int(((response.(map[string]interface{}))["UserInfo"].(map[string]interface{}))["UserId"].(float64))
	firstName := ((response.(map[string]interface{}))["UserInfo"].(map[string]interface{}))["FirstName"].(string)
	lastName := ((response.(map[string]interface{}))["UserInfo"].(map[string]interface{}))["LastName"].(string)
	fullName := firstName + " " + lastName

	data := map[string]interface{}{}

	data["fullname"] = fullName
	roles := []string{}

	personas := response.(map[string]interface{})["Personas"].([]interface{})
	for _, persona := range personas {
		roles = append(roles, persona.(map[string]interface{})["UrlFriendlyDescription"].(string))
	}

	data["roles"] = roles

	return data, "", nil
}
