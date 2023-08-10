package scraper

import (
	"go-find-pepe/internal/utils"
	"io"
	"os"
	"sync"
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

func (s *Scraper) Start(startHref string) *Scraper {
	var mutex sync.Mutex
	var wg sync.WaitGroup
	wgU := utils.WaitGroupUtil{WaitGroup: &wg}

	wgU.Wrapper(func() {
		s.httpScraper.Start(&mutex, startHref)
	})
	wgU.Wrapper(func() {
		s.imageScraper.Start(&mutex)
	})

	wg.Wait()

	return s
}
