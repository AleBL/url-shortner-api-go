package database

import (
	"log"
	"sync"
	"url-shortener/models"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

var db *gorm.DB
var once sync.Once

const dbFile = "./internal//database/database.db"

func ConnectDB() *gorm.DB {
	once.Do(func() {
		var err error
		db, err = gorm.Open("sqlite3", dbFile)
		if err != nil {
			log.Fatal("Error opening database:", err)
		}

		db.AutoMigrate(&models.URL{})
	})
	return db
}
