package environment

import (
	"os"
	"strconv"
)

type Environment struct {
	VisionApiUrl  string
	ImageLimit    int8
	HtmlLimit     int8
	ClassifyLimit int8
}

func readIntFromEnv(env string) (*int64, error) {
	iString := os.Getenv(env)
	var i int64
	i = 0

	if iString != "" {
		var err error
		i, err = strconv.ParseInt(iString, 10, 8)
		if err != nil {
			return nil, err
		}
	}

	return &i, nil
}

func ReadEnvironment() (*Environment, error) {
	visionApiUrl := os.Getenv("VISION_API_URL")
	if visionApiUrl == "" {
		panic("VISION_API_URL unset")
	}

	hrefLimit, err := readIntFromEnv("IMAGE_LIMIT")
	if err != nil {
		return nil, err
	}

	classifyLimit, err := readIntFromEnv("CLASSIFY_LIMIT")
	if err != nil {
		return nil, err
	}

	htmlLimit, err := readIntFromEnv("HTML_LIMIT")
	if err != nil {
		return nil, err
	}

	return &Environment{
		ImageLimit:    int8(*hrefLimit),
		ClassifyLimit: int8(*classifyLimit),
		HtmlLimit:     int8(*htmlLimit),
		VisionApiUrl:  visionApiUrl,
	}, nil
}
