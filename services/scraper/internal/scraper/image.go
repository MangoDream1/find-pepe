package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"go-find-pepe/internal/utils"

	"github.com/PuerkitoBio/goquery"
)

const ImageDir = "data/image"
const MaybeDir = ImageDir + "/maybe"
const PepeDir = ImageDir + "/pepe"
const NonPepeDir = ImageDir + "/non-pepe"
const UnclassifiedDir = ImageDir + "/unclassified"

var FileDirectories = [4]string{MaybeDir, PepeDir, NonPepeDir, UnclassifiedDir}

const PepeThreshold = 0.9
const MaybeThreshold = 0.3

const VisionImageKey = "file"

type ImageScraper struct {
	httpReaders       chan io.Reader
	allowedImageTypes []string
	visionApiUrl      string
	wg                *sync.WaitGroup
	done              *sync.Mutex
}

type imageRequest struct {
	fileName string
	response *http.Response
}

type visionResponse struct {
	Score float32 `json:"score"`
}

func (s *ImageScraper) Start() {
	hrefs := make(chan string)
	toBeClassified := make(chan string)

	done := make(chan bool)
	wgU := utils.WaitGroupUtil{WaitGroup: s.wg}

	dirPath := filepath.Join(getProjectPath(), UnclassifiedDir)

	wgU.Wrapper(func() {
		readNestedDir(dirPath, func(path string) {
			s.wg.Add(1)
			toBeClassified <- path
		})
	})

	go func() {
		s.done.Lock()
		done <- true
	}()

	s.wg.Done()
	for {
		select {
		case <-done:
			fmt.Println("ImageScraper exited")
			return
		case reader := <-s.httpReaders:
			wgU.Wrapper(func() {
				defer s.wg.Done()
				s.findHref(reader, hrefs)
			})
		case path := <-toBeClassified:
			wgU.Wrapper(func() {
				defer s.wg.Done()
				defer func() {
					if err := recover(); err != nil {
						if strings.Contains(err.(error).Error(), "no such file or directory") {
							fmt.Printf("Image %v does not exist anymore; ignoring\n", path)
							return
						} else {
							panic(err)
						}
					}
				}()

				s.classifyImageByPath(path)
			})
		case href := <-hrefs:
			wgU.Wrapper(func() {
				defer s.wg.Done()
				request, err := s.getImage(href)
				if err != nil {
					if err.Error() == "image type not allowed" || err.Error() == "image already exists" {
						return
					} else if err.Error() == "unsuccessful response" {
						fmt.Printf("Failed request %v; ignoring", href)
						return
					} else {
						panic(err)
					}
				}

				s.storeImageRequest(request, toBeClassified)
			})
		}
	}
}

func (s *ImageScraper) classifyImageByPath(path string) {
	blob := readFile(path)
	probability := s.retrieveImageProbability(path, blob)

	if probability >= PepeThreshold {
		newPath := replaceDir(path, UnclassifiedDir, PepeDir)
		moveFile(path, newPath)
	} else if probability >= MaybeThreshold {
		newPath := replaceDir(path, UnclassifiedDir, MaybeDir)
		moveFile(path, newPath)
	} else {
		newPath := replaceDir(path, UnclassifiedDir, NonPepeDir)
		moveFile(path, newPath)
	}
}

func (s *ImageScraper) storeImageRequest(request *imageRequest, output chan string) string {
	path := filepath.Join(getProjectPath(), UnclassifiedDir, request.fileName)

	blob, err := ioutil.ReadAll(request.response.Body)
	utils.Check(err)

	writeFile(path, blob)
	s.wg.Add(1)
	output <- path
	return path
}

func (s *ImageScraper) getImage(href string) (*imageRequest, error) {
	cleanedHref := fixMissingHttps(href)

	correctRequiredSubstrings := stringShouldContainOneFilter(cleanedHref, s.allowedImageTypes)
	if !correctRequiredSubstrings {
		return nil, errors.New("image type not allowed")
	}

	fileName := s.transformUrlIntoFilename(cleanedHref)
	if s.doesImageExist(fileName) {
		return nil, errors.New("image already exists")
	}

	response, success := getURL(cleanedHref, 1)

	if !success {
		return nil, errors.New("unsuccessful response")
	}

	return &imageRequest{fileName: fileName, response: response}, nil
}

func (s *ImageScraper) findHref(reader io.Reader, output chan string) *ImageScraper {
	doc, err := goquery.NewDocumentFromReader(reader)
	utils.Check(err)

	fileSelection := doc.Find("div .file").Find("div .fileText")
	fileSelection.Find("a").Each(func(i int, selection *goquery.Selection) {
		href, exists := selection.Attr("href")

		if exists {
			s.wg.Add(1)
			output <- href
		}
	})

	return s
}

func (s *ImageScraper) retrieveImageProbability(filePath string, blob []byte) float32 {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("An error has occurred while trying to classify an image with name: %v \n", filePath)
			panic(err)
		}
	}()

	b, w := createSingleFileMultiPart(VisionImageKey, filePath, blob)
	ct := w.FormDataContentType()

	// TODO: use utils.getURL instead
	res, err := http.Post(s.visionApiUrl, ct, b)
	utils.Check(err)

	// assume that if 500 was returned; something is wrong with the file
	// move to maybe
	if res.StatusCode == 500 {
		newPath := replaceDir(filePath, UnclassifiedDir, MaybeDir)
		moveFile(filePath, newPath)
		fmt.Printf("Unsuccessful %v POST with code 500; moved the image %v\n", s.visionApiUrl, filePath)
	}

	if res.StatusCode != 200 {
		panic(fmt.Errorf("unsuccessful %v POST with code %v", s.visionApiUrl, res.Status))
	}

	data, err := ioutil.ReadAll(res.Body)
	utils.Check(err)

	var vRes visionResponse
	err = json.Unmarshal(data, &vRes)
	utils.Check(err)

	return vRes.Score
}

func (s *ImageScraper) doesImageExist(fileName string) bool {
	for _, dir := range FileDirectories {
		path := filepath.Join(getProjectPath(), dir, fileName)
		if !doesFileExist(path) {
			return false
		}
	}

	return true
}

func (s *ImageScraper) transformUrlIntoFilename(url string) string {
	p := strings.Split(url, `/`)
	return strings.Join(p[2:], `/`)
}
