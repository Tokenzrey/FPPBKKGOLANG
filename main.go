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

    // Middleware CORS
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000"}, // Domain yang diizinkan
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, // Metode yang diizinkan
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, // Header yang diizinkan
        ExposeHeaders:    []string{"Content-Length"}, // Header yang dapat diakses di client
        AllowCredentials: true, // Mengizinkan kredensial (cookies, Authorization header, dll.)
        MaxAge:           12 * time.Hour, // Durasi cache untuk preflight request
    }))

    // Register routes
    router.GetRoute(r)

    // Jalankan server
    r.Run()
}
