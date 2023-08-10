package scraper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"go-find-pepe/internal/utils"

	"github.com/PuerkitoBio/goquery"
)

const httpDir = "data/http"

type HttpScraper struct {
	httpReaders            chan io.Reader
	allowedHrefSubstrings  []string
	requiredHrefSubstrings []string
	wg                     sync.WaitGroup
}

type httpRequest struct {
	id   string
	href string
	data []byte
}

func newHttpScraper(httpReaders chan io.Reader, allowedHrefSubstrings []string, requiredHrefSubstrings []string) *HttpScraper {
	return &HttpScraper{
		httpReaders:            httpReaders,
		allowedHrefSubstrings:  allowedHrefSubstrings,
		requiredHrefSubstrings: requiredHrefSubstrings,
		wg:                     sync.WaitGroup{},
	}
}

func (s *HttpScraper) Start(mutex *sync.Mutex, startHref string) {
	mutex.Lock()

	hrefs := make(chan string)
	requests := make(chan *httpRequest)

	done := make(chan bool)
	wgU := utils.WaitGroupUtil{WaitGroup: &s.wg}

	wgU.Wrapper(
		func() {
			request, err := s.getHttp(startHref)

			if err != nil {
				if err.Error() == "html already exists" {
					fmt.Printf("startHref already exists %v; continuing\n", startHref)
					return
				}

				panic(fmt.Errorf("failed to get startHref %v; %v", startHref, err))
			}

			s.storeHtml(request)
			s.wg.Add(1)
			requests <- request
		},
	)

	go func() {
		s.wg.Wait()
		done <- true
		mutex.Unlock()
	}()

	for {
		select {
		case <-done:
			fmt.Println("HttpScraper exited")
			return
		case request := <-requests:
			func() {
				defer s.wg.Done()
				wgU.Wrapper(func() {
					s.findHref(request, hrefs)
				})

				s.wg.Add(1)
				s.httpReaders <- bytes.NewBuffer(request.data)
			}()

		case href := <-hrefs:
			wgU.Wrapper(func() {
				defer s.wg.Done()

				request, err := s.getHttp(href)
				if err != nil {
					if err.Error() == "http unallowed source" || err.Error() == "html already exists" {
						return
					} else if err.Error() == "unsuccessful response" {
						fmt.Printf("Failed request %v; ignoring", href)
						return
					} else {
						panic(err)
					}
				}

				s.storeHtml(request)

				s.wg.Add(1)
				requests <- request
			})
		}
	}
}

func (s *HttpScraper) findHref(request *httpRequest, output chan string) *HttpScraper {
	reader := bytes.NewReader(request.data)
	doc, err := goquery.NewDocumentFromReader(reader)
	utils.Check(err)

	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")

		unallowed := [2]string{"javascript", "#"}
		for _, s := range unallowed {
			if strings.Contains(href, s) {
				return
			}
		}

		if !strings.Contains(href, "/") {
			return
		}

		cleanedHref := cleanUpUrl(href)

		hostname := getHostname(cleanedHref)
		if hostname == "" && cleanedHref[0] != '/' {
			cleanedHref = request.href + cleanedHref
		}

		if exists {
			s.wg.Add(1)
			output <- cleanedHref
		}
	})

	return s
}

func (s *HttpScraper) getHttp(href string) (*httpRequest, error) {
	requestId := hash(href)
	if s.doesHtmlExist(href) {
		return nil, errors.New("html already exists")
	}

	cleanedHref := cleanUpUrl(href)
	hostname := getHostname(cleanedHref)

	correctAllowedSubstrings := stringShouldContainOneFilter(hostname, s.allowedHrefSubstrings)
	correctRequiredSubstrings := stringShouldContainAllFilters(cleanedHref, s.requiredHrefSubstrings)

	if !correctAllowedSubstrings || !correctRequiredSubstrings {
		return nil, errors.New("http unallowed source")
	}

	response, success := getURL(cleanedHref, 1)

	if !success {
		return nil, errors.New("unsuccessful response")
	}

	data, err := ioutil.ReadAll(response.Body)
	utils.Check(err)

	return &httpRequest{id: requestId, data: data, href: href}, nil
}

func (s *HttpScraper) storeHtml(r *httpRequest) string {
	fileName := s.createFileName(r.href)
	path := filepath.Join(getProjectPath(), httpDir, fileName)
	writeFile(path, r.data)
	return path
}

func (s *HttpScraper) doesHtmlExist(href string) bool {
	fileName := s.createFileName(href)
	path := filepath.Join(getProjectPath(), httpDir, fileName)
	return doesFileExist(path)
}

func (s *HttpScraper) createFileName(href string) (fileName string) {
	fileName = href
	fileName = strings.Replace(fileName, "https://", "", 1)
	fileName = strings.Replace(fileName, "www.", "", 1)
	if fileName[len(fileName)-1] == '/' {
		fileName = fileName[0 : len(fileName)-1]
	}
	fileName = addExtension(fileName, "html")
	return
}
