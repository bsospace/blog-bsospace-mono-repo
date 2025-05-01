package response

import "github.com/gin-gonic/gin"

type Success struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Error struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Response helper
func JSONSuccess(c *gin.Context, message string, data interface{}) {
	c.JSON(200, Success{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func JSONError(c *gin.Context, status int, message string, code string) {
	c.JSON(status, Error{
		Success: false,
		Message: message,
		Code:    code,
	})
}
