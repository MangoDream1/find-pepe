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

type Html struct {
	gorm.Model
	NewHtml
}

type htmlTx struct {
	tx       *gorm.DB
	Rollback func()
	Commit   func()
	Deferral func()
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
		Deferral: func() {
			if err := recover(); err != nil {
				tx.Rollback()
			} else {
				tx.Commit()
			}
		},
	}
}

func (t *htmlTx) Create(new NewHtml) *Html {
	h := &Html{NewHtml: new}
	t.tx.Create(&h)
	return h
}

func (t *htmlTx) FindOneByID(ID uint) (i *Html, err error) {
	i = &Html{}
	r := t.tx.First(i, ID)
	err = r.Error
	return
}

func (t *htmlTx) ExistsByID(ID uint) bool {
	_, err := t.FindOneByID(ID)
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func (t *htmlTx) FindOneByHref(href string) (i *Html, err error) {
	i = &Html{}
	r := t.tx.First(i, "href = ?", href)
	err = r.Error
	return
}

func (t *htmlTx) ExistsByHref(href string) bool {
	_, err := t.FindOneByHref(href)
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func (t *htmlTx) DeleteById(ID uint) (err error) {
	r := t.tx.Delete(&Html{gorm.Model{ID: ID}, NewHtml{}})
	err = r.Error
	return
}

func (t *htmlTx) UpdateById(ID uint, update NewHtml) (err error) {
	u := &Html{gorm.Model{ID: ID}, update}
	r := t.tx.Save(u)
	err = r.Error
	return
}

func (t *htmlTx) FindAll(cb func(*Html)) (err error) {
	rows, err := t.tx.Rows()
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var html Html
		t.tx.ScanRows(rows, &html)
		cb(&html)
	}

	return
}
