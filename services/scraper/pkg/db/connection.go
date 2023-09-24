package db

import (
	"go-find-pepe/pkg/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DbConnection struct {
	db *gorm.DB
}

func Connect() *DbConnection {
	dsn := "host=localhost user=postgres password=mysecretpassword dbname=postgres port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	utils.Check(err)

	db.AutoMigrate(&image{})
	db.AutoMigrate(&html{})

	return &DbConnection{db: db}
}
