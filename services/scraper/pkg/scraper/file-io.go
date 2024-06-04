package scraper

import (
	"fmt"
	"go-find-pepe/pkg/utils"
	"io"
	"os"
	"path/filepath"
)

func writeFile(path string, file io.ReadCloser) {
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

	_, err = io.Copy(f, file)
	utils.Check(err)

	fmt.Printf("Successfully written file to %v\n", path)
}

func readFile(path string) io.ReadCloser {
	file, err := os.Open(path)
	utils.Check(err)

	return file
}

func getProjectPath() string {
	projectPath, err := os.Getwd()
	utils.Check(err)
	return projectPath
}

func addExtension(id string, extension string) string {
	return fmt.Sprintf("%s.%s", id, extension)
}

func getExtension(filename string) (extension string) {
	extension = filepath.Ext(filename)[1:]
	return
}
