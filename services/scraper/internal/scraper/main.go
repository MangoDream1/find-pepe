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
	httpReaders := make(chan io.Reader) // FIXME: both image and http use this same reader; should fan out to both; https://stackoverflow.com/questions/28527038/go-one-channel-with-multiple-listeners

	visionApiUrl := os.Getenv("VISION_API_URL")
	if visionApiUrl == "" {
		panic("VISION_API_URL unset")
	}

	httpScraper := newHttpScraper(httpReaders, allowedHrefSubstrings, requiredHrefSubstrings)
	imageScraper := newImageScraper(httpReaders, visionApiUrl, allowedImageTypes)

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
	// done := make(chan int)

	// go s.httpScraper.Start(startHref)
	s.imageScraper.Start()

	// TODO: actually await here
	// <-done

	return s
}
