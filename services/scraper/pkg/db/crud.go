package db

import (
	"errors"
	"go-find-pepe/pkg/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type transaction struct {
	tx       *gorm.DB
	Rollback func()
	Commit   func()
}

type DbConnection struct {
	db *gorm.DB
}

func Connect() *DbConnection {
	dsn := "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	utils.Check(err)

	db.AutoMigrate(&image{})

	return &DbConnection{db: db}
}

func (c *DbConnection) CreateTransaction() *transaction {
	tx := c.db.Begin()

	return &transaction{
		tx:       tx,
		Rollback: func() { tx.Rollback() },
		Commit:   func() { tx.Commit() },
	}
}

func (t *transaction) Create(new NewImage) uint {
	img := &image{NewImage: new}
	t.tx.Create(&img)

	return img.ID
}

func (t *transaction) FindOneByID(ID uint) (i *image, err error) {
	i = &image{}
	r := t.tx.First(i, ID)
	err = r.Error
	return
}

func (t *transaction) ExistsByID(ID uint) bool {
	_, err := t.FindOneByID(ID)
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func (t *transaction) FindOneByHref(href string) (i *image, err error) {
	i = &image{}
	r := t.tx.First(i, "href = ?", href)
	err = r.Error
	return
}

func (t *transaction) ExistsByHref(href string) bool {
	_, err := t.FindOneByHref(href)
	return !errors.Is(err, gorm.ErrRecordNotFound)
}

func (t *transaction) DeleteById(ID uint) (err error) {
	r := t.tx.Delete(&image{gorm.Model{ID: ID}, NewImage{}})
	err = r.Error
	return
}

func (t *transaction) UpdateClassificationById(ID uint, classification string) (err error) {
	exists := t.ExistsByID(ID)
	if !exists {
		return gorm.ErrRecordNotFound
	}

	u := &image{gorm.Model{ID: ID}, NewImage{Classification: classification}}
	r := t.tx.Save(u)
	err = r.Error
	return
}
