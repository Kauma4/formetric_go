package analytics

import (
	"encoding/json"
	"net/http"
    "testing"
    "github.com/stretchr/testify/assert"
    "my-auth-app/internal/database"
    "github.com/gin-gonic/gin"
    "net/http/httptest"
)

func TestGetSurveyAnalytics(t *testing.T) {
    db := database.SetupTestDB()

 	 _, err := db.Exec(`
        INSERT INTO users (id, username, password) 
        VALUES (1, 'testuser', 'password');
        
        INSERT INTO surveys (id, title, description, created_by, max_ball) 
        VALUES (1, 'Test Survey', 'Test Description', 1, 10);
    `)
    if err != nil {
        t.Fatal("Ошибка подготовки данных:", err)
    }

	handler := getSurveyAnalytics(db)

    t.Run("Survey with responses", func(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Params = []gin.Param{{Key: "id", Value: "1"}}
    c.Set("userID", 1)

    handler(c)

    assert.Equal(t, http.StatusOK, w.Code, "Ожидается статус 200")
    var response map[string]interface{}
    assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
    assert.NotNil(t, response["TotalResponses"], "Ответ не должен быть пустым")
})

    t.Run("Survey with no responses", func(t *testing.T) {
        //surveyID := 2 // Опрос без ответов
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        c.Params = []gin.Param{{Key: "id", Value: "2"}}
        c.Set("userID", 1)

        handler(c)

        assert.Equal(t, 200, w.Code)
        var response map[string]interface{}
        assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
        assert.Equal(t, float64(0), response["TotalResponses"])
    })
}