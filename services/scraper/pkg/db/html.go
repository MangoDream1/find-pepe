package db

import (
	"errors"

	"gorm.io/gorm"
)

type NewHtml struct {
	FilePath string
	Href     string `gorm:"index"`
	Board    string `gorm:"index"`
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
	r := t.tx.Take(i, ID)
	err = r.Error
	return
}

func (t *htmlTx) ExistsByID(ID uint) bool {
	_, err := t.FindOneByID(ID)
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func (t *htmlTx) FindOneByHref(href string) (i *Html, err error) {
	i = &Html{}
	r := t.tx.Take(i, "href = ?", href)
	err = r.Error
	return
}

func (t *htmlTx) ExistsByHref(href string) bool {
	var result struct {
		Found bool
	}

	t.tx.Raw(`SELECT EXISTS(SELECT 1 FROM htmls WHERE "href" = ? AND "deleted_at" IS NULL) AS found`,
		href).Scan(&result)

	return result.Found
}

func (t *htmlTx) DeleteById(ID uint) (err error) {
	r := t.tx.Delete(&Html{gorm.Model{ID: ID}, NewHtml{}})
	err = r.Error
	return
}

func (t *htmlTx) FindAll(cb func(*Html)) (err error) {
	rows, err := t.tx.Model(&Html{}).Rows()
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

func (t *htmlTx) DeleteAll() {
	t.tx.Exec("DELETE FROM htmls")
}
