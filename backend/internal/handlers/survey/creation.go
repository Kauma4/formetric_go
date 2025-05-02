package handlers

import (
    "database/sql"
    "log"
    "net/http"
    "my-auth-app/internal/models"
    "my-auth-app/internal/database"
    "github.com/gin-gonic/gin"
)

func createSurvey(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var survey models.Survey
        if err := c.ShouldBindJSON(&survey); err != nil {
          c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
          return
        }
    
        
        userID, err := getUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        survey.CreatedBy = userID

        if err := database.CreateSurvey(db, &survey); err != nil {
            log.Printf("Ошибка при создании опроса: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Не удалось создать опрос",
                "details": err.Error(),
            })
            return
        }

        log.Printf("Опрос создан: ID=%d, UserID=%d", survey.ID, userID)
        c.JSON(http.StatusOK, gin.H{
            "message": "Опрос успешно создан",
            "survey_id": survey.ID,
        })	
    }
}

func getSurveys(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        log.Println("Получен запрос на опросы")

        surveys, err := database.GetSurveys(db)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить опросы"})
            return
        }

        log.Println("Ответ с опросами:", surveys)
        c.JSON(http.StatusOK, surveys)
    }
}