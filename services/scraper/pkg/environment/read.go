package environment

import (
	"os"
	"strconv"
)

type Environment struct {
	VisionApiUrl  string
	HrefLimit     int8
	ClassifyLimit int8
}

func ReadEnvironment() (*Environment, error) {
	visionApiUrl := os.Getenv("VISION_API_URL")
	if visionApiUrl == "" {
		panic("VISION_API_URL unset")
	}

	hrefLimiterString := os.Getenv("HREF_LIMIT")
	var hrefLimit int64
	hrefLimit = 0

	if hrefLimiterString != "" {
		var err error
		hrefLimit, err = strconv.ParseInt(hrefLimiterString, 10, 8)
		if err != nil {
			return nil, err
		}
	}

	classifyLimiterString := os.Getenv("CLASSIFY_LIMIT")
	var classifyLimit int64
	classifyLimit = 0

	if classifyLimiterString != "" {
		var err error
		classifyLimit, err = strconv.ParseInt(classifyLimiterString, 10, 8)
		if err != nil {
			return nil, err
		}
	}

	return &Environment{
		HrefLimit:     int8(hrefLimit),
		ClassifyLimit: int8(classifyLimit),
		VisionApiUrl:  visionApiUrl,
	}, nil
}
