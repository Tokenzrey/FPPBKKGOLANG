package router

import (
	"github.com/Tokenzrey/FPPBKKGOLANG/api/controllers"
	"github.com/Tokenzrey/FPPBKKGOLANG/api/middleware"
	"github.com/gin-gonic/gin"
)

func GetRoute(r *gin.Engine) {
	// User routes
	r.POST("/api/signup", controllers.Signup)
	r.POST("/api/login", controllers.Login)

	// Blog routes without authentication
	r.GET("/api/blogs", controllers.GetBlogs)          // Get paginated blogs
	r.GET("/api/blogs/search", controllers.SearchBlogs) // Search blogs by query

	// Routes requiring authentication
	r.Use(middleware.RequireAuth)
	userRouter := r.Group("/api/users")
	{
		userRouter.GET("/", controllers.GetUserDetail)     // Get user details
		userRouter.PUT("/update", controllers.UpdateUser) // Update user details
	}
}
