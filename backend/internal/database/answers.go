package database

import (
    "strings"
    "database/sql"
    "encoding/json"
    "my-auth-app/internal/models"
)

func CreateAnswerUser(db *sql.DB, answer *models.AnswerUser) error {
    _, err := db.Exec(
        "INSERT INTO answers_users (user_id, question_id, answer_id, answer_user) VALUES ($1, $2, $3, $4)",
        answer.UserID, answer.QuestionID, answer.AnswerID, answer.AnswerText,
    )
    return err
}
/*
func GetUserSimpleSurveyAnswers(db *sql.DB, userID, surveyID int) ([]models.UserAnswerSimple, error) {
    rows, err := db.Query(
        `SELECT q.question_text, q.answers, q.correct_answer, q.is_test, 
        au.answer_id, au.answer_user 
        FROM answers_users au
        JOIN questions q ON au.question_id = q.id
        WHERE au.user_id = $1 AND q.survey_id = $2`,
        userID, surveyID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var answers []models.UserAnswerSimple
    for rows.Next() {
        var a models.UserAnswerSimple
        var answersJSON []byte
        var isTest bool
        var answerID int
        var answerText sql.NullString
        var dbCorrectAnswer sql.NullString
        
        if err := rows.Scan(
            &a.QuestionText, &answersJSON, &dbCorrectAnswer, 
            &isTest, &answerID, &answerText,
        ); err != nil {
            return nil, err
        }

        a.CorrectAnswer = dbCorrectAnswer.String

        if isTest {
            var qAnswers []models.Answer
            if err := json.Unmarshal(answersJSON, &qAnswers); err != nil {
                return nil, err
            }
            
            var correctAnswers []string
            for _, ans := range qAnswers {
                if ans.Correct {
                    correctAnswers = append(correctAnswers, ans.Text)
                }
            }
            a.CorrectAnswer = strings.Join(correctAnswers, ", ")

            if answerID > 0 && answerID <= len(qAnswers) {
                a.UserAnswer = qAnswers[answerID-1].Text
                isCorrect := qAnswers[answerID-1].Correct
                a.IsCorrect = &isCorrect
            }
        } else if answerText.Valid {
            a.UserAnswer = answerText.String
        }
        
        answers = append(answers, a)
    }
    return answers, nil
}
*/

func GetUserSimpleSurveyAnswers(db *sql.DB, userID, surveyID int) ([]models.UserAnswerSimple, error) {
    // Добавляем question_id в запрос
    rows, err := db.Query(
         `SELECT q.id AS question_id, q.question_text, q.answers, 
    q.correct_answer, q.is_test, 
    au.answer_id, au.answer_user 
    FROM answers_users au
    JOIN questions q ON au.question_id = q.id
    WHERE au.user_id = $1 AND q.survey_id = $2
    ORDER BY q.id`, 
    userID, surveyID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    answersMap := make(map[int]*models.UserAnswerSimple)
   for rows.Next() {
    var (
        questionID      int
        a               models.UserAnswerSimple
        answersJSON     []byte
        isTest         bool
        answerID       int
        answerText     sql.NullString
        dbCorrectAnswer sql.NullString
    )

    // Убрали multiple_answers из сканирования
    if err := rows.Scan(
        &questionID, &a.QuestionText, &answersJSON, 
        &dbCorrectAnswer, &isTest,
        &answerID, &answerText,
    ); err != nil {
        return nil, err
    }

    a.IsTest = isTest // Добавляем установку IsTest

        // Если вопрос еще не в мапе, создаем новую запись
        if _, exists := answersMap[questionID]; !exists {
            a.CorrectAnswer = dbCorrectAnswer.String
            a.UserAnswers = []string{}     // Новое поле для множества ответов
            a.IsCorrectMultiple = []bool{} // Новое поле для отметок правильности
            
            if isTest {
                var qAnswers []models.Answer
                if err := json.Unmarshal(answersJSON, &qAnswers); err != nil {
                    return nil, err
                }
                
                var correctAnswers []string
                for _, ans := range qAnswers {
                    if ans.Correct {
                        correctAnswers = append(correctAnswers, ans.Text)
                    }
                }
                a.CorrectAnswer = strings.Join(correctAnswers, ", ")
            }
            
            answersMap[questionID] = &a
        }

        currentAnswer := answersMap[questionID]
        
        // Обрабатываем ответ
        if isTest {
            var qAnswers []models.Answer
            if err := json.Unmarshal(answersJSON, &qAnswers); err != nil {
                return nil, err
            }

            if answerID > 0 && answerID <= len(qAnswers) {
                // Для множественных ответов добавляем в слайсы
                currentAnswer.UserAnswers = append(currentAnswer.UserAnswers, qAnswers[answerID-1].Text)
                currentAnswer.IsCorrectMultiple = append(currentAnswer.IsCorrectMultiple, qAnswers[answerID-1].Correct)
            }
        } else if answerText.Valid {
            currentAnswer.UserAnswer = answerText.String
        }
    }

    // Конвертируем мапу в слайс
    answers := make([]models.UserAnswerSimple, 0, len(answersMap))
    for _, a := range answersMap {
        // Для тестовых вопросов с множественными ответами объединяем ответы
        if a.IsTest && len(a.UserAnswers) > 0 {
            a.UserAnswer = strings.Join(a.UserAnswers, ", ")
        }
        answers = append(answers, *a)
    }

    return answers, nil
}

func UpdateUserResult(db *sql.DB, surveyID, userID, ballToAdd int) error {
    _, err := db.Exec(
        `INSERT INTO results (survey_id, user_id, total_ball) 
        VALUES ($1, $2, $3)
        ON CONFLICT (survey_id, user_id) 
        DO UPDATE SET total_ball = results.total_ball + $3`,
        surveyID, userID, ballToAdd,
    )
    return err
}

func GetSurveyIDByQuestionID(db *sql.DB, questionID int) (int, error) {
    var surveyID int
    err := db.QueryRow("SELECT survey_id FROM questions WHERE id = $1", questionID).Scan(&surveyID)
    return surveyID, err
}