package scraper

import (
	"io"
	"net/http"
)

type Scraper struct {
	httpScraper *HttpScraper
	httpReaders chan *io.Reader
}

type request struct {
	id       string
	response *http.Response
}

func NewScraper(allowedHrefSubstrings []string, requiredHrefSubstrings []string) *Scraper {
	httpReaders := make(chan *io.Reader)

	httpScraper := newHttpScraper(allowedHrefSubstrings, requiredHrefSubstrings, &httpReaders)

	return &Scraper{
		httpReaders: httpReaders,
		httpScraper: httpScraper,
	}
}

func (s *Scraper) ReadDownloadedIds() *Scraper {
	s.httpScraper.readDownloadedIds()
	return s
}

func (s *Scraper) Start(startHref string) *Scraper {
	go s.httpScraper.Start(startHref)

	requestId := hash(startHref)

	go s.httpScraper.getURL(requestId, startHref)

	for {
		select {
		case reader := <-s.httpReaders:
			go s.httpScraper.findHref(reader)
		}
	}
}
