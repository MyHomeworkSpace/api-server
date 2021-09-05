package columbia

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const ssolBaseURL = "https://ssol.columbia.edu/"

func (s *school) ssolRequest(method string, url string, formData url.Values, jar http.CookieJar) (*goquery.Document, error) {
	client := http.Client{Jar: jar}

	var requestBody io.Reader
	if method == http.MethodPost {
		requestBody = strings.NewReader(formData.Encode())
	}

	fullURL := ssolBaseURL + url
	req, err := http.NewRequest(method, fullURL, requestBody)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}
