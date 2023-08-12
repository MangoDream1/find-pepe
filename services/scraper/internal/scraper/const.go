package scraper

const ErrorDirectory = "data/error"

const HtmlDir = "data/html"

const ImageDir = "data/image"
const MaybeDir = ImageDir + "/maybe"
const PepeDir = ImageDir + "/pepe"
const NonPepeDir = ImageDir + "/non-pepe"
const UnclassifiedDir = ImageDir + "/unclassified"
const FaultyDir = ImageDir + "/faulty"

const PepeThreshold = 0.9
const MaybeThreshold = 0.3

const VisionImageKey = "file"

const MAX_RETRY_ATTEMPT = 5
