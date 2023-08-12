package main

import (
	"go-find-pepe/internal/scraper"
)

func main() {
	var allowedHrefSubstrings = []string{"4channel.org"}
	// var allowedHrefSubstrings = []string{"4chan.org", "4channel.org"}
	var requiredHrefSubstrings = []string{"https", "boards."}
	var allowedImageTypes = []string{".jpg", ".gif", ".png"}

	scraper := scraper.NewScraper(allowedHrefSubstrings, requiredHrefSubstrings, allowedImageTypes)

	scraper.Start("https://boards.4channel.org/g/")
}
