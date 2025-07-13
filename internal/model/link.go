package model

import (
	"time"

	"github.com/Alexey-zaliznuak/shortener/internal/repository/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Link struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	FullUrl  string `gorm:"uniqueIndex;not null"`
	ShortUrl string `gorm:"uniqueIndex;not null;size:8"`
}

func (s *Link) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

func init() {
	database.Client.AutoMigrate(&Link{})
}
