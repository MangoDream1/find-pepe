package main

import (
	"go-find-pepe/pkg/db"
	"go-find-pepe/pkg/environment"
	"go-find-pepe/pkg/scraper"
	"go-find-pepe/pkg/utils"
)

func main() {
	scraperEnv, err := environment.ReadScraper()
	utils.Check(err)

	dbEnv, err := environment.ReadDb()
	utils.Check(err)

	var allowedHrefSubstrings = []string{"4channel.org"}
	// var allowedHrefSubstrings = []string{"4chan.org", "4channel.org"}
	var requiredHrefSubstrings = []string{"https", "boards."}
	var allowedImageTypes = []string{".jpg", ".gif", ".png"}

	scraper := scraper.NewScraper(scraper.NewScraperArguments{
		AllowedHrefSubstrings:  allowedHrefSubstrings,
		RequiredHrefSubstrings: requiredHrefSubstrings,
		AllowedImageTypes:      allowedImageTypes,
		ScraperEnv:             *scraperEnv,
		DbConnection:           db.Connect(dbEnv),
	})

	scraper.Start("https://boards.4channel.org/g/")
}
