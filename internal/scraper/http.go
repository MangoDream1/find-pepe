package scraper

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"go-find-pepe/internal/utils"

	"github.com/PuerkitoBio/goquery"
)

const httpDir = "data/http"

type request struct {
	id       string
	response *http.Response
}

type HttpScraper struct {
	requests               chan request
	httpFileIds            *chan string
	readers                *chan *io.Reader
	hrefs                  chan string
	allowedHrefSubstrings  []string
	requiredHrefSubstrings []string
}

func newHttpScraper(allowedHrefSubstrings []string, requiredHrefSubstrings []string, readers *chan *io.Reader, httpFileIds *chan string) *HttpScraper {
	return &HttpScraper{
		requests:               make(chan request),
		httpFileIds:            httpFileIds,
		readers:                readers,
		hrefs:                  make(chan string),
		allowedHrefSubstrings:  allowedHrefSubstrings,
		requiredHrefSubstrings: requiredHrefSubstrings,
	}
}

// FIXME: nasty return if it does already exist
func (s *HttpScraper) GetHttp(href string) *HttpScraper {
	requestId := hash(href)
	if s.doesHtmlExist(requestId) {
		fmt.Printf("Ignoring %v; already exists \n", href)
		return s
	}

	cleanedHref := cleanUpUrl(href)
	hostname := getHostname(cleanedHref)

	correctAllowedSubstrings := stringShouldContainOneFilter(hostname, s.allowedHrefSubstrings)
	correctRequiredSubstrings := stringShouldContainAllFilters(cleanedHref, s.requiredHrefSubstrings)

	if correctAllowedSubstrings && correctRequiredSubstrings {
		go s.getURL(requestId, cleanedHref)
	}

	return s
}

func (s *HttpScraper) getURL(requestId string, url string) *HttpScraper {
	response, err := http.Get(url)

	if err != nil {
		fmt.Printf("Failed to GET %v; ignoring \n", url)
		fmt.Println(err.Error())
		return s
	}

	if response.StatusCode != 200 {
		fmt.Printf("Non-OK response: %v %v; ignoring \n", url, response.StatusCode)
		return s
	}

	fmt.Printf("Successfully fetched %v \n", url)
	s.requests <- request{id: requestId, response: response}
	return s
}

func (s *HttpScraper) storeHtml(r request) *HttpScraper {
	doc, err := ioutil.ReadAll(r.response.Body)
	utils.Check(err)

	writeFile(httpDir, r.id, "html", doc)
	*s.httpFileIds <- r.id

	return s
}

func (s *HttpScraper) loadHtml(fileId string) *HttpScraper {
	reader := createReader(httpDir, fileId, "html")
	*s.readers <- reader

	return s
}

func (s *HttpScraper) findHref(reader *io.Reader) *HttpScraper {
	doc, err := goquery.NewDocumentFromReader(*reader)
	utils.Check(err)

	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")

		if exists {
			s.hrefs <- href
		}
	})

	return s
}

func (s *HttpScraper) doesHtmlExist(fileId string) bool {
	return doesFileExist(httpDir, fileId, "html")
}
