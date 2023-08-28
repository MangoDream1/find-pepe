package scraper

import (
	"errors"
	"fmt"
	"go-find-pepe/internal/utils"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
	body *io.ReadCloser
}

func (s *HttpScraper) Start(startHref string) {
	hrefs := make(chan string)
	toBeScrapped := make(chan string)

	done := make(chan bool)
	wgU := WaitGroupUtil{WaitGroup: s.wg}

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
				parentHref := s.pathToUrl(path)

				wgU.Wrapper(func() {
					file := readFile(path)
					defer file.Close()
					s.findHtmlHref(parentHref, file, hrefs)
				})
				wgU.Wrapper(func() {
					file := readFile(path)
					defer file.Close()
					s.findImageHref(parentHref, file, s.imageHrefs)
				})
			}()
		case href := <-hrefs:
			wgU.Wrapper(func() {
				defer s.wg.Done()

				response, err := s.getHttp(href)
				if err != nil {
					if err.Error() == "not found" {
						// store something so it does not get picked up again
						file := io.NopCloser(strings.NewReader("404"))
						response = &httpResponse{body: &file, href: href}
					} else if err.Error() == "http unallowed source" || err.Error() == "html already exists" {
						return
					} else if err.Error() == "unsuccessful response" {
						fmt.Printf("Failed request %v; ignoring\n", href)
						return
					} else {
						panic(err)
					}
				}

				defer (*response.body).Close()
				path := s.storeHtml(response)

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
	response, statusCode, success := request.Do(1)

	if statusCode == 404 {
		return nil, errors.New("not found")
	}

	if !success {
		return nil, errors.New("unsuccessful response")
	}

	return &httpResponse{body: &response, href: href}, nil
}

func (s *HttpScraper) storeHtml(r *httpResponse) string {
	fileName := s.transformUrlIntoFilename(r.href)
	path := filepath.Join(getProjectPath(), HtmlDir, fileName)
	writeFile(path, *r.body)
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
