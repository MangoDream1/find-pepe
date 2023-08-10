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
	hrefs             chan string
	imageReaders      chan *io.Reader
	toBeClassified    chan string
	allowedImageTypes []string
	visionApiUrl      string
}

type imageRequest struct {
	fileName string
	response *http.Response
}

type visionResponse struct {
	Score float32 `json:"score"`
}

func newImageScraper(httpReaders chan io.Reader, visionApiUrl string, allowedImageTypes []string) *ImageScraper {
	return &ImageScraper{
		httpReaders:       httpReaders,
		hrefs:             make(chan string),
		imageReaders:      make(chan *io.Reader),
		toBeClassified:    make(chan string),
		allowedImageTypes: allowedImageTypes,
		visionApiUrl:      visionApiUrl,
	}
}

func (s *ImageScraper) Start() *ImageScraper {
	dirPath := filepath.Join(getProjectPath(), UnclassifiedDir)
	go readNestedDir(dirPath, s.toBeClassified)

	for {
		select {
		case reader := <-s.httpReaders:
			go s.findHref(reader)
		case path := <-s.toBeClassified:
			go s.classifyImageByPath(path)
		case href := <-s.hrefs:
			request, err := s.getImage(href)
			if err != nil {
				if err.Error() == "image type not allowed" || err.Error() == "image already exists" {
					continue
				}
				panic(err)
			}

			s.storeImageRequest(request, s.toBeClassified)
		}
	}
}

func (s *ImageScraper) classifyImageByPath(path string) {
	blob := readFile(path)
	probability := s.retrieveImageProbability(path, blob)

	if probability >= PepeThreshold {
		newPath := replaceDir(path, UnclassifiedDir, PepeDir)
		writeFile(newPath, blob)
	} else if probability >= MaybeThreshold {
		newPath := replaceDir(path, UnclassifiedDir, MaybeDir)
		writeFile(newPath, blob)
	} else {
		newPath := replaceDir(path, UnclassifiedDir, NonPepeDir)
		writeFile(newPath, blob)
	}

	deleteFile(path)
}

func (s *ImageScraper) storeImageRequest(request *imageRequest, output chan string) string {
	blob, err := ioutil.ReadAll(request.response.Body)
	utils.Check(err)

	path := filepath.Join(getProjectPath(), UnclassifiedDir, request.fileName)
	writeFile(path, blob)
	output <- path
	return path
}

func (s *ImageScraper) getImage(href string) (*imageRequest, error) {
	cleanedHref := cleanUpUrl(href)

	correctRequiredSubstrings := stringShouldContainOneFilter(cleanedHref, s.allowedImageTypes)
	if !correctRequiredSubstrings {
		return nil, errors.New("image type not allowed")
	}

	fileName := s.transformUrlIntoFilename(cleanedHref)
	if s.doesImageExist(fileName) {
		return nil, errors.New("image already exists")
	}

	response, success, canRetry := getURL(fileName, cleanedHref)

	if success {
		return &imageRequest{fileName: fileName, response: response}, nil
	} else if canRetry {
		fmt.Printf("Retrying url: %v\n", href)
		return s.getImage(href)
	}

	return nil, errors.New("unsuccessful response")
}

func (s *ImageScraper) findHref(reader io.Reader) *ImageScraper {
	doc, err := goquery.NewDocumentFromReader(reader)
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

func (s *ImageScraper) retrieveImageProbability(fileName string, blob []byte) float32 {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("An error has occurred while trying to classify an image with name: %v \n", fileName)
			panic(err)
		}
	}()

	b, w := createSingleFileMultiPart(VisionImageKey, fileName, blob)
	ct := w.FormDataContentType()

	res, err := http.Post(s.visionApiUrl, ct, b)
	utils.Check(err)

	if res.StatusCode != 200 {
		panic(fmt.Sprintf("Unsuccessful %v POST with code %v", s.visionApiUrl, res.Status))
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
