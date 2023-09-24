package scraper

import (
	"fmt"
	"go-find-pepe/pkg/db"
	"go-find-pepe/pkg/environment"
	"sync"
)

type Scraper struct {
	htmlScraper  *HtmlScraper
	imageScraper *Image
	wg           *sync.WaitGroup
	done         *sync.Mutex
}

type NewScraperArguments struct {
	environment.Environment
	AllowedHrefSubstrings  []string
	RequiredHrefSubstrings []string
	AllowedImageTypes      []string
	*db.DbConnection
}

func NewScraper(arg NewScraperArguments) *Scraper {
	imageHrefs := make(chan string)

	mutex := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	r := Request{url: fmt.Sprintf("%v/health", arg.VisionApiUrl), reuseConnection: false, method: "GET"}
	_, _, success := r.Do(1)
	if !success {
		panic("Failed to do VISION_API_URL health")
	}

	html := &HtmlScraper{
		allowedHrefSubstrings:  arg.AllowedHrefSubstrings,
		requiredHrefSubstrings: arg.RequiredHrefSubstrings,
		wg:                     wg,
		done:                   mutex,
		imageHrefs:             imageHrefs,
		htmlLimit:              arg.HtmlLimit,
		db:                     arg.InitHtml(),
	}
	image := &Image{
		allowedImageTypes: arg.AllowedImageTypes,
		visionApiUrl:      arg.VisionApiUrl,
		wg:                wg,
		done:              mutex,
		imageHrefs:        imageHrefs,
		imageLimit:        arg.ImageLimit,
		classifyLimit:     arg.ClassifyLimit,
		db:                arg.InitImage(),
	}

	return &Scraper{
		imageScraper: image,
		htmlScraper:  html,
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
		s.htmlScraper.Start(startHref)
	})
	wgU.Wrapper(s.imageScraper.Start)

	s.wg.Wait()
	s.done.Unlock()
	wg.Wait()

	return s
}
