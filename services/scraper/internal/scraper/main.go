package scraper

import (
	"fmt"
	"os"
	"sync"
)

type Scraper struct {
	httpScraper  *HttpScraper
	imageScraper *Image
	wg           *sync.WaitGroup
	done         *sync.Mutex
}

func NewScraper(allowedHrefSubstrings []string, requiredHrefSubstrings []string, allowedImageTypes []string) *Scraper {
	imageHrefs := make(chan string)

	mutex := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	visionApiUrl := os.Getenv("VISION_API_URL")
	if visionApiUrl == "" {
		panic("VISION_API_URL unset")
	}

	r := Request{url: fmt.Sprintf("%v/health", visionApiUrl), reuseConnection: false, method: "GET"}
	_, _, success := r.Do(1)
	if !success {
		panic("Failed to do VISION_API_URL health")
	}

	httpScraper := &HttpScraper{
		allowedHrefSubstrings:  allowedHrefSubstrings,
		requiredHrefSubstrings: requiredHrefSubstrings,
		wg:                     wg,
		done:                   mutex,
		imageHrefs:             imageHrefs,
	}
	imageScraper := &Image{
		allowedImageTypes: allowedImageTypes,
		visionApiUrl:      visionApiUrl,
		wg:                wg,
		done:              mutex,
		imageHrefs:        imageHrefs,
	}

	return &Scraper{
		imageScraper: imageScraper,
		httpScraper:  httpScraper,
		wg:           wg,
		done:         mutex,
	}
}

func (s *Scraper) Start(startHref string) *Scraper {
	wg := &sync.WaitGroup{}
	wgU := WaitGroupUtil{WaitGroup: wg}

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
