package initializers

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
)

var DB *gorm.DB

func ConnectDB() {
	var err error
	dns := os.Getenv("DNS")

	DB, err = gorm.Open(mysql.Open(dns), &gorm.Config{})

	if err != nil {
		panic("Database connection failed!")
	}
}
