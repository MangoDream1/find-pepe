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
	requests               chan httpRequest
	httpFileIds            chan string
	httpReaders            *chan *io.Reader
	hrefs                  chan string
	allowedHrefSubstrings  []string
	requiredHrefSubstrings []string
}

type httpRequest struct {
	id       string
	response *http.Response
}

func newHttpScraper(httpReaders *chan *io.Reader, allowedHrefSubstrings []string, requiredHrefSubstrings []string) *HttpScraper {
	return &HttpScraper{
		requests:               make(chan httpRequest),
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

	go func() {
		response, success, _ := getURL(requestId, startHref)
		if success {
			s.requests <- httpRequest{id: requestId, response: response}
		} else {
			panic(fmt.Sprintf("Failed to retrive startHref %v", startHref))
		}
	}()

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

		if exists {
			s.hrefs <- href
		}
	})

	return s
}

// FIXME: nasty return if it does already exist
func (s *HttpScraper) getHttp(href string) *HttpScraper {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("An error has occurred while trying to retrieve href: %v with error %v; ignoring \n", href, err)
		}
	}()

	requestId := hash(href)
	if s.doesHtmlExist(requestId) {
		return s
	}

	cleanedHref := cleanUpUrl(href)
	hostname := getHostname(cleanedHref)

	correctAllowedSubstrings := stringShouldContainOneFilter(hostname, s.allowedHrefSubstrings)
	correctRequiredSubstrings := stringShouldContainAllFilters(cleanedHref, s.requiredHrefSubstrings)

	if correctAllowedSubstrings && correctRequiredSubstrings {
		response, success, canRetry := getURL(requestId, cleanedHref)

		if success {
			s.requests <- httpRequest{id: requestId, response: response}
		} else if canRetry {
			fmt.Printf("Retrying url: %v", href)
			s.hrefs <- href
		}
	}

	return s
}

func (s *HttpScraper) storeHtml(r httpRequest) *HttpScraper {
	doc, err := ioutil.ReadAll(r.response.Body)
	utils.Check(err)

	writeFile(httpDir, addExtension(r.id, "html"), doc)
	s.httpFileIds <- r.id

	return s
}

func (s *HttpScraper) loadHtml(fileId string) *HttpScraper {
	reader := createReader(httpDir, addExtension(fileId, "html"))
	*s.httpReaders <- reader

	return s
}

func (s *HttpScraper) doesHtmlExist(fileId string) bool {
	return doesFileExist(httpDir, addExtension(fileId, "html"))
}
