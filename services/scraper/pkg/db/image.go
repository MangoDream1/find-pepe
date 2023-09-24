package db

import (
	"errors"

	"gorm.io/gorm"
)

type NewImage struct {
	FilePath       string `gorm:"index"`
	Classification string `gorm:"index"`
	Href           string
	Board          string `gorm:"index"`
	Parsed         bool
}

type image struct {
	gorm.Model
	NewImage
}

type imgTx struct {
	tx       *gorm.DB
	Rollback func()
	Commit   func()
}

func (c *DbConnection) CreateImageTransaction() *imgTx {
	tx := c.db.Begin()

	return &imgTx{
		tx:       tx,
		Rollback: func() { tx.Rollback() },
		Commit:   func() { tx.Commit() },
	}
}

func (t *imgTx) Create(new NewImage) uint {
	img := &image{NewImage: new}
	t.tx.Create(&img)

	return img.ID
}

func (t *imgTx) FindOneByID(ID uint) (i *image, err error) {
	i = &image{}
	r := t.tx.First(i, ID)
	err = r.Error
	return
}

func (t *imgTx) ExistsByID(ID uint) bool {
	_, err := t.FindOneByID(ID)
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func (t *imgTx) FindOneByHref(href string) (i *image, err error) {
	i = &image{}
	r := t.tx.First(i, "href = ?", href)
	err = r.Error
	return
}

func (t *imgTx) ExistsByHref(href string) bool {
	_, err := t.FindOneByHref(href)
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func (t *imgTx) DeleteById(ID uint) (err error) {
	r := t.tx.Delete(&image{gorm.Model{ID: ID}, NewImage{}})
	err = r.Error
	return
}

func (t *imgTx) UpdateById(ID uint, update NewImage) (err error) {
	u := &image{gorm.Model{ID: ID}, update}
	r := t.tx.Save(u)
	err = r.Error
	return
}
