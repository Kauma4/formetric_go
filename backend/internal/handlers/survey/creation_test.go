package handlers

import (
	"encoding/json"
	"net/http"
    "testing"
    "github.com/stretchr/testify/assert"
    "my-auth-app/internal/models"
    "my-auth-app/internal/database"
	"database/sql"
    "my-auth-app/internal/jwt"
    "github.com/gin-gonic/gin"
    "net/http/httptest"
    "bytes"
)

func TestCreateSurvey(t *testing.T) {
    db := database.SetupTestDB()
    handler := createSurvey(db)

    // Добавьте пользователя
    _, err := db.Exec(`
        INSERT INTO users (id, username, password) 
        VALUES (1, 'testuser', 'password')
    `)
    if err != nil {
        t.Fatal(err)
    }

    // Сгенерируйте токен
    token, err := jwt.GenerateJWT(1, "testuser")
	if err != nil {
		t.Fatal("Ошибка генерации токена:", err)
	}
	req.Header.Set("Authorization", "Bearer " + token)

    t.Run("Valid survey", func(t *testing.T) {
        survey := models.Survey{
            Title:       "Test Survey",
            Description: sql.NullString{String: "Description", Valid: true},
            CreatedBy:   1,
        }
        body, _ := json.Marshal(survey)
        req, _ := http.NewRequest("POST", "/survey", bytes.NewBuffer(body))
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Authorization", "Bearer " + token) // Используйте валидный токен

        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Request = req
        c.Set("userID", 1)

        handler(c)

        assert.Equal(t, http.StatusOK, w.Code)
        var response map[string]interface{}
        json.Unmarshal(w.Body.Bytes(), &response)
        assert.Equal(t, "Опрос успешно создан", response["message"])
    })

    t.Run("Invalid survey - empty title", func(t *testing.T) {
		survey := models.Survey{
		Title:       "Test Survey",
		Description: sql.NullString{String: "A test survey", Valid: true}, // Явно укажите Valid: true
		CreatedBy:   1,
		}
		body, _ := json.Marshal(survey)
		req, _ := http.NewRequest("POST", "/survey", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		// Не устанавливайте токен, чтобы проверить обработку ошибки валидации

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		handler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
})
}

