package api

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var Blackbaud_AjaxRegex = regexp.MustCompile("<input name=\"__RequestVerificationToken\" type=\"hidden\" value=\"(.*)\" \\/>")
const Blackbaud_AjaxPage = "https://dalton.myschoolapp.com/app/"
const Blackbaud_Domain = "https://dalton.myschoolapp.com/api/"

func Blackbaud_GetAjaxToken() (string, error) {
	resp, err := http.Get(Blackbaud_AjaxPage)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	strResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(Blackbaud_AjaxRegex.FindAllSubmatch(strResponse, -1)[0][1]), nil
}

func Blackbaud_Request(requestType string, path string, urlParams url.Values, postData map[string]interface{}, jar http.CookieJar, ajaxToken string) (interface{}, error) {
	client := http.Client{Jar: jar}

	url := Blackbaud_Domain + path
	url = url + "?" + urlParams.Encode()

	var requestBody io.Reader
	if (requestType == "POST") {
		data, err := json.Marshal(postData)
		if err != nil {
			return nil, err
		}
		requestBody = strings.NewReader(string(data))
	}

	req, err := http.NewRequest(requestType, url, requestBody)
	if err != nil {
		return nil, err
	}

	// blackbaud has an "enhanced security system to ensure a safe browsing experience"
	// i'm not 100% sure how it improves the security at all
	// but these headers are needed for stuff to work
	// it's like a more advanced version of referer checking done by courses/hsregistration
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36")
	req.Header.Set("Referer", "https://dalton.myschoolapp.com/app/")
	req.Header["RequestVerificationToken"] = []string{ ajaxToken }
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	strResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var decodedResponse interface{}
	err = json.Unmarshal(strResponse, &decodedResponse)
	if err != nil {
		return nil, err
	}

	return decodedResponse, nil
}