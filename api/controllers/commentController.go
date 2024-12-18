package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Tokenzrey/FPPBKKGOLANG/api/middleware"
	"github.com/Tokenzrey/FPPBKKGOLANG/db/initializers"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/helpers"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func PostComment(c *gin.Context) {
	// Extract user ID from the token
	userID, err := middleware.GetUserIDFromToken(c)
	if err != nil {
		// Respond with unauthorized if token is missing, invalid, or expired
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Token missing, invalid, or expired")
		return
	}

	// Bind Blog ID from request
	var inputComment struct {
		Comment string `json:"comment" binding:"required,min=5,max=250"`
		BlogID  uint   `json:"blog_id" binding:"required"`
	}

	// Bind and validate JSON input
	if err := c.ShouldBindJSON(&inputComment); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			// Gabungkan semua pesan error dalam satu string
			var errorMessage string
			for _, e := range errs {
				errorMessage += e.Field() + ": " + e.ActualTag() + "; "
			}
			// Trim karakter terakhir
			errorMessage = strings.TrimSpace(errorMessage)

			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Validation failed: "+errorMessage)
			return
		}
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid input format")
		return
	}

	// Check if blog exists
	var blog models.Blog
	if err := initializers.DB.First(&blog, inputComment.BlogID).Error; err != nil {
		if inputComment.BlogID == 0 {
			helpers.ErrorResponse(c, http.StatusNotFound, "Blog not found")
			return
		}
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Error checking blog")
		return
	}

	newComment := models.Comment{
		Comment: inputComment.Comment,
		UserID:  uint(userID),
		BlogID:  inputComment.BlogID,
	}

	if err := initializers.DB.Create(&newComment).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Error Commenting on Post")
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "Comment Posted!",
		"posted":  true,
	})
}

func ShowComments(c *gin.Context) {
	// Get blog_id from path parameter
	blogIDStr := c.Param("blog_id")
	blogID, err := strconv.ParseUint(blogIDStr, 10, 64)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid blog_id")
		return
	}

	// Check if blog exists
	var blog models.Blog
	if err := initializers.DB.First(&blog, blogID).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Blog not found")
		return
	}

	// Prepare a slice to store comments with user details
	var comments []struct {
		ID        uint      `json:"id"`
		Comment   string    `json:"comment"`
		CreatedAt time.Time `json:"created_at"`
		User_ID   uint      `json:"user_id"`
		User_Name string    `json:"user_name"`
	}

	// Fetch comments
	if err := initializers.DB.Preload("users").Model(&models.Comment{}).Select("comments.id, comments.comment, comments.created_at, users.id as user_id, users.name as user_name").Joins("LEFT JOIN users ON comments.user_id =users.id").Where("comments.blog_id = ?", blogID).Order("comments.created_at DESC").Scan(&comments).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Error Loading Comments")
		return
	}

	// If there's no comment
	if comments == nil {
		comments = []struct {
			ID        uint      `json:"id"`
			Comment   string    `json:"comment"`
			CreatedAt time.Time `json:"created_at"`
			User_ID   uint      `json:"user_id"`
			User_Name string    `json:"user_name"`
		}{}
	}

	c.JSON(http.StatusOK, gin.H{
		"blog_id":  blogID,
		"comments": comments,
		"count":    len(comments),
	})
}
