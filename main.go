package main

func main() {
	var allowedHrefSubstrings = []string{"4chan.org", "4channel.org"}
	var requiredHrefSubstrings = []string{"https", "boards."}

	s := NewScraper(allowedHrefSubstrings, requiredHrefSubstrings)

	go s.ReadDownloadedIds()
	s.Start("https://www.4chan.org/")

}
