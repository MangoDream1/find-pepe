package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-find-pepe/pkg/constants"
	"go-find-pepe/pkg/db"
	"go-find-pepe/pkg/limit"
	"go-find-pepe/pkg/utils"
	"io"
	"io/ioutil"
	"path/filepath"
	"sync"
)

type Image struct {
	allowedImageTypes []string
	visionApiUrl      string
	wg                *sync.WaitGroup
	done              *sync.Mutex
	imageHrefs        chan string
	imageLimit        int8
	classifyLimit     int8
	db                *db.ImageDbConnection
}

type imageResponse struct {
	href string
	body *io.ReadCloser
}

type visionResponse struct {
	Score float32 `json:"score"`
}

func (s *Image) Start() {
	toBeClassified := make(chan *db.Image)

	done := make(chan bool)
	wgU := WaitGroupUtil{WaitGroup: s.wg}

	wgU.Wrapper(func() {
		tx := s.db.CreateImageTransaction()
		defer tx.Deferral()

		tx.FindAllUnclassified(func(i *db.Image) {
			s.wg.Add(1)
			toBeClassified <- i
		})
	})

	hrefLimiter := limit.NewLimiter(s.imageLimit)
	classifyLimiter := limit.NewLimiter(s.classifyLimit)

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
		case img := <-toBeClassified:
			wgU.Wrapper(func() {
				classifyLimiter.Add()
				defer s.wg.Done()
				defer classifyLimiter.Done()
				s.classifyImage(img)
			})

		case href := <-s.imageHrefs:
			wgU.Wrapper(func() {
				hrefLimiter.Add()

				defer s.wg.Done()
				defer hrefLimiter.Done()

				response, err := s.getImage(href)

				if err != nil {
					if err.Error() == "image type not allowed" || err.Error() == "image already exists" {
						return
					} else if err.Error() == "unsuccessful response" {
						fmt.Printf("Failed request %v; ignoring\n", href)
						return
					} else {
						panic(err)
					}
				}
				defer (*response.body).Close()

				s.wg.Add(1)
				toBeClassified <- s.storeImageResponse(response)
			})
		}
	}
}

func (s *Image) classifyImage(img *db.Image) {
	file := readFile(img.FilePath)
	defer file.Close()

	probability, err := s.retrieveImageProbability(img.FilePath, file)

	if err != nil {
		if err.Error() == "faulty file" {
			fmt.Printf("Unsuccessful POST %v; faulty image %v\n", s.visionApiUrl, img.ID)
			s.updateClassificationById(img.ID, constants.CATEGORY_FAULTY, 0)
			return
		} else {
			panic(err)
		}
	}

	var category string
	if probability >= PepeThreshold {
		category = constants.CATEGORY_PEPE
	} else if probability >= MaybeThreshold {
		category = constants.CATEGORY_MAYBE
	} else {
		category = constants.CATEGORY_NON_PEPE
	}

	s.updateClassificationById(img.ID, category, probability)
	fmt.Printf("Successfully formatted %v; category: %v; probability: %v \n", img.ID, category, probability)
}

func (s *Image) updateClassificationById(id uint, category string, classification float32) {
	tx := s.db.CreateImageTransaction()
	defer tx.Deferral()
	err := tx.UpdateById(id, db.NewImage{Classification: classification, Category: category})
	utils.Check(err)
}

func (s *Image) storeImageResponse(r *imageResponse) *db.Image {
	tx := s.db.CreateImageTransaction()
	defer tx.Deferral()

	ext := getExtension(r.href)
	path := s.newPath(ext)

	i := tx.Create(db.NewImage{
		FilePath: path,
		Category: constants.CATEGORY_UNCLASSIFIED,
		Href:     r.href,
		Board:    "",
	})

	writeFile(path, *r.body)
	return i
}

func (s *Image) getImage(href string) (*imageResponse, error) {
	cleanedHref := fixMissingHttps(href)

	correctRequiredSubstrings := stringShouldContainOneFilter(cleanedHref, s.allowedImageTypes)
	if !correctRequiredSubstrings {
		return nil, errors.New("image type not allowed")
	}

	if s.doesImageExist(cleanedHref) {
		return nil, errors.New("image already exists")
	}

	request := Request{url: cleanedHref, reuseConnection: true, method: "GET"}
	response, _, success := request.Do(1)

	if !success {
		return nil, errors.New("unsuccessful response")
	}

	return &imageResponse{href: href, body: &response}, nil
}

func (s *Image) retrieveImageProbability(filePath string, file io.ReadCloser) (float32, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("An error has occurred while trying to classify an image with name: %v \n", filePath)
			panic(err)
		}
	}()

	b, w := createSingleFileMultiPart(VisionImageKey, filePath, file)
	ct := w.FormDataContentType()

	request := Request{url: s.visionApiUrl, reuseConnection: true, method: "POST", body: b, contentType: &ct}

	var do func(nRetry uint8) (float32, error)
	do = func(nRetry uint8) (float32, error) {
		response, statusCode, success := request.Do(nRetry)

		// assume that if 500 was returned; something is wrong with the file
		if statusCode == 500 {
			return 0, fmt.Errorf("faulty file")
		}

		if !success {
			// retry the request; not 500 or 200, most likely some temporary error
			return do(nRetry + 1)
		}

		data, err := ioutil.ReadAll(response)
		utils.Check(err)

		var vRes visionResponse
		err = json.Unmarshal(data, &vRes)
		utils.Check(err)

		return vRes.Score, nil
	}

	return do(1)
}

func (s *Image) doesImageExist(href string) bool {
	tx := s.db.CreateImageTransaction()
	defer tx.Deferral()
	return tx.ExistsByHref(href)
}

func (s *Image) newPath(extension string) (path string) {
	fileName := createUniqueId()
	fileName = addExtension(fileName, extension)
	path = filepath.Join(getProjectPath(), ImageDir, fileName)
	return
}
