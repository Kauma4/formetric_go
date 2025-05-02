package handlers

import (
    "bytes"
    "database/sql"
    "io"
    "log"
    "strconv"
    "net/http"
    "my-auth-app/internal/models"
    "my-auth-app/internal/database"
    "github.com/gin-gonic/gin"
)

func createAnswerUser(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        bodyBytes, err := io.ReadAll(c.Request.Body)
        if err != nil {
            log.Printf("Failed to read request body: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось прочитать запрос"})
            return
        }
        log.Printf("Received request: %s", string(bodyBytes))
        c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))

        userID, err := getUserIDFromToken(c)
        if err != nil {
            log.Printf("Auth error: %v", err)
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        var req struct {
            Answers []models.AnswerUser `json:"answers"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            log.Printf("JSON parse error: %v", err)
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Неверный формат запроса",
                "details": err.Error(),
            })
            return
        }

        tx, err := db.Begin()
        if err != nil {
            log.Printf("Transaction start failed: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка начала транзакции"})
            return
        }
        defer func() {
            if err != nil {
                tx.Rollback()
                return
            }
            tx.Commit()
        }()

        for _, answer := range req.Answers {
            answer.UserID = int(userID)
            
            question, err := database.GetQuestionWithAnswers(db, answer.QuestionID)
            if err != nil {
                log.Printf("Question fetch error: %v", err)
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "Ошибка получения вопроса",
                    "question_id": answer.QuestionID,
                })
                return
            }

            isCorrect, err := checkAnswerCorrectness(answer, question)
            if err != nil {
                c.JSON(http.StatusBadRequest, gin.H{
                    "error": "Ошибка проверки ответа",
                    "details": err.Error(),
                })
                return
            }

            if isCorrect {
                if err := processCorrectAnswer(db, question, answer); err != nil {
                    log.Printf("Score update error: %v", err)
                    c.JSON(http.StatusInternalServerError, gin.H{
                        "error": "Ошибка начисления баллов",
                    })
                    return
                }
            }

            if err := database.CreateAnswerUser(db, &answer); err != nil {
                log.Printf("Answer save error: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": "Ошибка сохранения ответа",
                })
                return
            }
        }

        c.JSON(http.StatusOK, gin.H{"message": "Ответы успешно сохранены"})
    }
}

func getUserSimpleSurveyAnswers(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, err := getUserIDFromToken(c)
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