package scraper

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"go-find-pepe/internal/utils"

	"github.com/PuerkitoBio/goquery"
)

const imageDir = "data/image"

type ImageScraper struct {
	imageFileNames    chan string
	hrefs             chan string
	requests          chan imageRequest
	imageReaders      chan *io.Reader
	allowedImageTypes []string
}

type imageRequest struct {
	fileName string
	response *http.Response
}

func newImageScraper(allowedImageTypes []string) *ImageScraper {
	return &ImageScraper{
		hrefs:             make(chan string),
		imageFileNames:    make(chan string),
		requests:          make(chan imageRequest),
		imageReaders:      make(chan *io.Reader),
		allowedImageTypes: allowedImageTypes,
	}
}

func (s *ImageScraper) Start() *ImageScraper {
	for {
		select {
		// case httpFileId := <-s.httpFileIds:
		// go s.loadHtml(httpFileId)
		case request := <-s.requests:
			go s.storeImage(request)
		case href := <-s.hrefs:
			go func() {
				defer func() {
					if err := recover(); err != nil {
						s.hrefs <- href
					}
				}()

				s.getImage(href)

			}()

		}
	}
}

// FIXME: nasty return if it does already exist
func (s *ImageScraper) getImage(href string) *ImageScraper {
	cleanedHref := cleanUpUrl(href)
	fileName := s.transformUrlIntoFilename(cleanedHref)

	correctRequiredSubstrings := stringShouldContainOneFilter(cleanedHref, s.allowedImageTypes)
	if !correctRequiredSubstrings {
		return s
	}

	if s.doesImageExist(fileName) {
		return s
	}

	go s.getURL(fileName, cleanedHref)

	return s
}

func (s *ImageScraper) findHref(reader *io.Reader) *ImageScraper {
	doc, err := goquery.NewDocumentFromReader(*reader)
	utils.Check(err)

	fileSelection := doc.Find("div .file").Find("div .fileText")
	fileSelection.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")

		if exists {
			s.hrefs <- href
		}
	})

	return s
}

func (s *ImageScraper) getURL(fileName string, url string) *ImageScraper {
	response, err := http.Get(url)

	if err != nil {
		msg := err.Error()

		if stringShouldContainOneFilter(msg, []string{"timeout", "connection reset"}) {
			fmt.Printf("Retrying %v after timeout \n", url)
			s.hrefs <- url
			return s
		}

		fmt.Printf("Failed to GET %v; error %v; ignoring \n", url, msg)
		return s
	}

	if response.StatusCode == 503 {
		fmt.Printf("503 response %v; retrying \n", url)
		s.hrefs <- url
		return s
	}

	if response.StatusCode == 404 {
		fmt.Printf("404 response: %v; ignoring \n", url)
		return s
	}

	if response.StatusCode != 200 {
		panic(fmt.Sprintf("Non-OK response: %v %v", url, response.StatusCode))
	}

	fmt.Printf("Successfully fetched %v \n", url)
	s.requests <- imageRequest{fileName: fileName, response: response}
	return s
}

func (s *ImageScraper) storeImage(r imageRequest) *ImageScraper {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("An error has occurred while trying to store an image with name: %v; error: %err \n", r.fileName, err)
		}
	}()

	doc, err := ioutil.ReadAll(r.response.Body)
	utils.Check(err)

	writeFile(imageDir, r.fileName, doc)
	s.imageFileNames <- r.fileName

	return s
}

func (s *ImageScraper) loadImage(fileName string) *ImageScraper {
	reader := createReader(imageDir, fileName)
	s.imageReaders <- reader

	return s
}

func (s *ImageScraper) doesImageExist(fileName string) bool {
	return doesFileExist(imageDir, fileName)
}

func (s *ImageScraper) transformUrlIntoFilename(url string) string {
	p := strings.Split(url, `/`)
	return strings.Join(p[2:], `/`)
}
