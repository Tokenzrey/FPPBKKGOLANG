package controllers

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Tokenzrey/FPPBKKGOLANG/db/initializers"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/helpers"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/models"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/pagination"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/validations"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Signup creates a new user
func Signup(c *gin.Context) {
	var userInput struct {
		Name     string `json:"name" binding:"required,min=2,max=50"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&userInput); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			validationErrors := validations.FormatValidationErrors(errs)
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Validation failed: "+validationErrors["Email"])
			return
		}
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid input format")
		return
	}

	if !validations.IsUniqueValue("users", "email", userInput.Email) {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Email already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userInput.Password), bcrypt.DefaultCost)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user := models.User{
		Name:     userInput.Name,
		Email:    userInput.Email,
		Password: string(hashedPassword),
	}

	if err := initializers.DB.Create(&user).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	helpers.SuccessResponse(c, user, "User created successfully")
}

// Login authenticates a user
func Login(c *gin.Context) {
	var userInput struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&userInput); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid input format")
		return
	}

	var user models.User
	if err := initializers.DB.First(&user, "email = ?", userInput.Email).Error; err != nil || user.ID == 0 {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userInput.Password)); err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create token")
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	helpers.SuccessResponse(c, gin.H{"token": tokenString}, "Login successful")
}

// Logout removes user authentication
func Logout(c *gin.Context) {
	c.SetCookie("Authorization", "", 0, "", "", false, true)
	helpers.SuccessResponse(c, nil, "Logout successful")
}

// GetUsers retrieves a list of users
func GetUsers(c *gin.Context) {
	var users []models.User

	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)

	perPageStr := c.DefaultQuery("perPage", "5")
	perPage, _ := strconv.Atoi(perPageStr)

	result, err := pagination.Paginate(initializers.DB, page, perPage, nil, &users)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve users")
		return
	}

	helpers.SuccessResponse(c, result, "Users retrieved successfully")
}

// EditUser retrieves a single user by ID
func EditUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	if err := initializers.DB.First(&user, id).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	helpers.SuccessResponse(c, user, "User retrieved successfully")
}

// UpdateUser modifies user details
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var userInput struct {
		Name  string `json:"name" binding:"required,min=2,max=50"`
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&userInput); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			validationErrors := validations.FormatValidationErrors(errs)
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Validation failed: "+validationErrors["Email"])
			return
		}
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid input format")
		return
	}

	var user models.User
	if err := initializers.DB.First(&user, id).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	if user.Email != userInput.Email && !validations.IsUniqueValue("users", "email", userInput.Email) {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Email already exists")
		return
	}

	user.Name = userInput.Name
	user.Email = userInput.Email

	if err := initializers.DB.Save(&user).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to update user")
		return
	}

	helpers.SuccessResponse(c, user, "User updated successfully")
}
