package scraper

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"go-find-pepe/internal/utils"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const ErrorDirectory = "data/error"

func hash(s string) string {
	sha1Bytes := sha1.Sum([]byte(s))
	return hex.EncodeToString(sha1Bytes[:])
}

func writeFile(path string, b []byte) {
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

	_, err = f.Write(b)
	utils.Check(err)
	fmt.Printf("Successfully written file to %v\n", path)
}

func readFile(path string) []byte {
	buffer, err := ioutil.ReadFile(path)
	utils.Check(err)

	return buffer
}

func deleteFile(path string) {
	err := os.Remove(path)
	utils.Check(err)
	fmt.Printf("Successfully deleted file %v\n", path)
}

func doesFileExist(fileDir string, fileName string) bool {
	path := createPath(fileDir, fileName)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// readNestedDir finds all nested files within the original dirPath and puts the path into output
func readNestedDir(dirPath string, output chan string) {
	// dirPath does not exists, return early
	if _, err := os.Stat(dirPath); err != nil {
		if os.IsNotExist(err) {
			return
		}
	}

	fs, err := ioutil.ReadDir(dirPath)
	utils.Check(err)

	for _, f := range fs {
		path := createPath(dirPath, f.Name())
		if f.IsDir() {
			go readNestedDir(path, output)
		} else {
			output <- path
		}
	}
}

func createPath(fileDir string, fileName string) string {
	return filepath.Join(getProjectPath(), fileDir, fileName)
}

func replaceDir(path string, oldDir string, newDir string) string {
	oldPath := filepath.SplitList(path)
	newPath := make([]string, len(oldPath))

	for i, section := range oldPath {
		if section == oldDir {
			newPath[i] = newDir
			continue
		}

		newPath[i] = section
	}

	return filepath.Join(newPath...)
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

func cleanUpUrl(url string) string {
	if len(url) < 2 {
		return url
	}

	if url[0:2] == "//" {
		return fmt.Sprintf("https://%s", url[2:])
	}

	return url
}

func getURL(fileName string, url string) (response *http.Response, success bool, canRetry bool) {
	fmt.Printf("Fetching %v\n", url)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("User-Agent", "PostmanRuntime/7.29.3")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")
	req.Header.Add("Connection", "keep-alive")

	response, err := client.Do(req)

	canRetry = false
	success = false

	if err != nil {
		msg := err.Error()

		if stringShouldContainOneFilter(msg, []string{"timeout", "connection reset"}) {
			fmt.Printf("Failed to GET %v; timeout\n", url)
			canRetry = true
			return
		}

		fmt.Printf("Failed to GET %v; unknown error %v\n", url, msg)
		return
	}

	if response.StatusCode == 503 {
		fmt.Printf("Failed to GET %v; 503 response\n", url)
		canRetry = true
		return
	}

	if response.StatusCode == 404 {
		fmt.Printf("Failed to GET %v; 404 response\n", url)
		return
	}

	if response.StatusCode != 200 {
		data, err := ioutil.ReadAll(response.Body)
		utils.Check(err)

		path := createPath(ErrorDirectory, fmt.Sprintf("%v/%v%v%v", response.StatusCode, url, time.Now().UTC(), ".html"))
		writeFile(path, data)
		panic(fmt.Sprintf("Failed to GET %v; non-OK response: %v", url, response.StatusCode))
	}

	fmt.Printf("Successfully fetched %v \n", url)
	success = true
	return
}

func createSingleFileMultiPart(key string, fileName string, file []byte) (*bytes.Buffer, *multipart.Writer) {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	part, err := writer.CreateFormFile(key, fileName)
	utils.Check(err)

	r := bytes.NewReader(file)
	_, err = io.Copy(part, r)
	utils.Check(err)

	err = writer.Close()
	utils.Check(err)

	return &b, writer
}
