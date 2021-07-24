package main

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

func hash(s string) string {
	sha1Bytes := sha1.Sum([]byte(s))
	return hex.EncodeToString(sha1Bytes[:])
}

func writeFile(fileDir string, id string, extension string, b []byte) {
	path := createPath(fileDir, id, extension)

	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	check(err)

	f, err := os.Create(path)
	check(err)
	defer f.Close()

	_, err = f.Write(b)
	check(err)
}

func readFile(fileDir string, id string, extension string) []byte {
	path := createPath(fileDir, id, extension)

	buffer, err := ioutil.ReadFile(path)
	check(err)

	return buffer
}

func doesFileExist(fileDir string, id string, extension string) bool {
	path := createPath(fileDir, id, extension)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func createReader(fileDir string, id string, extension string) *io.Reader {
	path := createPath(fileDir, id, extension)

	f, err := os.Open(path)
	check(err)

	var r io.Reader
	r = f

	return &r
}

func readDir(dir string) []fs.FileInfo {
	dir = filepath.Join(getProjectPath(), dir)
	err := os.MkdirAll(dir, os.ModePerm)
	check(err)

	fileInfos, err := ioutil.ReadDir(dir)
	check(err)
	return fileInfos
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func createPath(fileDir string, id string, extension string) string {
	return filepath.Join(getProjectPath(), fileDir, addExtension(id, extension))
}

func getProjectPath() string {
	projectPath, err := os.Getwd()
	check(err)
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
	for _, filter := range filters {
		if strings.Contains(s, filter) {
			count++
		}
	}
	return len(filters) == count
}

func getHostname(rawurl string) string {
	parsed, err := url.Parse(rawurl)
	check(err)
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
