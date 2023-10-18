package scraper

import "strings"

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
	_filters := filters
	for _, filter := range _filters {
		if strings.Contains(s, filter) {
			count++
		}
	}
	return len(_filters) == count
}
