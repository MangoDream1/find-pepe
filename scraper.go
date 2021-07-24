package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

const httpDir = "data/http"

type request struct {
	id       string
	response *http.Response
}

type Scraper struct {
	requests               chan request
	HttpFileIds            chan string
	readers                chan *io.Reader
	hrefs                  chan string
	allowedHrefSubstrings  []string
	requiredHrefSubstrings []string
}

func NewScraper(allowedHrefSubstrings []string, requiredHrefSubstrings []string) *Scraper {
	return &Scraper{
		requests:               make(chan request),
		HttpFileIds:            make(chan string),
		readers:                make(chan *io.Reader),
		hrefs:                  make(chan string),
		allowedHrefSubstrings:  allowedHrefSubstrings,
		requiredHrefSubstrings: requiredHrefSubstrings,
	}
}

func (s *Scraper) Start(startHref string) *Scraper {
	requestId := hash(startHref)
	go s.getURL(requestId, startHref)

	for {
		select {
		case request := <-s.requests:
			go s.storeHtml(request)
		case httpFileId := <-s.HttpFileIds:
			go s.loadHtml(httpFileId)
		case reader := <-s.readers:
			go s.findHref(reader)
		case href := <-s.hrefs:
			s.GetHttp(href)
		}
	}
}

// FIXME: nasty return if it does already exist
func (s *Scraper) GetHttp(href string) *Scraper {
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

func (s *Scraper) getURL(requestId string, url string) *Scraper {
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

func (s *Scraper) storeHtml(r request) *Scraper {
	doc, err := ioutil.ReadAll(r.response.Body)
	check(err)

	writeFile(httpDir, r.id, "html", doc)
	s.HttpFileIds <- r.id

	return s
}

func (s *Scraper) loadHtml(fileId string) *Scraper {
	reader := createReader(httpDir, fileId, "html")
	s.readers <- reader

	return s
}

func (s *Scraper) ReadDownloadedIds() *Scraper {
	fileInfos := readDir(httpDir)

	for _, file := range fileInfos {
		if file.IsDir() {
			continue
		}

		id := removeExtension(file.Name())
		s.HttpFileIds <- id
	}

	return s
}

func (s *Scraper) findHref(reader *io.Reader) *Scraper {
	doc, err := goquery.NewDocumentFromReader(*reader)
	check(err)

	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")

		if exists {
			s.hrefs <- href
		}
	})

	return s
}

func (s *Scraper) doesHtmlExist(fileId string) bool {
	return doesFileExist(httpDir, fileId, "html")
}
