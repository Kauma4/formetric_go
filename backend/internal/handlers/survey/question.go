package handlers

import (
   // "errors"
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

func createQuestion(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawBody, _ := c.GetRawData()
		log.Printf("Raw request body: %s", string(rawBody))
		c.Request.Body = io.NopCloser(bytes.NewBuffer(rawBody))

		var question models.Question
		if err := c.ShouldBindJSON(&question); err != nil {
			log.Printf("JSON bind error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный запрос: " + err.Error()})
			return
		}

		if question.QuestionText == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Текст вопроса не может быть пустым"})
			return
		}

		userID, err := getUserIDFromToken(c)
		if err != nil {
			log.Printf("Auth error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		surveyOwnerID, err := database.GetSurveyOwnerID(db, question.SurveyID)
		if err != nil {
			log.Printf("Survey owner check error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "Ошибка проверки опроса",
				"details":  err.Error(),
			})
			return
		}

		if surveyOwnerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Нет прав на добавление вопроса"})
			return
		}

		if question.IsTest {
			if len(question.Answers) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Для тестового вопроса нужны варианты ответов"})
				return
			}
			question.CorrectAnswer = ""
		} else {
			question.Answers = nil
		}

		// Начало транзакции
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

		// Создание вопроса в транзакции
		if err := database.CreateQuestion(tx, &question); err != nil {
			log.Printf("DB error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":    "Не удалось создать вопрос",
				"details":  err.Error(),
			})
			return
		}

		// Обновление max_ball в транзакции
		if err := database.UpdateSurveyMaxBall(tx, question.SurveyID, question.Ball); err != nil {
			log.Printf("Ошибка обновления max_ball: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":    "Ошибка обновления максимального балла",
				"details":  err.Error(),
			})
			return
		}

		log.Printf("Question created: ID=%d, SurveyID=%d", question.ID, question.SurveyID)
		
		c.JSON(http.StatusOK, gin.H{
			"message": "Вопрос успешно создан",
			"question_id": question.ID,
			"details": gin.H{
				"text":    question.QuestionText,
				"ball":    question.Ball,
				"type":    question.GetQuestionType(),
				"is_test": question.IsTest,
			},
		})
	}
}

func getQuestions(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        surveyIDStr := c.Param("survey_id")
        surveyID, err := strconv.Atoi(surveyIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат ID опроса"})
            return
        }

        questions, err := database.GetQuestions(db, surveyID)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить вопросы"})
            return
        }

        var result []gin.H
        for _, question := range questions {
            item := gin.H{
                "id":            question.ID,
                "survey_id":     question.SurveyID,
                "is_required":   question.IsRequired,
                "is_test":       question.IsTest,
                "question_text": question.QuestionText,
                "ball":          question.Ball,
                "answers":       question.Answers,
            }
            
            // Добавляем correct_answer только если это текстовый вопрос
            if !question.IsTest {
                item["correct_answer"] = question.CorrectAnswer
            }
            
            result = append(result, item)
        }

        c.JSON(http.StatusOK, gin.H{"questions": result})
    }
}
