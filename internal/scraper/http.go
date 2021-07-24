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

type HttpScraper struct {
	requests               chan request
	httpFileIds            chan string
	httpReaders            *chan *io.Reader
	hrefs                  chan string
	allowedHrefSubstrings  []string
	requiredHrefSubstrings []string
}

func newHttpScraper(allowedHrefSubstrings []string, requiredHrefSubstrings []string, httpReaders *chan *io.Reader) *HttpScraper {
	return &HttpScraper{
		requests:               make(chan request),
		httpFileIds:            make(chan string),
		httpReaders:            httpReaders,
		hrefs:                  make(chan string),
		allowedHrefSubstrings:  allowedHrefSubstrings,
		requiredHrefSubstrings: requiredHrefSubstrings,
	}
}

func (s *HttpScraper) readDownloadedIds() *HttpScraper {
	fileInfos := readDir(httpDir)

	for _, file := range fileInfos {
		if file.IsDir() {
			continue
		}

		id := removeExtension(file.Name())
		s.httpFileIds <- id
	}

	return s
}

func (s *HttpScraper) Start(startHref string) *HttpScraper {
	requestId := hash(startHref)

	go s.getURL(requestId, startHref)

	for {
		select {
		case httpFileId := <-s.httpFileIds:
			go s.loadHtml(httpFileId)
		case request := <-s.requests:
			go s.storeHtml(request)
		case href := <-s.hrefs:
			go s.getHttp(href)
		}
	}
}

func (s *HttpScraper) findHref(reader *io.Reader) *HttpScraper {
	doc, err := goquery.NewDocumentFromReader(*reader)
	utils.Check(err)

	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		fmt.Println(href)

		if exists {
			s.hrefs <- href
		}
	})

	return s
}

// FIXME: nasty return if it does already exist
func (s *HttpScraper) getHttp(href string) *HttpScraper {
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
	s.httpFileIds <- r.id

	return s
}

func (s *HttpScraper) loadHtml(fileId string) *HttpScraper {
	reader := createReader(httpDir, fileId, "html")

	fmt.Println("adding reader")
	*s.httpReaders <- reader
	fmt.Println("adding reader 2")

	return s
}

func (s *HttpScraper) doesHtmlExist(fileId string) bool {
	return doesFileExist(httpDir, fileId, "html")
}
