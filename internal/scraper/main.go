package scraper

import "io"

type Scraper struct {
	httpScraper *HttpScraper
	readers     chan *io.Reader
	httpFileIds chan string
}

func NewScraper(allowedHrefSubstrings []string, requiredHrefSubstrings []string) *Scraper {
	httpFileIds := make(chan string)
	readers := make(chan *io.Reader)

	httpScraper := newHttpScraper(allowedHrefSubstrings, requiredHrefSubstrings, &readers, &httpFileIds)

	return &Scraper{
		readers:     readers,
		httpFileIds: httpFileIds,
		httpScraper: httpScraper,
	}
}

func (s *Scraper) ReadDownloadedIds() *Scraper {
	fileInfos := readDir(httpDir)

	for _, file := range fileInfos {
		if file.IsDir() {
			continue
		}

		id := removeExtension(file.Name())
		s.httpFileIds <- id
	}

	return s
}

func (s *Scraper) Start(startHref string) *Scraper {
	requestId := hash(startHref)

	go s.httpScraper.getURL(requestId, startHref)

	for {
		select {
		case httpFileId := <-s.httpFileIds:
			go s.httpScraper.loadHtml(httpFileId)
		case reader := <-s.readers:
			go s.httpScraper.findHref(reader)
		case request := <-s.httpScraper.requests:
			go s.httpScraper.storeHtml(request)
		case href := <-s.httpScraper.hrefs:
			go s.httpScraper.GetHttp(href)
		}
	}
}
