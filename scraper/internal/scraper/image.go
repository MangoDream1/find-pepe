package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"go-find-pepe/internal/utils"

	"github.com/PuerkitoBio/goquery"
)

const ImageDir = "data/image"
const MaybeDir = ImageDir + "/maybe"
const PepeDir = ImageDir + "/pepe"
const NonPepeDir = ImageDir + "/non-pepe"

const PepeThreshold = 0.9
const MaybeThreshold = 0.3

const VisionApiUrl = "http://localhost:5000"
const VisionImageKey = "file"

type ImageScraper struct {
	hrefs             chan string
	requests          chan imageRequest
	imageReaders      chan *io.Reader
	classifiedImages  chan classifiedImage
	allowedImageTypes []string
}

type imageRequest struct {
	fileName string
	response *http.Response
}

type classifiedImage struct {
	propability float32
	fileName    string
	image       []byte
}

type visionResponse struct {
	Score float32 `json:"score"`
}

func newImageScraper(allowedImageTypes []string) *ImageScraper {
	return &ImageScraper{
		hrefs:             make(chan string),
		requests:          make(chan imageRequest),
		imageReaders:      make(chan *io.Reader),
		classifiedImages:  make(chan classifiedImage),
		allowedImageTypes: allowedImageTypes,
	}
}

func (s *ImageScraper) Start() *ImageScraper {
	for {
		select {
		case classifiedImage := <-s.classifiedImages:
			go s.storeImage(classifiedImage)
		case request := <-s.requests:
			go s.classifyImage(request)
		case href := <-s.hrefs:
			go s.getImage(href)
		}
	}
}

// FIXME: nasty return if it does already exist
func (s *ImageScraper) getImage(href string) *ImageScraper {
	cleanedHref := cleanUpUrl(href)

	correctRequiredSubstrings := stringShouldContainOneFilter(cleanedHref, s.allowedImageTypes)
	if !correctRequiredSubstrings {
		return s
	}

	fileName := s.transformUrlIntoFilename(cleanedHref)
	if s.doesImageExist(fileName) {
		return s
	}

	response, success, canRetry := getURL(fileName, cleanedHref)

	if success {
		s.requests <- imageRequest{fileName: fileName, response: response}
	} else if canRetry {
		fmt.Printf("Retrying url: %v\n", href)
		s.hrefs <- href
	}

	return s
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

func (s *ImageScraper) classifyImage(r imageRequest) *ImageScraper {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("An error has occurred while trying to classify an image with name: %v; error: %v \n", r.fileName, err)
		}
	}()

	doc, err := ioutil.ReadAll(r.response.Body)
	utils.Check(err)

	b, w := createSingleFileMultiPart(VisionImageKey, r.fileName, doc)
	ct := w.FormDataContentType()

	res, err := http.Post(VisionApiUrl, ct, b)
	utils.Check(err)

	if res.StatusCode != 200 {
		panic(fmt.Sprintf("Unsuccessful %v POST with code %v", VisionApiUrl, res.Status))
	}

	data, err := ioutil.ReadAll(res.Body)
	utils.Check(err)

	var vRes visionResponse
	err = json.Unmarshal(data, &vRes)
	utils.Check(err)

	s.classifiedImages <- classifiedImage{propability: vRes.Score, image: doc, fileName: r.fileName}
	return s
}

func (s *ImageScraper) storeImage(img classifiedImage) *ImageScraper {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("An error has occurred while trying to store an image with name: %v; error: %v \n", img.fileName, err)
		}
	}()

	if img.propability >= PepeThreshold {
		writeFile(PepeDir, img.fileName, img.image)
	} else if img.propability >= MaybeThreshold {
		writeFile(MaybeDir, img.fileName, img.image)
	} else {
		writeFile(NonPepeDir, img.fileName, img.image)
	}

	return s
}

func (s *ImageScraper) doesImageExist(fileName string) bool {
	return doesFileExist(PepeDir, fileName) || doesFileExist(NonPepeDir, fileName) || doesFileExist(MaybeDir, fileName)
}

func (s *ImageScraper) transformUrlIntoFilename(url string) string {
	p := strings.Split(url, `/`)
	return strings.Join(p[2:], `/`)
}
