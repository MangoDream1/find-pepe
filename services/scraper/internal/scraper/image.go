package scraper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"go-find-pepe/internal/utils"
)

var ImageFileDirectories = [5]string{MaybeDir, PepeDir, NonPepeDir, UnclassifiedDir, FaultyDir}

type Image struct {
	allowedImageTypes []string
	visionApiUrl      string
	wg                *sync.WaitGroup
	done              *sync.Mutex
	imageHrefs        chan string
}

type imageResponse struct {
	fileName string
	body     *[]byte
}

type visionResponse struct {
	Score float32 `json:"score"`
}

func (s *Image) Start() {
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
		case href := <-s.imageHrefs:
			wgU.Wrapper(func() {
				defer s.wg.Done()
				request, err := s.getImage(href)
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

				s.storeImageRequest(request, toBeClassified)
			})
		}
	}
}

func (s *Image) classifyImageByPath(path string) {
	blob := readFile(path)
	probability, err := s.retrieveImageProbability(path, blob)

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

func (s *Image) storeImageRequest(request *imageResponse, output chan string) string {
	path := filepath.Join(getProjectPath(), UnclassifiedDir, request.fileName)

	writeFile(path, request.body)
	s.wg.Add(1)
	output <- path
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

	data, err := ioutil.ReadAll(response)
	utils.Check(err)

	return &imageResponse{fileName: fileName, body: &data}, nil
}

func (s *Image) retrieveImageProbability(filePath string, blob *[]byte) (float32, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("An error has occurred while trying to classify an image with name: %v \n", filePath)
			panic(err)
		}
	}()

	b, w := createSingleFileMultiPart(VisionImageKey, filePath, blob)
	ct := w.FormDataContentType()

	request := Request{url: s.visionApiUrl, reuseConnection: true, method: "POST", body: b, contentType: &ct}
	response, statusCode, success := request.Do(1)

	// assume that if 500 was returned; something is wrong with the file
	if statusCode == 500 {
		return 0, fmt.Errorf("faulty file")
	}

	if !success {
		return 0, fmt.Errorf("non-success response")
	}

	data, err := ioutil.ReadAll(response)
	utils.Check(err)

	var vRes visionResponse
	err = json.Unmarshal(data, &vRes)
	utils.Check(err)

	return vRes.Score, nil
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
