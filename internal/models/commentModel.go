package models

import "gorm.io/gorm"

type Comment struct {
    gorm.Model
	Comment string `json:"comment"`
    UserID uint `json:"user_id"`
    BlogID uint `json:"blog_id"`

    User User `gorm:"foreignKey:UserID;references:ID"`
    Blog Blog `gorm:"foreignKey:BlogID;references:ID"`
}
