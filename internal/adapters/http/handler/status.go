package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":      "ok",
		"server_time": time.Now().Format(time.RFC3339),
	})
}
