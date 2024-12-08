package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Tokenzrey/FPPBKKGOLANG/api/middleware"
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
	query := db.Preload("User") // Preload user details

	switch sort {
		case "likes":
			// Subquery for sorting by likes
			return query.Select("blogs.*, (SELECT COUNT(*) FROM likes WHERE likes.blog_id = blogs.id) as like_count").
				Order("like_count DESC")
		case "comments":
			// Subquery for sorting by comments
			return query.Select("blogs.*, (SELECT COUNT(*) FROM comments WHERE comments.blog_id = blogs.id) as comment_count").
				Order("comment_count DESC")
		default:
			// Default sorting by blog creation date
			return query.Order("blogs.created_at DESC")
	}
}

	// Perform pagination and query execution
	result, err := pagination.Paginate(initializers.DB, page, perPage, rawFunc, &blogs)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("Error executing query: %v\n", err)
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
		// Base query with user details preloaded
		query := db.Preload("User").Order("blogs.created_at DESC")

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

		// Return the modified query
		return query
	}

	// Perform pagination and query execution
	result, err := pagination.Paginate(initializers.DB, page, perPage, rawFunc, &blogs)
	if err != nil {
		// Log the error for debugging
		fmt.Printf("Error executing query: %v\n", err)
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve blogs")
		return
	}

	// Return paginated blog search results
	helpers.SuccessResponse(c, result, "Blogs retrieved successfully")
}


// func PostBlog(c *gin.Context) {
// 	// Define user input structure
// 	userID, err := middleware.GetUserIDFromToken(c)
// 	if err != nil {
// 		// Respond with unauthorized if token is missing, invalid, or expired
// 		helpers.ErrorResponse(c, http.StatusUnauthorized, "Token missing, invalid, or expired")
// 		return
// 	}

// 	var blogInput struct {
// 		Judul     string `json:"judul" validate:"required"`
// 		Content   string `json:"content" validate:"required" gorm:"type:TEXT"`
// 		Thumbnail string `json:"thumbnail" validate:"required"`
// 	}

// 	// Bind JSON input
// 	if err := c.ShouldBindJSON(&blogInput); err != nil {
// 		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid input format")
// 		return
// 	}

// 	// Validate input fields
// 	if err := validate.Struct(blogInput); err != nil {
// 		if errs, ok := err.(validator.ValidationErrors); ok {
// 			// Concatenate all error messages into a single string
// 			var errorMessage string
// 			for _, e := range errs {
// 				errorMessage += e.Field() + ": " + e.ActualTag() + "; "
// 			}
// 			// Trim the trailing semicolon and space
// 			errorMessage = strings.TrimSuffix(errorMessage, "; ")

// 			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Validation failed: "+errorMessage)
// 			return
// 		}
// 		helpers.ErrorResponse(c, http.StatusBadRequest, "Validation error occurred")
// 		return
// 	}

// 	// Create a new user instance
// 	blog := models.Blog{
// 		Judul:     blogInput.Judul,
// 		Content:   blogInput.Content,
// 		Thumbnail: blogInput.Thumbnail,
// 		UserID:    uint(userID),
// 	}

// 	// Save the new user to the database
// 	if err := initializers.DB.Create(&blog).Error; err != nil {
// 		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create blog")
// 		return
// 	}

// 	// Prepare the response object excluding the password
// 	userResponse := struct {
// 		Judul     string    `json:"judul"`
// 		Content   string    `json:"content" gorm:"type:TEXT"`
// 		Thumbnail string    `json:"thumbnaill"`
// 		CreatedAt time.Time `json:"created_at"`
// 	}{
// 		Judul:     blog.Judul,
// 		Content:   blog.Content,
// 		Thumbnail: blog.Thumbnail,
// 		CreatedAt: blog.CreatedAt,
// 	}

// 	// Respond with the created user details
// 	helpers.SuccessResponse(c, userResponse, "Blog created successfully")
// }

