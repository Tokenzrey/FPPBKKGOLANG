package tests

import (
	"github.com/Tokenzrey/FPPBKKGOLANG/db/initializers"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/models"
	"github.com/joho/godotenv"
	"log"
)

// DatabaseRefresh runs fresh migration
func DatabaseRefresh() {
	// Load env
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect DB
	initializers.ConnectDB()

	// Drop all the tables
	err = initializers.DB.Migrator().DropTable(models.User{}, models.Like{}, models.Blog{}, models.Comment{})
	if err != nil {
		log.Fatal("Table dropping failed")
	}

	// Migrate again
	err = initializers.DB.AutoMigrate(models.User{}, models.Like{}, models.Blog{}, models.Comment{})

	if err != nil {
		log.Fatal("Migration failed")
	}
}
