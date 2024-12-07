package main

import (
	"fmt"
	"github.com/Tokenzrey/FPPBKKGOLANG/api/router"
	"github.com/Tokenzrey/FPPBKKGOLANG/config"
	"github.com/Tokenzrey/FPPBKKGOLANG/db/initializers"
	"github.com/gin-gonic/gin"
)

func init() {
	config.LoadEnvVariables()
	initializers.ConnectDB()
}

func main() {
	fmt.Println("Hello auth")
	r := gin.Default()
	router.GetRoute(r)

	r.Run()
}
