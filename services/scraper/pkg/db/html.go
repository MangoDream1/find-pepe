package db

import (
	"errors"

	"gorm.io/gorm"
)

type NewHtml struct {
	FilePath string `gorm:"index"`
	Href     string
	Board    string `gorm:"index"`
	Parsed   bool
}

type html struct {
	gorm.Model
	NewHtml
}

type htmlTx struct {
	tx       *gorm.DB
	Rollback func()
	Commit   func()
}

type HtmlDbConnection struct {
	db *gorm.DB
}

func (c *DbConnection) InitHtml() *HtmlDbConnection {
	return &HtmlDbConnection{db: c.db}
}

func (c *HtmlDbConnection) CreateTransaction() *htmlTx {
	tx := c.db.Begin()

	return &htmlTx{
		tx:       tx,
		Rollback: func() { tx.Rollback() },
		Commit:   func() { tx.Commit() },
	}
}

func (t *htmlTx) Create(new NewHtml) uint {
	img := &html{NewHtml: new}
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
	r := t.tx.Delete(&html{gorm.Model{ID: ID}, NewHtml{}})
	err = r.Error
	return
}

func (t *htmlTx) UpdateById(ID uint, update NewHtml) (err error) {
	u := &html{gorm.Model{ID: ID}, update}
	r := t.tx.Save(u)
	err = r.Error
	return
}
