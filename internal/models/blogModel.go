package models;

import "gorm.io/gorm"

type Blog struct {
	gorm.Model
	Judul     string `json:"judul"`
	Content    string `json:"content" gorm:"type:TEXT"`
	Thumbnaill string `json:"thumbnaill"`
}