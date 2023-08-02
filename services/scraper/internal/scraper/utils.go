package scraper

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"go-find-pepe/internal/utils"
	"io"
	"io/fs"
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

func writeFile(fileDir string, fileName string, b []byte) {
	path := createPath(fileDir, fileName)
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

func readFile(fileDir string, fileName string) []byte {
	path := createPath(fileDir, fileName)

	buffer, err := ioutil.ReadFile(path)
	utils.Check(err)

	return buffer
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

func createReader(fileDir string, fileName string) *io.Reader {
	path := createPath(fileDir, fileName)

	f, err := os.Open(path)
	utils.Check(err)

	var r io.Reader
	r = f

	return &r
}

func readDir(dir string) []fs.FileInfo {
	dir = filepath.Join(getProjectPath(), dir)
	err := os.MkdirAll(dir, os.ModePerm)
	utils.Check(err)

	fileInfos, err := ioutil.ReadDir(dir)
	utils.Check(err)
	return fileInfos
}

func createPath(fileDir string, fileName string) string {
	return filepath.Join(getProjectPath(), fileDir, fileName)
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

	response, err := http.Get(url)
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

		writeFile(ErrorDirectory, fmt.Sprintf("%v/%v%v%v", response.StatusCode, url, time.Now().UTC(), ".html"), data)
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
