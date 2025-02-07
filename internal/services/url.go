package services

import (
	"url-shortener/internal/database"
	"url-shortener/models"
)

var URL = struct {
	SaveURL func(originalURL, shortURL string) error
}{
	SaveURL: func(originalURL, shortURL string) error {
		db := database.ConnectDB()
		url := models.URL{OriginalURL: originalURL, ShortURL: shortURL}
		return db.Create(&url).Error
	},
}

func SaveURL(originalURL, shortURL string) error {
	db := database.ConnectDB()
	url := models.URL{OriginalURL: originalURL, ShortURL: shortURL}
	return db.Create(&url).Error
}

func FindUrlByShortUrl(shortURL string) (string, error) {
	db := database.ConnectDB()
	var url models.URL

	err := db.Find(&url, models.URL{ShortURL: shortURL}).Error

	if err != nil {
		return "", err
	}

	return url.OriginalURL, nil
}
