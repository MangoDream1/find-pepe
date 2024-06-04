package scraper

import (
	"bytes"
	"fmt"
	"go-find-pepe/pkg/utils"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

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
		time.Sleep(time.Second * time.Duration(backoff))
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

	success = false
	if err != nil {
		msg := err.Error()

		if stringShouldContainOneFilter(msg, []string{"timeout", "connection reset"}) {
			fmt.Printf("Failed to %v %v; timeout\n", r.method, r.url)
			return retry()
		}

		if stringShouldContainOneFilter(msg, []string{"connection refused"}) {
			fmt.Printf("Failed to %v %v; connection refused\n", r.method, r.url)
			return retry()
		}

		if stringShouldContainOneFilter(msg, []string{"EOF"}) {
			fmt.Printf("Failed to %v %v; EOF\n", r.method, r.url)
			return retry()
		}

		fmt.Printf("Failed to %v %v; unknown error %v\n", r.method, r.url, msg)
		return
	}

	statusCode = response.StatusCode
	reader = response.Body

	if response.StatusCode == 503 {
		fmt.Printf("Failed to %v %v; 503 response\n", r.method, r.url)
		return retry()
	}

	if response.StatusCode == 404 {
		fmt.Printf("Failed to %v %v; 404 response\n", r.method, r.url)
		return
	}

	if response.StatusCode != 200 {
		path := filepath.Join(getProjectPath(), ErrorDirectory, fmt.Sprintf("%v/%v%v%v", response.StatusCode, r.url, time.Now().UTC(), ".html"))
		writeFile(path, reader)
		return
	}

	fmt.Printf("Successfully fetched %v %v \n", r.method, r.url)
	success = true
	return
}

func createSingleFileMultiPart(key string, fileName string, file io.ReadCloser) (*bytes.Buffer, *multipart.Writer) {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	part, err := writer.CreateFormFile(key, fileName)
	utils.Check(err)

	_, err = io.Copy(part, file)
	utils.Check(err)

	err = writer.Close()
	utils.Check(err)

	return &b, writer
}

func calculateExponentialBackoffInSec(a uint8) float64 {
	return math.Pow(2, float64(a))
}
