package handlers

import (
	"strconv"
    "database/sql"
    "log"
    "net/http"
    "my-auth-app/internal/models"
    "my-auth-app/internal/utils"
    "my-auth-app/internal/database"
    "github.com/gin-gonic/gin"
)

func updateSurvey(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        surveyIDStr := c.Param("id")
        surveyID, err := strconv.Atoi(surveyIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID опроса"})
            return
        }

        // Получаем текущего пользователя
        userID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        // Проверка прав на опрос
        ownerID, err := database.GetSurveyOwnerID(db, surveyID)
        if err != nil || ownerID != userID {
            c.JSON(http.StatusForbidden, gin.H{"error": "Нет прав на редактирование"})
            return
        }

        // Парсим входные данные
        var request struct {
            Title       string             `json:"title"`
            Description string             `json:"description"`
            Questions   []models.Question `json:"questions"`
        }
        if err := c.ShouldBindJSON(&request); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
            return
        }

        // Начинаем транзакцию
        tx, err := db.Begin()
        if err != nil {
            log.Printf("Ошибка начала транзакции: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
            return
        }
        defer func() {
            if err != nil {
                tx.Rollback()
                return
            }
            tx.Commit()
        }()

        // Обновляем основную информацию опроса
        if err := database.UpdateSurvey(db, surveyID, request.Title, request.Description); err != nil {
            log.Printf("Ошибка обновления опроса: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления опроса"})
            return
        }

        // Обрабатываем вопросы
        var keepIDs []int
        for _, q := range request.Questions {
            q.SurveyID = surveyID
            
            // Валидация вопроса
            if q.QuestionText == "" {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Текст вопроса не может быть пустым"})
                return
            }
            
            if q.IsTest && len(q.Answers) == 0 {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Тестовый вопрос должен иметь варианты ответов"})
                return
            }

            if q.ID != 0 {
                // Обновление существующего вопроса
                keepIDs = append(keepIDs, q.ID)
                if err := database.UpdateQuestion(db, &q); err != nil {
                    log.Printf("Ошибка обновления вопроса %d: %v", q.ID, err)
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления вопроса"})
                    return
                }
            } else {
                // Создание нового вопроса
                if err := database.CreateQuestion(db, &q); err != nil {
                    log.Printf("Ошибка создания вопроса: %v", err)
                    c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания вопроса"})
                    return
                }
                keepIDs = append(keepIDs, q.ID)
            }
        }

        // Удаляем отсутствующие вопросы
        if err := database.DeleteQuestionsExcept(tx, surveyID, keepIDs); err != nil {
            log.Printf("Ошибка удаления вопросов: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления вопросов"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message": "Опрос и вопросы успешно обновлены",
            "survey_id": surveyID,
            "updated_questions": len(request.Questions),
        })
    }
}