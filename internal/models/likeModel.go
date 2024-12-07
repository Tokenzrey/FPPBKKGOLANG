package models

import "gorm.io/gorm"

type Like struct {
    gorm.Model
    UserID uint `json:"user_id"` // Foreign key untuk User
    BlogID uint `json:"blog_id"` // Foreign key untuk Blog

    User User `gorm:"foreignKey:UserID;references:ID"`
    Blog Blog `gorm:"foreignKey:BlogID;references:ID"`
}
