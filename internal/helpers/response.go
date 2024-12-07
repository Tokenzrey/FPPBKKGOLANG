package helpers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response struct for standardized API response
type APIResponse struct {
	Status  string      `json:"status"`  // "success" or "error"
	Data    interface{} `json:"data"`    // Data payload (null if error)
	Message string      `json:"message"` // Description of the response
}

// Helper function to send success response
func SuccessResponse(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, APIResponse{
		Status:  "success",
		Data:    data,
		Message: message,
	})
}

// Helper function to send error response
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, APIResponse{
		Status:  "error",
		Data:    nil,
		Message: message,
	})
}