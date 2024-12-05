package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tweet struct {
	ID            string         `gorm:"type:uuid;primaryKey"`
	UserID        string         `gorm:"type:uuid;index;not null"`
	Content       string         `gorm:"size:280;not null"`
	Tags          []Tag          `gorm:"many2many:tweet_tags"`
	Comments      []Comment      `gorm:"foreignKey:TweetID"`
	CountComments int            `gorm:"type:int;not null;default:0"`
	Likes         int            `gorm:"type:int;not null;default:0"`
	Shares        int            `gorm:"type:int;not null;default:0"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (tweet *Tweet) BeforeCreate(tx *gorm.DB) (err error) {
	if tweet.ID == "" {
		tweet.ID = uuid.New().String()
	}
	return
}

type Tag struct {
	ID        string         `gorm:"type:uuid;primaryKey"`
	Name      string         `gorm:"size:50;unique;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (tag *Tag) BeforeCreate(tx *gorm.DB) (err error) {
	if tag.ID == "" {
		tag.ID = uuid.New().String()
	}
	return
}

type Comment struct {
	ID        string         `gorm:"type:uuid;primaryKey"`
	UserID    string         `gorm:"type:uuid;index;not null"`
	TweetID   string         `gorm:"type:uuid;index;not null"`
	Content   string         `gorm:"size:280;not null"`
	Likes     int            `gorm:"type:int;not null,default:0"`
	Shares    int            `gorm:"type:int;not null,default:0"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (comment *Comment) BeforeCreate(tx *gorm.DB) (err error) {
	if comment.ID == "" {
		comment.ID = uuid.New().String()
	}
	return
}
