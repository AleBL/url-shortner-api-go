package models

type URL struct {
	ID          uint   `gorm:"primary_key"`
	OriginalURL string `gorm:"type:text;not null"`
	ShortURL    string `gorm:"type:text;unique;not null"`
}
