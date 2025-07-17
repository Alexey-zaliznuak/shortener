package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Link struct {
	// TODO: если получится освободиться от авто тестов, можно использовать постгрес фичи
	// ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	CreatedAt time.Time
	UpdatedAt time.Time

	FullURL  string `gorm:"index;not null"`
	Shortcut string `gorm:"uniqueIndex;not null;size:8"`
}

func (s *Link) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
