package scraper

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

func writeToPanicFile() {
	if err := recover(); err != nil {

		switch x := err.(type) {
		case string:
			err = x
		case error:
			err = x.Error()
		default:
			// Fallback err (per specs, error strings should be lowercase w/o punctuation
			err = "unknown panic"
		}

		path := filepath.Join(getProjectPath(), ErrorDirectory, "panic", time.Now().UTC().String()+".txt")

		stack := string(debug.Stack()[:])
		file := io.NopCloser(strings.NewReader(fmt.Sprintf("%v\n%v", stack, err)))

		writeFile(path, file)

		panic(err)
	}
}
