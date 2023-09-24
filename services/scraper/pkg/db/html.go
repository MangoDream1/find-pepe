package db

import (
	"errors"

	"gorm.io/gorm"
)

type NewHtml struct {
	FilePath string `gorm:"index"`
	Href     string
	Board    string `gorm:"index"`
}

type html struct {
	gorm.Model
	NewImage
}

type htmlTx struct {
	tx       *gorm.DB
	Rollback func()
	Commit   func()
}

func (c *DbConnection) CreateHtmlTransaction() *htmlTx {
	tx := c.db.Begin()

	return &htmlTx{
		tx:       tx,
		Rollback: func() { tx.Rollback() },
		Commit:   func() { tx.Commit() },
	}
}

func (t *htmlTx) Create(new NewImage) uint {
	img := &html{NewImage: new}
	t.tx.Create(&img)

	return img.ID
}

func (t *htmlTx) FindOneByID(ID uint) (i *html, err error) {
	i = &html{}
	r := t.tx.First(i, ID)
	err = r.Error
	return
}

func (t *htmlTx) ExistsByID(ID uint) bool {
	_, err := t.FindOneByID(ID)
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func (t *htmlTx) FindOneByHref(href string) (i *html, err error) {
	i = &html{}
	r := t.tx.First(i, "href = ?", href)
	err = r.Error
	return
}

func (t *htmlTx) ExistsByHref(href string) bool {
	_, err := t.FindOneByHref(href)
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func (t *htmlTx) DeleteById(ID uint) (err error) {
	r := t.tx.Delete(&html{gorm.Model{ID: ID}, NewImage{}})
	err = r.Error
	return
}

func (t *htmlTx) UpdateClassificationById(ID uint, classification string) (err error) {
	u := &html{gorm.Model{ID: ID}, NewImage{Classification: classification}}
	r := t.tx.Save(u)
	err = r.Error
	return
}
