package main

import "go-find-pepe/internal/scraper"

func main() {
	var allowedHrefSubstrings = []string{"4chan.org", "4channel.org"}
	var requiredHrefSubstrings = []string{"https", "boards."}

	scraper := scraper.NewScraper(allowedHrefSubstrings, requiredHrefSubstrings)
	go scraper.ReadDownloadedIds()

	scraper.Start("https://www.4chan.org/")

}
