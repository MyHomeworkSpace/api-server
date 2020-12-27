package blackbaud

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var AjaxRegex = regexp.MustCompile("<input name=\"__RequestVerificationToken\" type=\"hidden\" value=\"(.*)\" \\/>")

func GetBaseDomain(schoolSlug string) string {
	return "https://" + schoolSlug + ".myschoolapp.com"
}

func GetAjaxPageURL(schoolSlug string) string {
	return GetBaseDomain(schoolSlug) + "/app/"
}

func GetAPIBaseURL(schoolSlug string) string {
	return GetBaseDomain(schoolSlug) + "/api/"
}

func GetAjaxToken(schoolSlug string) (string, error) {
	resp, err := http.Get(GetAjaxPageURL(schoolSlug))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	strResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(AjaxRegex.FindAllSubmatch(strResponse, -1)[0][1]), nil
}

func GetAuth0State(jar http.CookieJar) (string, error) {
	client := http.Client{Jar: jar}

	baseurl := "https://s21aidntoken00blkbapp01.nxt.blackbaud.com/auth0/state"

	values := url.Values{}
	values.Set("redirectUrl", "https://dalton.myschoolapp.com/app?bb_id=1#login")
	values.Set("sky_auth", "true")

	requrl := baseurl + "?" + values.Encode()

	req, err := http.NewRequest("POST", requrl, nil)

	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:70.0) Gecko/20100101 Firefox/70.0")
	req.Header.Set("Referer", "https://signin.blackbaud.com/")
	req.Header.Set("Origin", "https://signin.blackbaud.com")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	strResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var decodedResponse interface{}
	err = json.Unmarshal(strResponse, &decodedResponse)
	if err != nil {
		return "", err
	}

	return decodedResponse.(map[string]interface{})["state"].(string), nil

}

func GetAuth0LoginPage(email string, state string, jar http.CookieJar) (map[string]interface{}, error) {
	client := http.Client{Jar: jar}
	values := url.Values{}

	values.Set("client_id", "F6MAbpG5xPoi749rDIjqxsXFdbeg0mjU")
	values.Set("redirect_uri", "https://s21aidntoken00blkbapp01.nxt.blackbaud.com/auth0/callback?connection=dalton-org")
	values.Set("connection", "dalton-org")
	values.Set("response_type", "code")
	values.Set("state", state)
	values.Set("login_hint", email)
	values.Set("scope", "openid profile email")
	values.Set("auth0Client", "eyJuYW1lIjoiYXV0aDAuanMiLCJ2ZXJzaW9uIjoiOS4xMC4yIn0=")

	baseurl := "https://blackbaudinc.auth0.com/authorize"
	requrl := baseurl + "?" + values.Encode()

	req, err := http.NewRequest("GET", requrl, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.14; rv:70.0) Gecko/20100101 Firefox/70.0")
	req.Header.Set("Referer", "https://signin.blackbaud.com/")
	req.Header.Set("Origin", "https://signin.blackbaud.com")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	strResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	regexpression, _ := regexp.Compile("\\$Config=(.*?);\\n//]]></script>")
	jsonData := regexpression.FindAllStringSubmatch(string(strResponse), 1)[0][1]

	var decodedData interface{}
	err = json.Unmarshal([]byte(jsonData), &decodedData)

	decodedDataMap := decodedData.(map[string]interface{})
	return decodedDataMap, err
}

func MakeADLogin(loginDetails map[string]interface{}, email string, password string, jar http.CookieJar) (map[string]interface{}, error) {
	ctx := loginDetails["sCtx"].(string)
	canary := loginDetails["canary"].(string)
	hpgrequestid := loginDetails["sessionId"].(string)
	flowtoken := loginDetails["oGetCredTypeResult"].(map[string]interface{})["FlowToken"].(string)

	params := url.Values{}
	params.Add("i13", "0")
	params.Add("login", email)
	params.Add("loginfmt", email)
	params.Add("type", "11")
	params.Add("LoginOptions", "3")
	params.Add("lrt", "")
	params.Add("lrtPartition", "")
	params.Add("hisRegion", "")
	params.Add("hisScaleUnit", "")
	params.Add("passwd", password)
	params.Add("ps", "2")
	params.Add("psRNGCDefaultType", "")
	params.Add("psRNGCEntropy", "")
	params.Add("psRNGCSLK", "")
	params.Add("canary", canary)
	params.Add("ctx", ctx)
	params.Add("hpgrequestid", hpgrequestid)
	params.Add("flowToken", flowtoken)
	params.Add("PPSX", "")
	params.Add("NewUser", "1")
	params.Add("FoundMSAs", "")
	params.Add("fspost", "0")
	params.Add("i21", "0")
	params.Add("CookieDisclosure", "0")
	params.Add("IsFidoSupported", "0")
	params.Add("i2", "1")
	params.Add("i17", "")
	params.Add("i18", "")
	params.Add("i19", strconv.Itoa(1000+rand.Intn(4000)))

	client := http.Client{Jar: jar}

	resp, err := client.PostForm("https://login.microsoftonline.com/dalton.org/login", params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	strResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(string(strResponse), "<meta name=\"PageID\" content=\"KmsiInterrupt\" />") {
		// the login wasn't successful
		return nil, errors.New("creds_incorrect")
	}

	regexpression, _ := regexp.Compile("\\$Config=(.*?);\\n//]]></script>")
	jsonData := regexpression.FindAllStringSubmatch(string(strResponse), 1)[0][1]

	var decodedData interface{}
	err = json.Unmarshal([]byte(jsonData), &decodedData)

	decodedDataMap := decodedData.(map[string]interface{})
	return decodedDataMap, err
}

func BypassKMSIPage(loginDetails map[string]interface{}, jar http.CookieJar) error {
	ctx := loginDetails["sCtx"].(string)
	canary := loginDetails["canary"].(string)
	hpgrequestid := loginDetails["sessionId"].(string)
	flowtoken := loginDetails["sFT"].(string)

	params := url.Values{}

	params.Add("LoginOptions", "3")
	params.Add("type", "28")
	params.Add("ctx", ctx)
	params.Add("hpgrequestid", hpgrequestid)
	params.Add("flowToken", flowtoken)
	params.Add("canary", canary)
	params.Add("i2", "")
	params.Add("i17", "")
	params.Add("i18", "")
	params.Add("i19", strconv.Itoa(1000+rand.Intn(4000)))

	client := http.Client{Jar: jar}

	_, err := client.PostForm("https://login.microsoftonline.com/kmsi", params)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return nil
}

func GetBlackbaudToken(jar http.CookieJar) (map[string]interface{}, error) {
	client := http.Client{Jar: jar}
	req, err := http.NewRequest("POST", "https://s21aidntoken00blkbapp01.nxt.blackbaud.com/oauth2/token", nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/57.0.2987.133 Safari/537.36")
	req.Header.Set("Referer", "https://s21aidntoken00blkbapp01.nxt.blackbaud.com/oauth2/token")
	req.Header.Set("X-CSRF", "token_needed")

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
	return decodedResponse.(map[string]interface{}), err
}

func Request(schoolSlug string, requestType string, path string, urlParams url.Values, postData map[string]interface{}, jar http.CookieJar, ajaxToken string, bearerToken string) (interface{}, error) {
	client := http.Client{Jar: jar}

	url := GetAPIBaseURL(schoolSlug) + path
	url = url + "?" + urlParams.Encode()

	var requestBody io.Reader
	if requestType == "POST" {
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
	req.Header.Set("Referer", GetAjaxPageURL(schoolSlug))
	req.Header["RequestVerificationToken"] = []string{ajaxToken}
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

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
