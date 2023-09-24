package environment

type ScraperEnv struct {
	VisionApiUrl  string
	ImageLimit    int8
	HtmlLimit     int8
	ClassifyLimit int8
}

func ReadScraper() (*ScraperEnv, error) {
	visionApiUrl, err := readString("VISION_API_URL", "", true)
	if err != nil {
		return nil, err
	}

	hrefLimit, err := readInt("IMAGE_LIMIT", 0, false)
	if err != nil {
		return nil, err
	}

	classifyLimit, err := readInt("CLASSIFY_LIMIT", 0, false)
	if err != nil {
		return nil, err
	}

	htmlLimit, err := readInt("HTML_LIMIT", 0, false)
	if err != nil {
		return nil, err
	}

	return &ScraperEnv{
		ImageLimit:    int8(*hrefLimit),
		ClassifyLimit: int8(*classifyLimit),
		HtmlLimit:     int8(*htmlLimit),
		VisionApiUrl:  *visionApiUrl,
	}, nil
}
