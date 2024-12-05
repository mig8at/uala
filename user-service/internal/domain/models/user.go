package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        string    `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	Email     string    `gorm:"not null;unique"`
	Nickname  string    `gorm:"not null;unique"`
	Bio       string    `gorm:"type:text"`
	Avatar    string    `gorm:"type:text"`
	Followers int       `gorm:"default:0"`
	Following int       `gorm:"default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (tag *User) BeforeCreate(tx *gorm.DB) (err error) {
	if tag.ID == "" {
		tag.ID = uuid.New().String()
	}
	return
}

type Follower struct {
	ID         string `gorm:"primaryKey"`
	UserID     string `gorm:"index;not null"`
	FollowerID string `gorm:"index;not null"`
}

func (tag *Follower) BeforeCreate(tx *gorm.DB) (err error) {
	if tag.ID == "" {
		tag.ID = uuid.New().String()
	}
	return
}