func DeleteBlog(c *gin.Context) {
	// Get the blog ID from the URL parameter
	blogID := c.Param("id")

	// Get the user ID from the token
	userID, err := middleware.GetUserIDFromToken(c)
	if err != nil {
		// Respond with unauthorized if token is missing, invalid, or expired
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Token missing, invalid, or expired")
		return
	}

	// Find the blog in the database
	var blog models.Blog
	if err := initializers.DB.First(&blog, blogID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Blog not found
			helpers.ErrorResponse(c, http.StatusNotFound, "Blog not found")
			return
		}
		// General error finding the blog
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to find blog")
		return
	}

	// Check if the blog belongs to the current user
	if blog.UserID != uint(userID) {
		helpers.ErrorResponse(c, http.StatusForbidden, "You are not authorized to delete this blog")
		return
	}

	// Delete the blog
	if err := initializers.DB.Delete(&blog).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete blog")
		return
	}

	// Respond with success
	helpers.SuccessResponse(c, gin.H{"id": blog.ID}, "Blog deleted successfully")
}

func PostBlog(c *gin.Context) {
	// Get the user ID from the token
	userID, err := middleware.GetUserIDFromToken(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Token missing, invalid, or expired")
		return
	}

	// Parse form data
	judul := c.PostForm("judul")
	content := c.PostForm("content")

	// Validate input fields
	if judul == "" || content == "" {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Judul and content are required")
		return
	}

	// Check if the form contains an image file
	file, err := c.FormFile("thumbnail")
	if err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Thumbnail image is required")
		return
	}

	// Validate the file size (max 3MB)
	const maxFileSize = 3 * 1024 * 1024
	if file.Size > maxFileSize {
		helpers.ErrorResponse(c, http.StatusBadRequest, "File size exceeds the 3MB limit")
		return
	}

	// Create an upload directory if it doesn't exist
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create upload directory")
		return
	}

	// Save the uploaded file with a unique filename
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext) // Unique file name
	// filePath := filepath.Join(uploadDir, fileName)             // Save to uploads directory

	if err := c.SaveUploadedFile(file, fileName); err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to save the file")
		return
	}

	// Save blog data to the database
	blog := models.Blog{
		Judul:     judul,
		Content:   content,
		Thumbnail: fileName,
		UserID:    uint(userID),
	}

	if err := initializers.DB.Create(&blog).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create blog")
		return
	}

	// Respond with success
	helpers.SuccessResponse(c, gin.H{
		"message": "Blog created successfully",
		"blog": gin.H{
			"id":        blog.ID,
			"judul":     blog.Judul,
			"content":   blog.Content,
			"thumbnail": fmt.Sprintf("/uploads/%s", fileName), // Publicly accessible path
		},
	}, "Blog created successfully")
}

// GetBlogByID retrieves a blog post by its ID, including likes and comments.
func GetBlog(c *gin.Context) {
	// Get the Blog ID from the request parameters
	blogID := c.Param("id")
	if blogID == "" {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Blog ID is required")
		return
	}

	// Retrieve the authenticated user ID from the token
	userID, err := middleware.GetUserIDFromToken(c)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Token missing, invalid, or expired")
		return
	}

	// Retrieve the blog details
	var blog models.Blog
	if err := initializers.DB.Preload("User").First(&blog, blogID).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "Blog not found")
		return
	}

	// Retrieve the number of likes for the blog
	var likesCount int64
	if err := initializers.DB.Model(&models.Like{}).Where("blog_id = ?", blogID).Count(&likesCount).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Error counting likes")
		return
	}

	// Check if the user has already liked the blog
	var hasLiked bool
	if err := initializers.DB.Model(&models.Like{}).Where("user_id = ? AND blog_id = ?", userID, blogID).First(&models.Like{}).Error; err == nil {
		hasLiked = true
	}

	// Retrieve comments for the blog
	var comments []models.Comment
	if err := initializers.DB.Where("blog_id = ?", blogID).Find(&comments).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Error fetching comments")
		return
	}

	// Construct the response payload
	response := gin.H{
		"id":        blog.ID,
		"title":     blog.Judul,
		"content":   blog.Content,
		"thumbnail": blog.Thumbnail,
		"author": gin.H{
			"name":  blog.User.Name,
			"email": blog.User.Email,
		},
		"likes": gin.H{
			"count":     likesCount,
			"userLiked": hasLiked,
		},
		"comments": comments,
	}

	helpers.SuccessResponse(c, response, "Blog fetched successfully")
}
