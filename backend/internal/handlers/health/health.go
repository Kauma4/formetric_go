package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func healthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "Server is running!"})
}
