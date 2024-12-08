package models

import "gorm.io/gorm"

type Blog struct {
	gorm.Model
	Judul     string `json:"judul"`
	Content   string `json:"content" gorm:"type:TEXT"`
	Thumbnail string `json:"thumbnail"`

	// Virtual fields (not stored in DB)
	LikeCount    int64 `json:"like_count" gorm:"-"`
	CommentCount int64 `json:"comment_count" gorm:"-"`
	UserID       uint  `json:"user_id"`
	User         User  `gorm:"foreignKey:UserID;references:ID"`
}
