package scraper

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-find-pepe/internal/utils"

	"github.com/PuerkitoBio/goquery"
)

type HttpScraper struct {
	allowedHrefSubstrings  []string
	requiredHrefSubstrings []string
	wg                     *sync.WaitGroup
	done                   *sync.Mutex
	imageHrefs             chan string
}

type httpResponse struct {
	href string
	body *[]byte
}

func (s *HttpScraper) Start(startHref string) {
	hrefs := make(chan string)
	toBeScrapped := make(chan string)

	done := make(chan bool)
	wgU := utils.WaitGroupUtil{WaitGroup: s.wg}

	dirPath := filepath.Join(getProjectPath(), HtmlDir)
	wgU.Wrapper(func() {
		readNestedDir(dirPath, func(path string) {
			s.wg.Add(1)
			toBeScrapped <- path
		})
	})

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

			path := s.storeHtml(request)
			s.wg.Add(1)
			toBeScrapped <- path
		},
	)

	go func() {
		s.done.Lock()
		done <- true
	}()

	s.wg.Done()
	for {
		select {
		case <-done:
			s.cleanup()
			fmt.Println("HttpScraper exited")
			return
		case path := <-toBeScrapped:
			func() {
				defer s.wg.Done()
				html := readFile(path)
				reader := bytes.NewReader(*html)
				parentHref := s.pathToUrl(path)

				wgU.Wrapper(func() {
					s.findHtmlHref(parentHref, reader, hrefs)
				})
				wgU.Wrapper(func() {
					s.findImageHref(parentHref, reader, s.imageHrefs)
				})
			}()
		case href := <-hrefs:
			wgU.Wrapper(func() {
				defer s.wg.Done()

				request, err := s.getHttp(href)

				if err != nil {
					if err.Error() == "http unallowed source" || err.Error() == "html already exists" {
						return
					} else if err.Error() == "unsuccessful response" {
						fmt.Printf("Failed request %v; ignoring\n", href)
						return
					} else {
						panic(err)
					}
				}

				path := s.storeHtml(request)

				s.wg.Add(1)
				toBeScrapped <- path
			})
		}
	}
}

// cleanup html folder after all images and other html are found so the next pass
// will find the rest
func (s *HttpScraper) cleanup() {
	path := filepath.Join(getProjectPath(), HtmlDir)
	fmt.Printf("Beginning to delete %v directory\n", path)
	err := os.RemoveAll(path)
	utils.Check(err)
	fmt.Printf("Cleaned %v directory\n", path)
}

func (s *HttpScraper) findHtmlHref(parentHref string, reader io.Reader, output chan string) *HttpScraper {
	doc, err := goquery.NewDocumentFromReader(reader)
	utils.Check(err)

	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if !exists {
			return
		}

		unallowed := [6]string{"javascript", "#", " ", "<", ">", ":"}
		for _, s := range unallowed {
			if strings.Contains(href, s) {
				return
			}
		}

		if !strings.Contains(href, "/") {
			return
		}

		cleanedHref := fixMissingHttps(href)
		hostname := getHostname(cleanedHref)
		if hostname == "" && cleanedHref[0] != '/' {
			cleanedHref = parentHref + cleanedHref
		}
		s.wg.Add(1)
		output <- cleanedHref
	})

	return s
}

func (s *HttpScraper) findImageHref(parentHref string, reader io.Reader, output chan string) *HttpScraper {
	doc, err := goquery.NewDocumentFromReader(reader)
	utils.Check(err)

	fileSelection := doc.Find("div .file").Find("div .fileText")
	fileSelection.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")
		if !exists {
			return
		}

		unallowed := [6]string{"javascript", "#", " ", "<", ">", ":"}
		for _, s := range unallowed {
			if strings.Contains(href, s) {
				return
			}
		}

		cleanedHref := fixMissingHttps(href)
		hostname := getHostname(cleanedHref)
		if hostname == "" && cleanedHref[0] != '/' {
			cleanedHref = parentHref + cleanedHref
		}

		s.wg.Add(1)
		output <- cleanedHref
	})

	return s
}

func (s *HttpScraper) getHttp(href string) (*httpResponse, error) {
	if s.doesHtmlExist(href) {
		return nil, errors.New("html already exists")
	}

	cleanedHref := fixMissingHttps(href)
	hostname := getHostname(cleanedHref)

	correctAllowedSubstrings := stringShouldContainOneFilter(hostname, s.allowedHrefSubstrings)
	correctRequiredSubstrings := stringShouldContainAllFilters(cleanedHref, s.requiredHrefSubstrings)

	if !correctAllowedSubstrings || !correctRequiredSubstrings {
		return nil, errors.New("http unallowed source")
	}

	request := Request{url: cleanedHref, reuseConnection: true, method: "GET"}
	response, _, success := request.Do(1)

	if !success {
		return nil, errors.New("unsuccessful response")
	}

	data, err := ioutil.ReadAll(response)
	utils.Check(err)

	return &httpResponse{body: &data, href: href}, nil
}

func (s *HttpScraper) storeHtml(r *httpResponse) string {
	fileName := s.transformUrlIntoFilename(r.href)
	path := filepath.Join(getProjectPath(), HtmlDir, fileName)
	writeFile(path, r.body)
	return path
}

func (s *HttpScraper) doesHtmlExist(href string) bool {
	fileName := s.transformUrlIntoFilename(href)
	path := filepath.Join(getProjectPath(), HtmlDir, fileName)
	return doesFileExist(path)
}

func (s *HttpScraper) transformUrlIntoFilename(href string) (fileName string) {
	fileName = href
	if fileName[len(fileName)-1] == '/' {
		fileName = fileName[0 : len(fileName)-1]
	}
	fileName = addExtension(fileName, "html")
	return
}

func (s *HttpScraper) pathToUrl(path string) (url string) {
	storage := filepath.Join(getProjectPath(), HtmlDir) + "/"
	url = strings.Replace(removeExtension(path), storage, "", 1) + "/"
	url = strings.Replace(url, "https:/", "https://", 1)

	return
}
