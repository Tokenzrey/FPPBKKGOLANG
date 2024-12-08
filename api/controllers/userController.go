package controllers

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Tokenzrey/FPPBKKGOLANG/api/middleware"
	"github.com/Tokenzrey/FPPBKKGOLANG/db/initializers"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/helpers"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/models"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/validations"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

// Signup handles user registration
// @Description Registers a new user with detailed information (name, email, password, tanggal_lahir, biografi).
// @Tags User
// @Accept json
// @Produce json
// @Param user body object{ name=string, email=string, password=string, tanggal_lahir=string, biografi=string } true "User input"
// @Success 200 {object} object{status=string, data=models.User, message=string}
// @Failure 400 {object} object{status=string, message=string}
// @Failure 422 {object} object{status=string, message=string}
// @Failure 500 {object} object{status=string, message=string}
// @Router /signup [post]
func Signup(c *gin.Context) {
	// Define user input structure
	var userInput struct {
		Name         string `json:"name" validate:"required,min=2,max=50"`  // Minimum 2 characters, maximum 50
		Email        string `json:"email" validate:"required,email"`       // Valid email address required
		Password     string `json:"password" validate:"required,min=6"`    // Minimum 6 characters
		TanggalLahir string `json:"tanggal_lahir" validate:"required,datetime=2006-01-02"` // Date in `YYYY-MM-DD` format
		Biografi     string `json:"biografi" validate:"required,max=500"`  // Maximum 500 characters
	}

	// Bind JSON input
	if err := c.ShouldBindJSON(&userInput); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid input format")
		return
	}

	// Validate input fields
	if err := validate.Struct(userInput); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			// Concatenate all error messages into a single string
			var errorMessage string
			for _, e := range errs {
				errorMessage += e.Field() + ": " + e.ActualTag() + "; "
			}
			// Trim the trailing semicolon and space
			errorMessage = strings.TrimSuffix(errorMessage, "; ")

			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Validation failed: "+errorMessage)
			return
		}
		helpers.ErrorResponse(c, http.StatusBadRequest, "Validation error occurred")
		return
	}

	// Check if the email already exists in the database
	if validations.IsUniqueValue("users", "email", userInput.Email) {
		helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Email already exists")
		return
	}

	// Hash the user's password before saving to the database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userInput.Password), bcrypt.DefaultCost)
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Create a new user instance
	user := models.User{
		Name:         userInput.Name,
		Email:        userInput.Email,
		Password:     string(hashedPassword),
		TanggalLahir: userInput.TanggalLahir,
		Biografi:     userInput.Biografi,
	}

	// Save the new user to the database
	if err := initializers.DB.Create(&user).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Prepare the response object excluding the password
	userResponse := struct {
		ID           uint      `json:"id"`
		Name         string    `json:"name"`
		Email        string    `json:"email"`
		TanggalLahir string    `json:"tanggal_lahir"`
		Biografi     string    `json:"biografi"`
		CreatedAt    time.Time `json:"created_at"`
	}{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		TanggalLahir: user.TanggalLahir,
		Biografi:     user.Biografi,
		CreatedAt:    user.CreatedAt,
	}

	// Respond with the created user details
	helpers.SuccessResponse(c, userResponse, "User created successfully")
}

// Login authenticates a user and returns a JWT token
// @Description Authenticates a user using email and password, then returns a JWT token for session management.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body object{ email=string, password=string } true "User credentials"
// @Success 200 {object} object{status=string, data=object{token=string}, message=string}
// @Failure 400 {object} object{status=string, message=string}
// @Failure 401 {object} object{status=string, message=string}
// @Failure 500 {object} object{status=string, message=string}
// @Router /login [post]
func Login(c *gin.Context) {
	// Define the structure for user input with validation tags
	var userInput struct {
		Email    string `json:"email" validate:"required,email"` // Validate email format
		Password string `json:"password" validate:"required"`    // Password is required
	}

	// Bind JSON input
	if err := c.ShouldBindJSON(&userInput); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid input format")
		return
	}

	// Validate input fields
	if err := validate.Struct(userInput); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			// Concatenate all error messages into a single string
			var errorMessage string
			for _, e := range errs {
				errorMessage += e.Field() + ": " + e.ActualTag() + "; "
			}
			// Trim the trailing semicolon and space
			errorMessage = strings.TrimSuffix(errorMessage, "; ")

			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Validation failed: "+errorMessage)
			return
		}
		helpers.ErrorResponse(c, http.StatusBadRequest, "Validation error occurred")
		return
	}

	// Find the user in the database using email
	var user models.User
	if err := initializers.DB.First(&user, "email = ?", userInput.Email).Error; err != nil || user.ID == 0 {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Compare the provided password with the stored hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userInput.Password)); err != nil {
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate a JWT token for the authenticated user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,                              // Subject (user ID)
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(), // Expiration (30 days)
	})
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))
	if err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to create token")
		return
	}

	// Return the JWT token in the response
	responseData := gin.H{
		"token": tokenString,
	}

	// Success response with token
	helpers.SuccessResponse(c, responseData, "Login successful")
}

