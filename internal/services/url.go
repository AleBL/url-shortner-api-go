package services

import (
	"errors"
	"url-shortener/internal/database"
	"url-shortener/models"

	"github.com/jinzhu/gorm"
	"github.com/mattn/go-sqlite3"
)

var (
	ErrNotFound      = errors.New("short URL not found")
	ErrShortURLTaken = errors.New("short URL already taken")
)

func SaveURL(originalURL, shortURL string) error {
	db := database.ConnectDB()
	url := models.URL{OriginalURL: originalURL, ShortURL: shortURL}

	err := db.Create(&url).Error
	var sqliteErr sqlite3.Error
	if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
		return ErrShortURLTaken
	}
	return err
}

func FindUrlByShortUrl(shortURL string) (string, error) {
	db := database.ConnectDB()
	var url models.URL

	err := db.Where("short_url = ?", shortURL).First(&url).Error
	if gorm.IsRecordNotFoundError(err) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	return url.OriginalURL, nil
}
