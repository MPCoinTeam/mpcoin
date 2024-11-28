package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// origin := c.Request.Header.Get("Origin")

		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", "*")                                            // Allow all origins, you can restrict this
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")             // Allowed methods
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization") // Allowed headers
		c.Header("Access-Control-Allow-Credentials", "true")                                    // If you want to allow credentials (cookies)
		c.Header("Access-Control-Max-Age", "86400")                                             // Cache preflight response for 1 day

		// Handle preflight request (OPTIONS)
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		// Continue to the next middleware/handler
		c.Next()
	}
}
