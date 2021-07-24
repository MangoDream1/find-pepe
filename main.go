package main

import "go-find-pepe/internal/scraper"

func main() {
	var allowedHrefSubstrings = []string{"4chan.org", "4channel.org"}
	var requiredHrefSubstrings = []string{"https", "boards."}

	s := scraper.NewScraper(allowedHrefSubstrings, requiredHrefSubstrings)

	go s.ReadDownloadedIds()
	go s.Start("https://www.4chan.org/")

}
