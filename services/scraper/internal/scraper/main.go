package scraper

import (
	"io"
	"os"
)

type Scraper struct {
	httpScraper  *HttpScraper
	imageScraper *ImageScraper
	httpReaders  chan io.Reader
}

func NewScraper(allowedHrefSubstrings []string, requiredHrefSubstrings []string, allowedImageTypes []string) *Scraper {
	httpReaders := make(chan io.Reader)

	visionApiUrl := os.Getenv("VISION_API_URL")
	if visionApiUrl == "" {
		panic("VISION_API_URL unset")
	}

	httpScraper := newHttpScraper(httpReaders, allowedHrefSubstrings, requiredHrefSubstrings)
	imageScraper := newImageScraper(visionApiUrl, allowedImageTypes)

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
