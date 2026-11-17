package database

import (
	"log"
	"os"
	"path/filepath"
	"sync"
	"url-shortener/models"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

var db *gorm.DB
var once sync.Once

const defaultDBPath = "./internal/database/database.db"

// Path returns the SQLite file location, configurable through DB_PATH so the
// database can live on a persistent volume in production.
func Path() string {
	if path := os.Getenv("DB_PATH"); path != "" {
		return path
	}
	return defaultDBPath
}

func ConnectDB() *gorm.DB {
	once.Do(func() {
		path := Path()
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			log.Fatal("Error creating database directory:", err)
		}

		var err error
		// busy_timeout makes concurrent writers wait instead of failing with "database is locked".
		db, err = gorm.Open("sqlite3", path+"?_busy_timeout=5000")
		if err != nil {
			log.Fatal("Error opening database:", err)
		}

		db.AutoMigrate(&models.URL{})
	})
	return db
}

func Close() error {
	if db == nil {
		return nil
	}
	return db.Close()
}
