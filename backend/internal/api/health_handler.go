package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewHealthHandler(group *gin.RouterGroup) {

	group.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	group.HEAD("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
}
