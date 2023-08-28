package main

import (
	"go-find-pepe/internal/environment"
	"go-find-pepe/internal/scraper"
	"go-find-pepe/internal/utils"
)

func main() {
	env, err := environment.ReadEnvironment()
	utils.Check(err)

	var allowedHrefSubstrings = []string{"4channel.org"}
	// var allowedHrefSubstrings = []string{"4chan.org", "4channel.org"}
	var requiredHrefSubstrings = []string{"https", "boards."}
	var allowedImageTypes = []string{".jpg", ".gif", ".png"}

	newScraperArgs := &scraper.NewScraperArguments{
		AllowedHrefSubstrings:  allowedHrefSubstrings,
		RequiredHrefSubstrings: requiredHrefSubstrings,
		AllowedImageTypes:      allowedImageTypes,
		Environment:            *env,
	}

	scraper := scraper.NewScraper(newScraperArgs)

	scraper.Start("https://boards.4channel.org/g/")
}
