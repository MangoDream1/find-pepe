package db

import (
	"gorm.io/gorm"
)

type NewImage struct {
	FilePath       string
	Classification string `gorm:"index"`
	Href           string `gorm:"index"`
	Board          string `gorm:"index"`
}

type image struct {
	gorm.Model
	NewImage
}
