package db

import (
	"errors"
	"go-find-pepe/pkg/constants"

	"gorm.io/gorm"
)

type NewImage struct {
	FilePath       string
	Category       string `gorm:"index"`
	Classification float32
	Href           string `gorm:"index"`
	Board          string `gorm:"index"`
}

type Image struct {
	gorm.Model
	NewImage
}

type imgTx struct {
	tx       *gorm.DB
	Rollback func()
	Commit   func()
	Deferral func()
}

type ImageDbConnection struct {
	db *gorm.DB
}

func (c *DbConnection) InitImage() *ImageDbConnection {
	return &ImageDbConnection{db: c.db}
}

func (c *ImageDbConnection) CreateImageTransaction() *imgTx {
	tx := c.db.Begin()

	return &imgTx{
		tx:       tx,
		Rollback: func() { tx.Rollback() },
		Commit:   func() { tx.Commit() },
		Deferral: func() {
			if err := recover(); err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		},
	}
}

func (t *imgTx) Create(new NewImage) *Image {
	img := &Image{NewImage: new}
	t.tx.Create(&img)
	return img
}

func (t *imgTx) FindOneByID(ID uint) (i *Image, err error) {
	i = &Image{}
	r := t.tx.Take(i, ID)
	err = r.Error
	return
}

func (t *imgTx) ExistsByID(ID uint) bool {
	_, err := t.FindOneByID(ID)
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func (t *imgTx) FindOneByHref(href string) (i *Image, err error) {
	i = &Image{}
	r := t.tx.Take(i, "href = ?", href)
	err = r.Error
	return
}

func (t *imgTx) ExistsByHref(href string) bool {
	var result struct {
		Found bool
	}

	t.tx.Raw(`SELECT EXISTS(SELECT 1 FROM images WHERE "href" = ? AND "deleted_at" IS NULL) AS found`,
		href).Scan(&result)

	return result.Found
}

func (t *imgTx) DeleteById(ID uint) (err error) {
	r := t.tx.Delete(&Image{gorm.Model{ID: ID}, NewImage{}})
	err = r.Error
	return
}

func (t *imgTx) UpdateById(ID uint, update NewImage) (err error) {
	r := t.tx.Model(&Image{}).Where(&Image{
		Model:    gorm.Model{ID: ID},
		NewImage: NewImage{},
	}).Updates(&Image{
		Model:    gorm.Model{},
		NewImage: update,
	})
	err = r.Error
	return
}

func (t *imgTx) FindAllUnclassified(cb func(*Image)) (err error) {
	rows, err := t.tx.Model(&Image{}).Where("category = ?", constants.CATEGORY_UNCLASSIFIED).Rows()
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var img Image
		t.tx.ScanRows(rows, &img)
		cb(&img)
	}

	return
}
