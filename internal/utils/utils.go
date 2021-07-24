package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func Hash(s string) string {
	sha1Bytes := sha1.Sum([]byte(s))
	return hex.EncodeToString(sha1Bytes[:])
}

func WriteFile(fileDir string, id string, extension string, b []byte) {
	path := createPath(fileDir, id, extension)

	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	Check(err)

	f, err := os.Create(path)
	Check(err)
	defer f.Close()

	_, err = f.Write(b)
	Check(err)
}

func ReadFile(fileDir string, id string, extension string) []byte {
	path := createPath(fileDir, id, extension)

	buffer, err := ioutil.ReadFile(path)
	Check(err)

	return buffer
}

func DoesFileExist(fileDir string, id string, extension string) bool {
	path := createPath(fileDir, id, extension)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func CreateReader(fileDir string, id string, extension string) *io.Reader {
	path := createPath(fileDir, id, extension)

	f, err := os.Open(path)
	Check(err)

	var r io.Reader
	r = f

	return &r
}

func ReadDir(dir string) []fs.FileInfo {
	dir = filepath.Join(getProjectPath(), dir)
	err := os.MkdirAll(dir, os.ModePerm)
	Check(err)

	fileInfos, err := ioutil.ReadDir(dir)
	Check(err)
	return fileInfos
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func createPath(fileDir string, id string, extension string) string {
	return filepath.Join(getProjectPath(), fileDir, addExtension(id, extension))
}

func getProjectPath() string {
	projectPath, err := os.Getwd()
	Check(err)
	return projectPath
}

func addExtension(id string, extension string) string {
	return fmt.Sprintf("%s.%s", id, extension)
}

func RemoveExtension(filename string) string {
	extension := filepath.Ext(filename)
	n := strings.LastIndex(filename, extension)
	return filename[:n]
}

func StringShouldContainOneFilter(s string, filters []string) bool {
	for _, filter := range filters {
		if strings.Contains(s, filter) {
			return true
		}
	}
	return false
}

func StringShouldContainAllFilters(s string, filters []string) bool {
	count := 0
	for _, filter := range filters {
		if strings.Contains(s, filter) {
			count++
		}
	}
	return len(filters) == count
}

func GetHostname(rawurl string) string {
	parsed, err := url.Parse(rawurl)
	Check(err)
	return parsed.Hostname()
}

func CleanUpUrl(url string) string {
	if len(url) < 2 {
		return url
	}

	if url[0:2] == "//" {
		return fmt.Sprintf("https://%s", url[2:])
	}

	return url
}
