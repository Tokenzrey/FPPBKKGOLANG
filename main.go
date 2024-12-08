package main

import (
	"fmt"
	"time"

	"github.com/Tokenzrey/FPPBKKGOLANG/api/router"
	"github.com/Tokenzrey/FPPBKKGOLANG/config"
	"github.com/Tokenzrey/FPPBKKGOLANG/db/initializers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
	config.LoadEnvVariables()
	initializers.ConnectDB()
}

func main() {
	fmt.Println("BE Berhasil!")

	// Inisialisasi router
	r := gin.Default()

	r.Static("/uploads", "./uploads")

	// Middleware CORS: Mengizinkan semua origin
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: false, //
		MaxAge:           12 * time.Hour,
	}))

	// Register routes
	router.GetRoute(r)

	// Jalankan server
	r.Run()
}
