package scraper

import (
	"errors"
	"fmt"
	"go-find-pepe/pkg/db"
	"go-find-pepe/pkg/limit"
	"go-find-pepe/pkg/utils"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

type Html struct {
	allowedHrefSubstrings  []string
	requiredHrefSubstrings []string
	wg                     *sync.WaitGroup
	done                   *sync.Mutex
	imageHrefs             chan string
	htmlLimit              int8
	db                     *db.HtmlDbConnection
}

type htmlResponse struct {
	href string
	body *io.ReadCloser
}

func (s *Html) Start(startHref string) {
	hrefs := make(chan string)
	toBeScrapped := make(chan *db.Html)

	done := make(chan bool)
	wgU := WaitGroupHelper{WaitGroup: s.wg}

	wgU.Wrapper(func() {
		tx := s.db.CreateTransaction()
		defer tx.Deferral()

		tx.FindAll(func(h *db.Html) {
			s.wg.Add(1)
			toBeScrapped <- h
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

	htmlLimiter := limit.NewLimiter(s.htmlLimit)

	s.wg.Done()
	for {
		select {
		case <-done:
			s.cleanup()
			fmt.Println("HttpScraper exited")
			return
		case html := <-toBeScrapped:
			func() {
				defer s.wg.Done()
				wgU.Wrapper(func() {
					file := readFile(html.FilePath)
					defer file.Close()
					s.findHtmlHref(html.Href, file, hrefs)
				})

				wgU.Wrapper(func() {
					file := readFile(html.FilePath)
					defer file.Close()
					s.findImageHref(html.Href, file, s.imageHrefs)
				})
			}()
		case href := <-hrefs:
			wgU.Wrapper(func() {
				htmlLimiter.Add()
				defer s.wg.Done()
				defer htmlLimiter.Done()

				response, err := s.getHttp(href)
				if err != nil {
					if err.Error() == "not found" {
						// store something so it does not get picked up again
						file := io.NopCloser(strings.NewReader("404"))
						response = &htmlResponse{body: &file, href: href}
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

func (s *Html) cleanup() {
	fmt.Printf("Beginning to delete html database\n")
	tx := s.db.CreateTransaction()
	tx.DeleteAll()
	defer tx.Commit()

	fmt.Printf("Cleaned html database\n")
}

func (s *Html) findHtmlHref(parentHref string, reader io.Reader, output chan string) *Html {
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

func (s *Html) findImageHref(parentHref string, reader io.Reader, output chan string) *Html {
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

func (s *Html) getHttp(href string) (*htmlResponse, error) {
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

	return &htmlResponse{body: &response, href: href}, nil
}

func (s *Html) storeHtml(r *htmlResponse) *db.Html {
	tx := s.db.CreateTransaction()
	defer tx.Deferral()

	path := s.newPath()
	html := tx.Create(db.NewHtml{
		FilePath: path,
		Href:     r.href,
		Board:    "",
	})

	writeFile(path, *r.body)
	return html
}

func (s *Html) doesHtmlExist(href string) bool {
	tx := s.db.CreateTransaction()
	defer tx.Deferral()
	return tx.ExistsByHref(href)
}

func (s *Html) newPath() (path string) {
	fileName := createUniqueId()
	fileName = addExtension(fileName, "html")

	path = filepath.Join(getProjectPath(), HtmlDir, fileName)
	return
}
