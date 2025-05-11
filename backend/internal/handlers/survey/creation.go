package handlers

import (
    "strconv"
    "database/sql"
    "log"
    "net/http"
    "my-auth-app/internal/models"
    "my-auth-app/internal/database"
    "github.com/gin-gonic/gin"
    "my-auth-app/internal/utils"
)

func createSurvey(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var survey models.Survey
        if err := c.ShouldBindJSON(&survey); err != nil {
          c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос"})
          return
        }
    
        
        userID, err := utils.GetUserIDFromToken(c)
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

func getUserSurveys(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
            return
        }

        surveys, err := database.GetSurveysByCreator(db, userID)
        if err != nil {
            log.Printf("Ошибка получения опросов: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Не удалось получить опросы",
                "details": err.Error(),
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "surveys": surveys,
        })
    }
}

func getSurveyParticipants(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Получаем ID опроса
        surveyID, err := strconv.Atoi(c.Param("survey_id"))
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID опроса"})
            return
        }

        // Проверяем права доступа
        userID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
            return
        }

        ownerID, err := database.GetSurveyOwnerID(db, surveyID)
        if err != nil || ownerID != userID {
            c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к статистике опроса"})
            return
        }

        // Получаем данные
        participants, err := database.GetSurveyParticipants(db, surveyID)
        if err != nil {
            log.Printf("Ошибка получения участников: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения данных"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "survey_id": surveyID,
            "participants": participants,
        })
    }
}