package utils

import (
    "errors"
    "my-auth-app/internal/jwt"
    "github.com/gin-gonic/gin"
)

func GetUserIDFromToken(c *gin.Context) (int, error) {
    tokenString := c.GetHeader("Authorization")
    if tokenString == "" {
        return 0, errors.New("отсутствует токен")
    }
    if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
        tokenString = tokenString[7:]
    }
    return jwt.ExtractUserIDFromJWT(tokenString)
}