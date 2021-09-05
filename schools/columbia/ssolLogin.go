package columbia

import (
	"errors"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (s *school) parseSSOLLoginForm(doc *goquery.Document) (string, url.Values, error) {
	ret := url.Values{}

	loginForm := doc.Find("#login #logact form")

	fields := loginForm.Find("input[type=hidden], input[type=submit]")
	for i := 0; i < fields.Length(); i++ {
		field := fields.Eq(i)
		fieldName := field.AttrOr("name", "")
		fieldValue := field.AttrOr("value", "")

		if fieldName != "" {
			ret[fieldName] = []string{fieldValue}
		}
	}

	return loginForm.AttrOr("action", ""), ret, nil
}

func (s *school) parseSSOLLoginResult(doc *goquery.Document) (bool, string) {
	loginForm := doc.Find("#login #logact form")
	if loginForm.Length() == 0 {
		return true, ""
	}

	loginErr := loginForm.Find(".clsFormGridErrorMsg")
	if loginErr.Length() == 0 {
		return true, ""
	}

	return false, strings.TrimSpace(loginErr.Text())
}

func (s *school) findSSOLScheduleURL(doc *goquery.Document) (string, error) {
	navItems := doc.Find("#MainMenu ul li ul li a")
	for i := 0; i < navItems.Length(); i++ {
		navItem := navItems.Eq(i)
		navText := strings.TrimSpace(navItem.Text())
		navHref, exists := navItem.Attr("href")
		if !exists {
			continue
		}

		// normal links aren't cursed enough, so ssl uses links with urls of the form
		// javascript:contentReplace('/cgi-bin/ssol/blablabla_secret_session_stuff')
		// we need to parse that middle part out
		navURL := strings.Replace(navHref, "javascript:contentReplace('", "", 1)
		navURL = navURL[:len(navURL)-2]

		if navText == "Student Schedule" {
			return navURL, nil
		}
	}

	return "", errors.New("findSSOLScheduleURL: could not find Schedule link")
}
