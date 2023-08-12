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
	wg           *sync.WaitGroup
	done         *sync.Mutex
}

func NewScraper(allowedHrefSubstrings []string, requiredHrefSubstrings []string, allowedImageTypes []string) *Scraper {
	httpReaders := make(chan io.Reader) // FIXME: both image and http use this same reader; should fan out to both; https://stackoverflow.com/questions/28527038/go-one-channel-with-multiple-listeners

	mutex := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	visionApiUrl := os.Getenv("VISION_API_URL")
	if visionApiUrl == "" {
		panic("VISION_API_URL unset")
	}

	httpScraper := &HttpScraper{
		httpReaders:            httpReaders,
		allowedHrefSubstrings:  allowedHrefSubstrings,
		requiredHrefSubstrings: requiredHrefSubstrings,
		wg:                     wg,
		done:                   mutex,
	}
	imageScraper := &ImageScraper{
		httpReaders:       httpReaders,
		allowedImageTypes: allowedImageTypes,
		visionApiUrl:      visionApiUrl,
		wg:                wg,
		done:              mutex,
	}

	return &Scraper{
		httpReaders:  httpReaders,
		imageScraper: imageScraper,
		httpScraper:  httpScraper,
		wg:           wg,
		done:         mutex,
	}
}

func (s *Scraper) Start(startHref string) *Scraper {
	wg := &sync.WaitGroup{}
	wgU := utils.WaitGroupUtil{WaitGroup: wg}

	s.done.Lock()

	s.wg.Add(2)

	wgU.Wrapper(func() {
		s.httpScraper.Start(startHref)
	})
	wgU.Wrapper(s.imageScraper.Start)

	s.wg.Wait()
	s.done.Unlock()
	wg.Wait()

	return s
}
