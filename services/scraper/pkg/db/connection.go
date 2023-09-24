package db

import (
	"fmt"
	"go-find-pepe/pkg/environment"
	"go-find-pepe/pkg/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DbConnection struct {
	db *gorm.DB
}

func Connect(env *environment.DbEnv) *DbConnection {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v", env.Host, env.User, env.Password, env.DbName, env.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	utils.Check(err)

	db.AutoMigrate(&image{})
	db.AutoMigrate(&html{})

	return &DbConnection{db: db}
}
