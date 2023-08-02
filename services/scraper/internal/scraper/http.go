package scraper

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"go-find-pepe/internal/utils"

	"github.com/PuerkitoBio/goquery"
)

const httpDir = "data/http"

type HttpScraper struct {
	requests               chan httpRequest
	httpReaders            chan io.Reader
	hrefs                  chan string
	allowedHrefSubstrings  []string
	requiredHrefSubstrings []string
}

type httpRequest struct {
	id   string
	href string
	data []byte
}

func newHttpScraper(httpReaders chan io.Reader, allowedHrefSubstrings []string, requiredHrefSubstrings []string) *HttpScraper {
	return &HttpScraper{
		requests:               make(chan httpRequest),
		httpReaders:            httpReaders,
		hrefs:                  make(chan string),
		allowedHrefSubstrings:  allowedHrefSubstrings,
		requiredHrefSubstrings: requiredHrefSubstrings,
	}
}

// TODO: refactor
// func (s *HttpScraper) readDownloadedIds() *HttpScraper {
// 	fileInfos := readDir(httpDir)

// 	for _, file := range fileInfos {
// 		if file.IsDir() {
// 			continue
// 		}

// 		id := removeExtension(file.Name())
// 	}

// 	return s
// }

func (s *HttpScraper) Start(startHref string) *HttpScraper {
	requestId := hash(startHref)

	go func() {
		response, success, _ := getURL(requestId, startHref)
		if success {
			data, err := ioutil.ReadAll(response.Body)
			utils.Check(err)
			s.requests <- httpRequest{id: requestId, data: data, href: startHref}
		}
	}()

	for {
		select {
		case request := <-s.requests:
			go s.storeHtml(request)
			go s.findHref(request)
			s.httpReaders <- bytes.NewBuffer(request.data)
		case href := <-s.hrefs:
			go s.getHttp(href)
		}
	}
}

func (s *HttpScraper) findHref(request httpRequest) *HttpScraper {
	reader := bytes.NewReader(request.data)
	doc, err := goquery.NewDocumentFromReader(reader)
	utils.Check(err)

	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")

		unallowed := [2]string{"javascript", "#"}
		for _, s := range unallowed {
			if strings.Contains(href, s) {
				return
			}
		}

		if !strings.Contains(href, "/") {
			return
		}

		cleanedHref := cleanUpUrl(href)

		hostname := getHostname(cleanedHref)
		if hostname == "" && cleanedHref[0] != '/' {
			cleanedHref = request.href + cleanedHref
		}

		if exists {
			s.hrefs <- cleanedHref
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
	if s.doesHtmlExist(href) {
		return s
	}

	cleanedHref := cleanUpUrl(href)
	hostname := getHostname(cleanedHref)

	correctAllowedSubstrings := stringShouldContainOneFilter(hostname, s.allowedHrefSubstrings)
	correctRequiredSubstrings := stringShouldContainAllFilters(cleanedHref, s.requiredHrefSubstrings)

	if correctAllowedSubstrings && correctRequiredSubstrings {
		response, success, canRetry := getURL(requestId, cleanedHref)

		if success {
			data, err := ioutil.ReadAll(response.Body)
			utils.Check(err)

			s.requests <- httpRequest{id: requestId, data: data, href: href}
		} else if canRetry {
			fmt.Printf("Retrying url: %v\n", href)
			s.hrefs <- href
		}
	}

	return s
}

func (s *HttpScraper) storeHtml(r httpRequest) *HttpScraper {
	fileName := s.createFileName(r.href)
	writeFile(httpDir, fileName, r.data)
	return s
}

// TODO: refactor
// func (s *HttpScraper) loadHtml(fileId string) *HttpScraper {
// 	reader := createReader(httpDir, addExtension(fileId, "html"))
// 	*s.httpReaders <- reader

// 	return s
// }

func (s *HttpScraper) doesHtmlExist(href string) bool {
	fileName := s.createFileName(href)
	return doesFileExist(httpDir, fileName)
}

func (s *HttpScraper) createFileName(href string) (fileName string) {
	fileName = href
	fileName = strings.Replace(fileName, "https://", "", 1)
	fileName = strings.Replace(fileName, "www.", "", 1)
	if fileName[len(fileName)-1] == '/' {
		fileName = fileName[0 : len(fileName)-1]
	}
	fileName = addExtension(fileName, "html")
	return
}
