package handlers

import (
    "database/sql"
    "log"
    "strconv"
    "net/http"
    "my-auth-app/internal/models"
    "my-auth-app/internal/database"
    "my-auth-app/internal/utils"
    "github.com/gin-gonic/gin"
)


func createAnswerUser(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Получаем ID пользователя из токена
        userID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            log.Printf("Auth error: %v", err)
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        // Парсим тело запроса
        var req struct {
            Answers []models.AnswerUser `json:"answers"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            log.Printf("JSON parse error: %v", err)
            c.JSON(http.StatusBadRequest, gin.H{
                "error":    "Неверный формат запроса",
                "details":  err.Error(),
            })
            return
        }

        // Проверяем наличие ответов
        if len(req.Answers) == 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Нет ответов для сохранения"})
            return
        }

        // Начинаем транзакцию
        tx, err := db.Begin()
        if err != nil {
            log.Printf("Transaction start failed: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка начала транзакции"})
            return
        }
        defer func() {
            if err != nil {
                tx.Rollback()
            }
        }()

        // Получаем ID опроса из первого вопроса
        surveyID, err := database.GetSurveyIDByQuestionID(db, req.Answers[0].QuestionID)
        if err != nil {
            log.Printf("Survey ID fetch error: %v", err)
            c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка определения опроса"})
            return
        }

        // Удаляем предыдущие данные
        if err := database.DeleteUserAnswers(db, userID, surveyID); err != nil {
            log.Printf("Delete answers error: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка очистки ответов"})
            return
        }

        if err := database.DeleteUserResult(db, userID, surveyID); err != nil {
            log.Printf("Delete result error: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка очистки результатов"})
            return
        }

        totalScore := 0
        for _, answer := range req.Answers {
            // Привязываем пользователя к ответу
            answer.UserID = userID

            // Получаем полную информацию о вопросе
            question, err := database.GetQuestionWithAnswers(db, answer.QuestionID)
            if err != nil {
                log.Printf("Question fetch error: %v", err)
                c.JSON(http.StatusBadRequest, gin.H{
                    "error":        "Ошибка получения вопроса",
                    "question_id":  answer.QuestionID,
                })
                return
            }

            // Проверяем правильность ответа
            isCorrect, err := checkAnswerCorrectness(answer, question)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error":    "Ошибка проверки ответа",
                    "details":  err.Error(),
                })
                return
            }

            // Суммируем баллы
            if isCorrect {
                totalScore += question.Ball
            }

            // Сохраняем ответ
            if err := database.CreateAnswerUser(db, &answer); err != nil {
                log.Printf("Answer save error: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": "Ошибка сохранения ответа",
                })
                return
            }
        }

        // Сохраняем общий результат
        if err := database.SaveUserResult(db, surveyID, userID, totalScore); err != nil {
            log.Printf("Save result error: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения результата"})
            return
        }

        // Фиксируем транзакцию
        if err := tx.Commit(); err != nil {
            log.Printf("Transaction commit error: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка фиксации транзакции"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "message":     "Результаты успешно обновлены",
            "total_score": totalScore,
        })
    }
}


func getUserSimpleSurveyAnswers(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        surveyIDStr := c.Param("survey_id")
        surveyID, err := strconv.Atoi(surveyIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID опроса"})
            return
        }

        answers, err := database.GetUserSimpleSurveyAnswers(db, userID, surveyID)
        if err != nil {
            log.Printf("Ошибка при получении ответов: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить ответы"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "survey_id": surveyID,
            "user_id": userID,
            "answers": answers,
        })
    }
}

func getUserAnswersByID(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Проверяем токен для авторизации (ID создателя опроса)
        creatorID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        // Получаем survey_id из параметров URL
        surveyIDStr := c.Param("survey_id")
        surveyID, err := strconv.Atoi(surveyIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID опроса"})
            return
        }

        // Получаем user_id из тела запроса
        var requestBody struct {
            UserID int `json:"user_id"`
        }
        if err := c.ShouldBindJSON(&requestBody); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Требуется user_id в теле запроса"})
            return
        }

        // Проверяем, что опрос принадлежит создателю
        var creator int
        err = db.QueryRow("SELECT created_by FROM surveys WHERE id = $1", surveyID).Scan(&creator)
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "Опрос не найден"})
            return
        }
        if err != nil {
            log.Printf("Ошибка при проверке создателя опроса: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера при проверке доступа"})
            return
        }
        if creator != creatorID {
            c.JSON(http.StatusForbidden, gin.H{"error": "Вы не являетесь создателем опроса"})
            return
}

        // Получаем ответы пользователя
        answers, err := database.GetUserSimpleSurveyAnswers(db, requestBody.UserID, surveyID)
        if err != nil {
            log.Printf("Ошибка при получении ответов: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить ответы"})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "survey_id": surveyID,
            "user_id":   requestBody.UserID,
            "answers":   answers,
        })
    }
}
