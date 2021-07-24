package scraper

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"go-find-pepe/internal/utils"
	"io"
	"io/fs"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func hash(s string) string {
	sha1Bytes := sha1.Sum([]byte(s))
	return hex.EncodeToString(sha1Bytes[:])
}

func writeFile(fileDir string, fileName string, b []byte) {
	path := createPath(fileDir, fileName)

	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	utils.Check(err)

	f, err := os.Create(path)
	utils.Check(err)
	defer f.Close()

	_, err = f.Write(b)
	utils.Check(err)
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
