package router

import (
	"net/http"

	"github.com/Tokenzrey/FPPBKKGOLANG/api/controllers"
	"github.com/Tokenzrey/FPPBKKGOLANG/api/middleware"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/helpers"
	"github.com/gin-gonic/gin"
)

func GetRoute(r *gin.Engine) {
	// Middleware untuk menangani rute yang tidak ditemukan
	r.NoRoute(func(c *gin.Context) {
		helpers.ErrorResponse(c, http.StatusNotFound, "Route not found")
	})

	// Public routes (no authentication required)
	r.POST("/api/signup", controllers.Signup)           // User signup
	r.POST("/api/login", controllers.Login)             // User login
	r.GET("/api/blogs", controllers.GetBlogs)           // Get paginated blogs
	r.GET("/api/blogs/search", controllers.SearchBlogs) // Search blogs by query

	// Routes requiring authentication
	authRouter := r.Group("/")
	authRouter.Use(middleware.RequireAuth)
	{
		// User-related routes
		userRouter := authRouter.Group("/api/users")
		{
			userRouter.GET("/", controllers.GetUserDetail)    // Get user details
			userRouter.PUT("/update", controllers.UpdateUser) // Update user details
			userRouter.POST("/like", controllers.GenerateLike)
			userRouter.POST("/comment", controllers.PostComment)
		}
	}
}
