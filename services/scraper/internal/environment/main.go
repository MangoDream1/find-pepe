package environment

import (
	"os"
	"strconv"
)

type Environment struct {
	VisionApiUrl  string
	HrefLimit     uint8
	ClassifyLimit uint8
}

func ReadEnvironment() (*Environment, error) {
	visionApiUrl := os.Getenv("VISION_API_URL")
	if visionApiUrl == "" {
		panic("VISION_API_URL unset")
	}

	hrefLimiterString := os.Getenv("HREF_LIMIT")
	var hrefLimit uint64
	hrefLimit = 0

	if hrefLimiterString != "" {
		var err error
		hrefLimit, err = strconv.ParseUint(hrefLimiterString, 10, 8)
		if err != nil {
			return nil, err
		}
	}

	classifyLimiterString := os.Getenv("CLASSIFY_LIMIT")
	var classifyLimit uint64
	classifyLimit = 0

	if classifyLimiterString != "" {
		var err error
		classifyLimit, err = strconv.ParseUint(classifyLimiterString, 10, 8)
		if err != nil {
			return nil, err
		}
	}

	return &Environment{
		HrefLimit:     uint8(hrefLimit),
		ClassifyLimit: uint8(classifyLimit),
		VisionApiUrl:  visionApiUrl,
	}, nil
}
