package scraper

import (
	"bytes"
	"fmt"
	"go-find-pepe/internal/utils"
	"io"
	"io/ioutil"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func writeFile(path string, b *[]byte) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("An error has occurred while trying to store an file with name: %v \n", path)
			panic(err)
		}
	}()

	fmt.Printf("Writing file to %v\n", path)

	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	utils.Check(err)

	f, err := os.Create(path)
	utils.Check(err)
	defer f.Close()

	_, err = f.Write(*b)
	utils.Check(err)
	fmt.Printf("Successfully written file to %v\n", path)
}

func readFile(path string) *[]byte {
	buffer, err := ioutil.ReadFile(path)
	utils.Check(err)

	return &buffer
}

func deleteFile(path string) {
	if !doesFileExist(path) {
		return
	}

	err := os.Remove(path)
	utils.Check(err)
	fmt.Printf("Successfully deleted file %v\n", path)
}

func moveFile(oldPath string, newPath string) {
	blob := readFile(oldPath)
	writeFile(newPath, blob)
	deleteFile(oldPath)
}

func doesFileExist(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// readNestedDir finds all nested files within the original dirPath and puts the path into output
func readNestedDir(dirPath string, output func(string)) {
	if !doesFileExist(dirPath) {
		fmt.Printf("Stopped reading nested directory; %v does not exists\n", dirPath)
		return
	}

	var wg sync.WaitGroup

	var inner func(dirPath string)
	inner = func(dirPath string) {
		defer wg.Done()

		fs, err := os.ReadDir(dirPath)
		utils.Check(err)

		for _, f := range fs {
			path := filepath.Join(dirPath, f.Name())
			if f.IsDir() {
				wg.Add(1)
				go inner(path)
			} else {
				output(path)
			}
		}
	}

	wg.Add(1)
	go inner(dirPath)
	wg.Wait()
}

func replaceDir(oldPath string, oldDir string, newDir string) string {
	newPath := strings.Replace(oldPath, oldDir, newDir, 1)

	if newPath == oldPath {
		panic(fmt.Errorf("failed to replaceDir; oldDir %v not in oldPath %v", newDir, oldPath))
	}

	return newPath
}

func getProjectPath() string {
	projectPath, err := os.Getwd()
	utils.Check(err)
	return projectPath
}

func addExtension(id string, extension string) string {
	return fmt.Sprintf("%s.%s", id, extension)
}

func removeExtension(filename string) string {
	extension := filepath.Ext(filename)
	n := strings.LastIndex(filename, extension)
	return filename[:n]
}

func stringShouldContainOneFilter(s string, filters []string) bool {
	for _, filter := range filters {
		if strings.Contains(s, filter) {
			return true
		}
	}
	return false
}

func stringShouldContainAllFilters(s string, filters []string) bool {
	count := 0
	_filters := filters
	for _, filter := range _filters {
		if strings.Contains(s, filter) {
			count++
		}
	}
	return len(_filters) == count
}

func getHostname(rawurl string) string {
	parsed, err := url.Parse(rawurl)
	utils.Check(err)
	return parsed.Hostname()
}

func fixMissingHttps(url string) string {
	if len(url) < 2 {
		return url
	}

	if url[0:2] == "//" {
		return fmt.Sprintf("https://%s", url[2:])
	}

	return url
}

type Request struct {
	url             string
	reuseConnection bool
	method          string
	body            io.Reader
	contentType     *string
}

func (r *Request) Do(nAttempt uint8) (reader io.ReadCloser, statusCode int, success bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("An error has occurred while trying to retrieve href: %v %v\n", r.method, r.url)
			panic(err)
		}
	}()

	if nAttempt >= MAX_RETRY_ATTEMPT {
		panic(fmt.Errorf("failed to %v %v after MAX_ATTEMPT=%v", r.method, r.url, MAX_RETRY_ATTEMPT))
	}

	retry := func() (response io.ReadCloser, statusCode int, success bool) {
		backoff := calculateExponentialBackoffInSec(nAttempt)
		fmt.Printf("Retrying %v %v after %v\n", r.method, r.url, backoff)
		time.Sleep(time.Second * time.Duration(nAttempt))
		return r.Do(nAttempt + 1)
	}

	fmt.Printf("Fetching %v %v\n", r.method, r.url)

	client := &http.Client{}
	req, _ := http.NewRequest(r.method, r.url, r.body)

	req.Header.Add("User-Agent", "PostmanRuntime/7.29.3")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	if r.contentType != nil {
		req.Header.Set("Content-Type", *r.contentType)
	}

	if r.reuseConnection {
		req.Header.Add("Connection", "keep-alive")
	} else {
		req.Close = true
	}

	response, err := client.Do(req)
	statusCode = response.StatusCode

	success = false
	if err != nil {
		msg := err.Error()

		if stringShouldContainOneFilter(msg, []string{"timeout", "connection reset"}) {
			fmt.Printf("Failed to %v %v; timeout\n", r.method, r.url)
			return retry()
		}

		if stringShouldContainOneFilter(msg, []string{"EOF"}) {
			fmt.Printf("Failed to %v %v; EOF\n", r.method, r.url)
			return retry()
		}

		fmt.Printf("Failed to %v %v; unknown error %v\n", r.method, r.url, msg)
		return
	}

	if response.StatusCode == 503 {
		fmt.Printf("Failed to %v %v; 503 response\n", r.method, r.url)
		return retry()
	}

	if response.StatusCode == 404 {
		fmt.Printf("Failed to %v %v; 404 response\n", r.method, r.url)
		return
	}

	if response.StatusCode != 200 {
		data, err := ioutil.ReadAll(response.Body)
		utils.Check(err)

		path := filepath.Join(getProjectPath(), ErrorDirectory, fmt.Sprintf("%v/%v%v%v", response.StatusCode, r.url, time.Now().UTC(), ".html"))
		writeFile(path, &data)
		panic(fmt.Sprintf("Failed to %v %v; non-OK response: %v", r.method, r.url, response.StatusCode))
	}

	fmt.Printf("Successfully fetched %v %v \n", r.method, r.url)
	reader = response.Body
	success = true
	return
}

func createSingleFileMultiPart(key string, fileName string, file *[]byte) (*bytes.Buffer, *multipart.Writer) {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	part, err := writer.CreateFormFile(key, fileName)
	utils.Check(err)

	r := bytes.NewReader(*file)
	_, err = io.Copy(part, r)
	utils.Check(err)

	err = writer.Close()
	utils.Check(err)

	return &b, writer
}

func calculateExponentialBackoffInSec(a uint8) float64 {
	return math.Pow(2, float64(a))
}
