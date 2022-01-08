package scraper

import (
	"io"
)

type Scraper struct {
	httpScraper  *HttpScraper
	imageScraper *ImageScraper
	httpReaders  chan io.Reader
}

func NewScraper(allowedHrefSubstrings []string, requiredHrefSubstrings []string, allowedImageTypes []string) *Scraper {
	httpReaders := make(chan io.Reader)

	httpScraper := newHttpScraper(httpReaders, allowedHrefSubstrings, requiredHrefSubstrings)
	imageScraper := newImageScraper(allowedImageTypes)

	return &Scraper{
		httpReaders:  httpReaders,
		imageScraper: imageScraper,
		httpScraper:  httpScraper,
	}
}

// TODO: refactor
// func (s *Scraper) ReadDownloadedIds() *Scraper {
// 	s.httpScraper.readDownloadedIds()
// 	return s
// }

func (s *Scraper) Start(startHref string) *Scraper {
	go s.httpScraper.Start(startHref)
	go s.imageScraper.Start()

	for {
		reader := <-s.httpReaders
		go s.imageScraper.findHref(reader)
	}
}
