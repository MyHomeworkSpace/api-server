package auth

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/MyHomeworkSpace/api-server/blackbaud"
)

func DaltonLogin(username string, password string) (map[string]interface{}, string, string, http.CookieJar, error) {
	username = username + "@dalton.org"

	schoolSlug := "dalton"

	// set up ajax token and stuff
	ajaxToken, err := blackbaud.GetAjaxToken(schoolSlug)
	if err != nil {
		return nil, "internal_server_error", "", nil, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, "internal_server_error", "", nil, err
	}

	// sign in to blackbaud
	response, err := blackbaud.Request(schoolSlug, "POST", "Bbid/StatusByName", url.Values{}, map[string]interface{}{
		"userName":     username,
		"rememberType": "2",
	}, jar, ajaxToken, "")

	if !response.(map[string]interface{})["Linkable"].(bool) {
		return nil, "creds_incorrect", "", nil, nil
	}

	auth0State, err := blackbaud.GetAuth0State(jar)
	if err != nil {
		return nil, "internal_server_error", "", nil, err
	}

	adLoginData, err := blackbaud.GetAuth0LoginPage(username, auth0State, jar)

	adKmsiData, err := blackbaud.MakeADLogin(adLoginData, username, password, jar)
	if err != nil && err.Error() == "creds_incorrect" {
		return nil, err.Error(), "", nil, err
	} else if err != nil {
		return nil, "internal_server_error", "", nil, err
	}

	err = blackbaud.BypassKMSIPage(adKmsiData, jar)
	if err != nil {
		return nil, "internal_server_error", "", nil, err
	}

	tokens, err := blackbaud.GetBlackbaudToken(jar)
	if err != nil {
		return nil, "internal_server_error", "", nil, err
	}

	accesstoken := tokens["access_token"].(string)

	_, err = blackbaud.Request("dalton", "GET", "bbid/login", url.Values{}, nil, jar, ajaxToken, accesstoken)
	if err != nil {
		return nil, "internal_server_error", "", nil, err
	}

	// get user id
	response, err = blackbaud.Request(schoolSlug, "GET", "webapp/context", url.Values{}, map[string]interface{}{}, jar, ajaxToken, accesstoken)
	if err != nil {
		return nil, "internal_server_error", "", nil, err
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

	return data, "", ajaxToken, jar, nil
}
