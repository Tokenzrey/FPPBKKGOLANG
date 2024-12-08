package controllers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Tokenzrey/FPPBKKGOLANG/api/middleware"
	"github.com/Tokenzrey/FPPBKKGOLANG/db/initializers"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/helpers"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// Will generate likes
// @Description Registers a new row in the "like" table
// @Tags Like
// @Accept json
// @Produce json
// @Param like body object{ user_id=uint, blog_id=uint }
// @Success 200 {object} object{status=string, data=models.Like, message=string}
// @Failure 400 {object} object{status=string, message=string}
// @Failure 422 {object} object{status=string, message=string}
// @Failure 500 {object} object{status=string, message=string}
// @Router /like/{blog_id}/{user_id} [post]

func GenerateLike(c *gin.Context) {
	// Extract user ID from the token
	userID, err := middleware.GetUserIDFromToken(c)
	if err != nil {
		// Respond with unauthorized if token is missing, invalid, or expired
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Token missing, invalid, or expired")
		return
	}

	// Bind Blog ID from request
	var blogLiked struct {
		BlogID uint `json:"blog_id" binding:"required"`
	}

	// Bind and validate JSON input
	if err := c.ShouldBindJSON(&blogLiked); err != nil {
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
	if err := initializers.DB.First(&blog, blogLiked.BlogID).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Blog not found")
		return
	}

	// Check if user already liked blog
	var already_liked models.Like
	result := initializers.DB.Where("user_id = ? AND blog_id = ?", userID, blogLiked.BlogID).First(&already_liked)

	if result.Error == nil { // blog is already liked, so we unlike it
		if err := initializers.DB.Delete(&already_liked).Error; err != nil {
			helpers.ErrorResponse(c, http.StatusInternalServerError, "Error removing like")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Blog unliked successfully",
			"liked":   false,
		})
		return
	}

	// blog is not liked yet
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		newLike := models.Like{
			UserID: uint(userID),
			BlogID: blogLiked.BlogID,
		}

		if err := initializers.DB.Create(&newLike).Error; err != nil {
			helpers.ErrorResponse(c, http.StatusInternalServerError, "Error liking blog")
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Blog liked successfully",
			"liked":   true,
		})
		return
	}

	helpers.ErrorResponse(c, http.StatusInternalServerError, "Unexpected Error Processing Like")
}

func ShowLike(c *gin.Context) {
	// Extract user ID from the token
	userID, errUser := middleware.GetUserIDFromToken(c)
	if errUser != nil {
		// Respond with unauthorized if token is missing, invalid, or expired
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Token missing, invalid, or expired")
		return
	}

	// Bind Blog ID from request
	var blogLiked struct {
		BlogID uint `json:"blog_id" binding:"required"`
	}

	// Bind and validate JSON input
	if err := c.ShouldBindJSON(&blogLiked); err != nil {
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
	if err := initializers.DB.First(&blog, blogLiked.BlogID).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Blog not found")
		return
	}

	// count the likes for the specified blog
	var likeCount int64
	if err := initializers.DB.Model(&models.Like{}).Where("blog_id = ?", blogLiked.BlogID).Count(&likeCount).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Error Counting Likes")
		return
	}

	// check if user has liked the blog or not
	var hasLiked bool

	var likeExists models.Like
	result := initializers.DB.Where("user_id = ? AND blog_id = ?", userID, blogLiked.BlogID).First(likeExists)
	hasLiked = result.Error == nil

	// returns JSON
	c.JSON(http.StatusOK, gin.H{
		"blog_id":       blogLiked.BlogID,
		"likes_count":   likeCount,
		"liked_by_user": hasLiked,
	})
}
