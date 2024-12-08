package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Tokenzrey/FPPBKKGOLANG/db/initializers"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/helpers"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/models"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/pagination"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetBlogs retrieves a paginated list of blogs sorted by likes or comments
//
// @Summary Get blog list
// @Description Retrieves a list of blogs with pagination and sorting options (by likes or comments).
// @Tags Blog
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param perPage query int false "Items per page" default(10)
// @Param sort query string false "Sort by 'likes' or 'comments'" Enums(likes, comments)
// @Success 200 {object} object{status=string,data=object{blogs=[]models.Blog},message=string} "Blogs retrieved successfully"
// @Failure 400 {object} object{status=string,message=string} "Invalid sort parameter"
// @Failure 500 {object} object{status=string,message=string} "Internal server error"
// @Router /blogs [get]
func GetBlogs(c *gin.Context) {
	// Get query parameters for pagination and sorting
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid page parameter")
		return
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("perPage", "10"))
	if err != nil || perPage <= 0 || perPage > 100 {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid perPage parameter")
		return
	}

	sort := c.DefaultQuery("sort", "") // Default: no sorting

	// Validate sort parameter
	if sort != "" && sort != "likes" && sort != "comments" {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid sort parameter. Must be 'likes' or 'comments'")
		return
	}

	// Define the output structure
	var blogs []models.Blog

	// Apply sorting logic based on the 'sort' query parameter
	rawFunc := func(db *gorm.DB) *gorm.DB {
		switch sort {
		case "likes":
			// Join blogs with likes and count likes
			return db.Joins("LEFT JOIN likes ON likes.blog_id = blogs.id").
				Group("blogs.id").
				Select("blogs.*, COUNT(likes.id) as like_count").
				Order("like_count DESC")
		case "comments":
			// Join blogs with comments and count comments
			return db.Joins("LEFT JOIN comments ON comments.blog_id = blogs.id").
				Group("blogs.id").
				Select("blogs.*, COUNT(comments.id) as comment_count").
				Order("comment_count DESC")
		default:
			// Default sorting by blog creation date
			return db.Order("blogs.created_at DESC")
		}
	}

	// Perform pagination and query execution
	result, err := pagination.Paginate(initializers.DB, page, perPage, rawFunc, &blogs)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve blogs")
		return
	}

	// Return paginated blog list
	helpers.SuccessResponse(c, result, "Blogs retrieved successfully")
}

// SearchBlogs retrieves a paginated list of blogs filtered by search query
//
// @Summary Search blogs
// @Description Searches blogs by username, judul, or content with pagination.
// @Tags Blog
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param perPage query int false "Items per page" default(10)
// @Param search query string true "Search keyword"
// @Param filter query string false "Filter by 'username', 'judul', 'content' or 'all'" Enums(username, judul, content, all) default(all)
// @Success 200 {object} object{status=string,data=object{blogs=[]models.Blog},message=string} "Blogs retrieved successfully"
// @Failure 400 {object} object{status=string,message=string} "Invalid search filter"
// @Failure 500 {object} object{status=string,message=string} "Internal server error"
// @Router /blogs/search [get]
func SearchBlogs(c *gin.Context) {
	// Get query parameters for pagination and search
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page <= 0 {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid page parameter")
		return
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("perPage", "10"))
	if err != nil || perPage <= 0 || perPage > 100 {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid perPage parameter")
		return
	}

	search := strings.TrimSpace(c.Query("search")) // Trim whitespace to handle empty input
	filter := c.DefaultQuery("filter", "all")

	// Validate filter parameter
	validFilters := map[string]bool{"username": true, "judul": true, "content": true, "all": true}
	if _, isValid := validFilters[filter]; !isValid {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid filter parameter. Must be 'username', 'judul', 'content', or 'all'")
		return
	}

	// Define the output structure
	var blogs []models.Blog

	// Apply search logic based on the 'filter' query parameter
	rawFunc := func(db *gorm.DB) *gorm.DB {
		// Base query
		query := db.Order("blogs.created_at DESC")

		// Add search conditions if search parameter is not empty
		if search != "" {
			switch filter {
			case "username":
				// Search by username
				query = query.Joins("JOIN users ON users.id = blogs.user_id").
					Where("users.name LIKE ?", "%"+search+"%")
			case "judul":
				// Search by blog title
				query = query.Where("blogs.judul LIKE ?", "%"+search+"%")
			case "content":
				// Search by blog content
				query = query.Where("blogs.content LIKE ?", "%"+search+"%")
			default:
				// Search in all fields
				query = query.Joins("LEFT JOIN users ON users.id = blogs.user_id").
					Where("users.name LIKE ? OR blogs.judul LIKE ? OR blogs.content LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
			}
		}

		return query
	}

	// Perform pagination and query execution
	result, err := pagination.Paginate(initializers.DB, page, perPage, rawFunc, &blogs)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve blogs")
		return
	}

	// Return paginated blog search results
	helpers.SuccessResponse(c, result, "Blogs retrieved successfully")
}