// GetUserDetail retrieves user details from the database based on the JWT token.
//
// @Summary Retrieve user details
// @Description Extracts user ID from the JWT token and retrieves corresponding user details from the database.
// @Tags User
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} object{status=string,data=object{id=uint,name=string,email=string,tanggal_lahir=string,biografi=string},message=string} "User retrieved successfully"
// @Failure 401 {object} object{status=string,message=string} "Unauthorized: Token missing, invalid, or expired"
// @Failure 404 {object} object{status=string,message=string} "User not found"
// @Router /user/details [get]
func GetUserDetail(c *gin.Context) {
	// Extract user ID from the token
	userID, err := middleware.GetUserIDFromToken(c)
	if err != nil {
		// Respond with unauthorized if token is missing, invalid, or expired
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Token missing, invalid, or expired")
		return
	}

	// Find the user in the database using the extracted user ID
	var user models.User
	if err := initializers.DB.First(&user, userID).Error; err != nil || user.ID == 0 {
		helpers.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	// Prepare response data excluding sensitive fields
	userResponse := gin.H{
		"id":           user.ID,
		"name":         user.Name,
		"email":        user.Email,
		"tanggal_lahir": user.TanggalLahir,
		"biografi":     user.Biografi,
	}

	// Send successful response
	helpers.SuccessResponse(c, userResponse, "User retrieved successfully")
}

// UpdateUser modifies user details in the database.
//
// @Summary Update user details
// @Description Updates the user's name, email, tanggal_lahir, and biografi in the database.
// @Tags User
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body object{name=string,email=string,tanggal_lahir=string,biografi=string} true "Updated user data"
// @Success 200 {object} object{status=string,data=models.User,message=string} "User updated successfully"
// @Failure 400 {object} object{status=string,message=string} "Invalid input format"
// @Failure 404 {object} object{status=string,message=string} "User not found"
// @Failure 422 {object} object{status=string,message=string} "Validation failed or email already exists"
// @Failure 500 {object} object{status=string,message=string} "Internal server error"
// @Router /users/{id} [put]
func UpdateUser(c *gin.Context) {
	// Extract user ID from the token
	id, err := middleware.GetUserIDFromToken(c)
	if err != nil {
		// Respond with unauthorized if token is missing, invalid, or expired
		helpers.ErrorResponse(c, http.StatusUnauthorized, "Token missing, invalid, or expired")
		return
	}

	// Define the structure for input validation
	var userInput struct {
		Name         string `json:"name" validate:"required,min=2,max=50"`      // Name: Min 2, Max 50 characters
		Email        string `json:"email" validate:"required,email"`           // Email: Valid email format
		TanggalLahir string `json:"tanggal_lahir" validate:"required,datetime=2006-01-02"` // Tanggal Lahir: Date format required (e.g., YYYY-MM-DD)
		Biografi     string `json:"biografi" validate:"omitempty,max=500"`     // Biografi: Optional, Max 500 characters
	}

	// Bind JSON input
	if err := c.ShouldBindJSON(&userInput); err != nil {
		helpers.ErrorResponse(c, http.StatusBadRequest, "Invalid input format")
		return
	}

	// Validate input fields
	if err := validate.Struct(userInput); err != nil {
		if errs, ok := err.(validator.ValidationErrors); ok {
			// Concatenate all error messages into a single string
			var errorMessage string
			for _, e := range errs {
				errorMessage += e.Field() + ": " + e.ActualTag() + "; "
			}
			// Trim the trailing semicolon and space
			errorMessage = strings.TrimSuffix(errorMessage, "; ")

			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Validation failed: "+errorMessage)
			return
		}
		helpers.ErrorResponse(c, http.StatusBadRequest, "Validation error occurred")
		return
	}

	// Find the user in the database
	var user models.User
	if err := initializers.DB.First(&user, id).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	// If the new email is different, check if it's unique
	if user.Email != userInput.Email {
		if validations.IsUniqueValue("users", "email", userInput.Email) {
			helpers.ErrorResponse(c, http.StatusUnprocessableEntity, "Email already exists")
			return
		}
	}

	// Update user fields
	user.Name = userInput.Name
	user.Email = userInput.Email
	user.TanggalLahir = userInput.TanggalLahir
	user.Biografi = userInput.Biografi

	// Save changes to the database
	if err := initializers.DB.Save(&user).Error; err != nil {
		helpers.ErrorResponse(c, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// Respond with the updated user data
	userResponse := gin.H{
		"id":           user.ID,
		"name":         user.Name,
		"email":        user.Email,
		"tanggal_lahir": user.TanggalLahir,
		"biografi":     user.Biografi,
	}
	helpers.SuccessResponse(c, userResponse, "User updated successfully")
}
