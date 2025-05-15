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
    "fmt"
    "strings"
)

func createAnswerUser(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Получаем ID пользователя из токена
        userID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            log.Printf("Ошибка авторизации: %v", err)
            c.JSON(401, gin.H{"error": err.Error()})
            return
        }

        // Парсим тело запроса
        var req struct {
            Answers []models.AnswerUser `json:"answers"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            log.Printf("Ошибка парсинга JSON: %v", err)
            c.JSON(400, gin.H{
                "error":   "Неверный формат запроса",
                "details": err.Error(),
            })
            return
        }

        // Проверяем наличие ответов
        if len(req.Answers) == 0 {
            c.JSON(400, gin.H{"error": "Нет ответов для сохранения"})
            return
        }

        // Начинаем транзакцию
        tx, err := db.Begin()
        if err != nil {
            log.Printf("Ошибка начала транзакции: %v", err)
            c.JSON(500, gin.H{"error": "Ошибка начала транзакции"})
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
            log.Printf("Ошибка получения ID опроса: %v", err)
            c.JSON(400, gin.H{"error": "Ошибка определения опроса"})
            return
        }

        // Удаляем предыдущие ответы
        if err := database.DeleteUserAnswers(db, userID, surveyID); err != nil {
            log.Printf("Ошибка удаления ответов: %v", err)
            c.JSON(500, gin.H{"error": "Ошибка очистки ответов"})
            return
        }

        if err := database.DeleteUserResult(db, userID, surveyID); err != nil {
            log.Printf("Ошибка удаления результатов: %v", err)
            c.JSON(500, gin.H{"error": "Ошибка очистки результатов"})
            return
        }

        totalScore := 0
        // Группируем ответы по question_id
        answersByQuestion := make(map[int][]models.AnswerUser)
        for _, answer := range req.Answers {
            answersByQuestion[answer.QuestionID] = append(answersByQuestion[answer.QuestionID], answer)
        }

        for questionID, answers := range answersByQuestion {
            // Получаем информацию о вопросе
            question, err := database.GetQuestionWithAnswers(db, questionID)
            if err != nil {
                log.Printf("Ошибка получения вопроса %d: %v", questionID, err)
                c.JSON(400, gin.H{
                    "error":       "Ошибка получения вопроса",
                    "question_id": questionID,
                })
                return
            }

            // Проверяем правильность ответа
            isCorrect, err := checkAnswerCorrectness2(answers, question)
            if err != nil {
                log.Printf("Ошибка проверки ответа для вопроса %d: %v", questionID, err)
                c.JSON(400, gin.H{
                    "error":   "Ошибка проверки ответа",
                    "details": err.Error(),
                })
                return
            }

            // Начисляем баллы, если ответ правильный
            if isCorrect {
                totalScore += question.Ball
            }

            // Сохраняем ответы
            for _, answer := range answers {
                answer.UserID = userID
                if err := database.CreateAnswerUser(db, &answer); err != nil {
                    log.Printf("Ошибка сохранения ответа для вопроса %d: %v", questionID, err)
                    c.JSON(500, gin.H{
                        "error": "Ошибка сохранения ответа",
                    })
                    return
                }
            }
        }

        // Сохраняем общий результат
        if err := database.SaveUserResult(db, surveyID, userID, totalScore); err != nil {
            log.Printf("Ошибка сохранения результата: %v", err)
            c.JSON(500, gin.H{"error": "Ошибка сохранения результата"})
            return
        }

        // Фиксируем транзакцию
        if err := tx.Commit(); err != nil {
            log.Printf("Ошибка фиксации транзакции: %v", err)
            c.JSON(500, gin.H{"error": "Ошибка фиксации транзакции"})
            return
        }

        c.JSON(200, gin.H{
            "message":     "Ответы успешно сохранены",
            "total_score": totalScore,
        })
    }
}

func checkAnswerCorrectness2(answers []models.AnswerUser, question *models.Question) (bool, error) {
    log.Printf("Checking question %d: IsTest=%v, MultipleAnswers=%v, Answers=%+v, Available Answers=%+v", 
        question.ID, question.IsTest, question.MultipleAnswers, answers, question.Answers)
    
    if question.IsTest {
        if question.MultipleAnswers {
            selectedIDs := make([]int, 0, len(answers))
            for _, answer := range answers {
                if answer.AnswerID <= 0 {
                    err := fmt.Errorf("answer_id должен быть больше 0 для вопроса %d", question.ID)
                    log.Printf("Error: %v", err)
                    return false, err
                }
                selectedIDs = append(selectedIDs, answer.AnswerID)
            }

            if len(selectedIDs) == 0 {
                err := fmt.Errorf("не выбрано ни одного ответа для вопроса %d", question.ID)
                log.Printf("Error: %v", err)
                return false, err
            }

            correctIndices := make([]int, 0)
            for i, ans := range question.Answers {
                if ans.Correct {
                    correctIndices = append(correctIndices, i+1)
                }
            }
            log.Printf("Selected IDs: %v, Correct Indices: %v", selectedIDs, correctIndices)

            if len(selectedIDs) != len(correctIndices) {
                log.Printf("Mismatch in number of answers: selected=%d, correct=%d", 
                    len(selectedIDs), len(correctIndices))
                return false, nil
            }

            for _, selectedID := range selectedIDs {
                if selectedID > len(question.Answers) {
                    err := fmt.Errorf("answer_id %d выходит за пределы доступных ответов (%d) для вопроса %d", 
                        selectedID, len(question.Answers), question.ID)
                    log.Printf("Error: %v", err)
                    return false, err
                }
                found := false
                for _, correctID := range correctIndices {
                    if selectedID == correctID {
                        found = true
                        break
                    }
                }
                if !found {
                    log.Printf("Selected ID %d not in correct indices", selectedID)
                    return false, nil
                }
            }

            for _, correctID := range correctIndices {
                found := false
                for _, selectedID := range selectedIDs {
                    if selectedID == correctID {
                        found = true
                        break
                    }
                }
                if !found {
                    log.Printf("Correct ID %d not selected", correctID)
                    return false, nil
                }
            }

            return true, nil
        } else {
            // Для обычных тестовых вопросов
            if len(answers) != 1 {
                err := fmt.Errorf("ожидается ровно один ответ для вопроса %d, получено %d", 
                    question.ID, len(answers))
                log.Printf("Error: %v", err)
                return false, err
            }
            answer := answers[0]
            if answer.AnswerID <= 0 {
                err := fmt.Errorf("answer_id должен быть больше 0 для вопроса %d", question.ID)
                log.Printf("Error: %v", err)
                return false, err
            }

            if answer.AnswerID > len(question.Answers) {
                err := fmt.Errorf("answer_id %d выходит за пределы доступных ответов (%d) для вопроса %d", 
                    answer.AnswerID, len(question.Answers), question.ID)
                log.Printf("Error: %v", err)
                return false, err
            }

            isCorrect := question.Answers[answer.AnswerID-1].Correct
            log.Printf("Answer ID %d is correct=%v", answer.AnswerID, isCorrect)
            return isCorrect, nil
        }
    } else {
        // Для текстовых вопросов
        if len(answers) != 1 {
            err := fmt.Errorf("ожидается ровно один ответ для текстового вопроса %d, получено %d", 
                question.ID, len(answers))
            log.Printf("Error: %v", err)
            return false, err
        }
        answer := answers[0]
        if answer.AnswerText == "" {
            err := fmt.Errorf("текстовый ответ пуст для вопроса %d", question.ID)
            log.Printf("Error: %v", err)
            return false, err
        }
        isCorrect := strings.TrimSpace(strings.ToLower(answer.AnswerText)) == 
            strings.TrimSpace(strings.ToLower(question.CorrectAnswer))
        log.Printf("Text answer correct=%v", isCorrect)
        return isCorrect, nil
    }
}

func getUserSimpleSurveyAnswers(db *sql.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        surveyIDStr := c.Param("id")
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
        creatorID, err := utils.GetUserIDFromToken(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            return
        }

        surveyIDStr := c.Param("id")
        surveyID, err := strconv.Atoi(surveyIDStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID опроса"})
            return
        }

        var requestBody struct {
            UserID int `json:"user_id"`
        }
        if err := c.ShouldBindJSON(&requestBody); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Требуется user_id в теле запроса"})
            return
        }

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

         if requestBody.UserID == 0 {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный user_id"})
            return
        }

        // Добавьте логирование перед запросом
        log.Printf("Запрос ответов для user_id: %d, survey_id: %d", requestBody.UserID, surveyID)

        // Получаем ответы пользователя
        answers, err := database.GetUserSimpleSurveyAnswers(db, requestBody.UserID, surveyID)
         if err != nil {
            log.Printf("Ошибка при получении ответов: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Не удалось получить ответы",
                "details": err.Error(), // Добавляем детали ошибки
            })
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "survey_id": surveyID,
            "user_id":   requestBody.UserID,
            "answers":   answers,
        })
    }
}
