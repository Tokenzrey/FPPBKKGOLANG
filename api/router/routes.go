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

	r.Use(middleware.RequireAuth)
	r.POST("/api/logout", controllers.Logout)
	userRouter := r.Group("/api/users")
	{
		userRouter.GET("/", controllers.GetUsers)
		userRouter.GET("/:id/edit", controllers.EditUser)
		userRouter.PUT("/:id/update", controllers.UpdateUser)
	}

	// Comment routes
	// commentRouter := r.Group("/api/posts/:id/comment")
	// {
	// 	commentRouter.POST("/store", controllers.CommentOnPost)
	// 	commentRouter.GET("/:comment_id/edit", controllers.EditComment)
	// 	commentRouter.PUT("/:comment_id/update", controllers.UpdateComment)
	// 	commentRouter.DELETE("/:comment_id/delete", controllers.DeleteComment)
	// }
}
