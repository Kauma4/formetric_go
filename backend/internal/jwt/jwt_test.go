package jwt

import (
	"time"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestExtractUserIDFromJWT(t *testing.T) {
    // Тест 1: Валидный токен
    t.Run("Valid token", func(t *testing.T) {
        token, err := GenerateJWT(1, "testuser")
        assert.NoError(t, err)

        userID, err := ExtractUserIDFromJWT(token)
        assert.NoError(t, err)
        assert.Equal(t, 1, userID)
    })

    // Тест 2: Истекший токен
    t.Run("Expired token", func(t *testing.T) {
		expiresAt := time.Now().Add(-1 * time.Hour)
        expiredToken, _ := GenerateJWTWithExpiration(1, "testuser", expiresAt) // Истекший токен
        _, err := ExtractUserIDFromJWT(expiredToken)
        assert.Error(t, err)
        assert.ErrorContains(t, err, "Token is expired")
    })

    // Тест 3: Некорректный токен
    t.Run("Invalid token", func(t *testing.T) {
        _, err := ExtractUserIDFromJWT("invalid.token.here")
        assert.Error(t, err)
        assert.ErrorContains(t, err, "invalid character")
    })
}