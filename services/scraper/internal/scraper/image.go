package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-find-pepe/internal/utils"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
)

var ImageFileDirectories = [5]string{MaybeDir, PepeDir, NonPepeDir, UnclassifiedDir, FaultyDir}

type Image struct {
	allowedImageTypes []string
	visionApiUrl      string
	wg                *sync.WaitGroup
	done              *sync.Mutex
	imageHrefs        chan string
	hrefLimit         uint8
	classifyLimit     uint8
}

type imageResponse struct {
	fileName string
	body     *io.ReadCloser
}

type visionResponse struct {
	Score float32 `json:"score"`
}

func (s *Image) Start() {
	toBeClassified := make(chan string)

	done := make(chan bool)
	wgU := WaitGroupUtil{WaitGroup: s.wg}

	wgU.Wrapper(func() {
		dirPath := filepath.Join(getProjectPath(), UnclassifiedDir)
		readNestedDir(dirPath, func(path string) {
			s.wg.Add(1)
			toBeClassified <- path
		})
	})

	hrefLimiter := NewLimiter(s.hrefLimit)
	classifyLimiter := NewLimiter(s.classifyLimit)

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
		case path := <-toBeClassified:
			wgU.Wrapper(func() {
				classifyLimiter.Add()
				defer s.wg.Done()
				defer classifyLimiter.Done()
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

func (s *Image) classifyImageByPath(path string) {
	file := readFile(path)
	defer file.Close()

	probability, err := s.retrieveImageProbability(path, file)

	if err != nil {
		if err.Error() == "faulty file" {
			fmt.Printf("Unsuccessful POST %v; moved image %v\n", s.visionApiUrl, path)
			newPath := replaceDir(path, UnclassifiedDir, FaultyDir)
			moveFile(path, newPath)
			return
		}

		panic(err)
	}

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

func (s *Image) storeImageResponse(request *imageResponse) string {
	path := filepath.Join(getProjectPath(), UnclassifiedDir, request.fileName)

	writeFile(path, *request.body)
	return path
}

func (s *Image) getImage(href string) (*imageResponse, error) {
	cleanedHref := fixMissingHttps(href)

	correctRequiredSubstrings := stringShouldContainOneFilter(cleanedHref, s.allowedImageTypes)
	if !correctRequiredSubstrings {
		return nil, errors.New("image type not allowed")
	}

	fileName := s.transformUrlIntoFilename(cleanedHref)
	if s.doesImageExist(fileName) {
		return nil, errors.New("image already exists")
	}

	request := Request{url: cleanedHref, reuseConnection: true, method: "GET"}
	response, _, success := request.Do(1)

	if !success {
		return nil, errors.New("unsuccessful response")
	}

	return &imageResponse{fileName: fileName, body: &response}, nil
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

func (s *Image) doesImageExist(fileName string) bool {
	for _, dir := range ImageFileDirectories {
		path := filepath.Join(getProjectPath(), dir, fileName)
		if !doesFileExist(path) {
			return false
		}
	}

	return true
}

func (s *Image) transformUrlIntoFilename(url string) string {
	p := strings.Split(url, `/`)
	return strings.Join(p[2:], `/`)
}
