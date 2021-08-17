package lib

import (
	"log"
	"math/rand"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Movie struct {
	ID int
	Title string
	Year int
	Genre string
	Actors string
	Rating int
	Comment string
}

func(obj *Movie) GetRandom()  int{
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
		return 404
	}
	rand.Seed(time.Now().UnixNano())
	db.Where("id = ?", rand.Intn(obj.Count() - 1)).First(&obj)
	return 0
}

func (obj *Movie) Count() int {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
		return 404
	}
	var movies []Movie
	var count int64
	db.Find(&movies).Count(&count)
	return int(count)
}

func (obj *Movie) GetById(id int) int {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
		return 404
	}
	db.Where("id = ?",id).First(&obj)
	return 0
}

func (obj *Movie) Add() int{
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database")
		return 404
	}
	if db.Create(&obj).RowsAffected != 1 {
		log.Println("Не удалось добавить фильм")
	}
	return 0
}
