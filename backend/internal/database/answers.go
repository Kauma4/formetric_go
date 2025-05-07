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
        
        if err := rows.Scan(
            &a.QuestionText, &answersJSON, &a.CorrectAnswer, 
            &isTest, &answerID, &answerText,
        ); err != nil {
            return nil, err
        }

        if isTest {
            var qAnswers []models.Answer
            if err := json.Unmarshal(answersJSON, &qAnswers); err != nil {
                return nil, err
            }
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
            
            // Собираем все правильные ответы
            var correctAnswers []string
            for _, ans := range qAnswers {
                if ans.Correct {
                    correctAnswers = append(correctAnswers, ans.Text)
                }
            }
            a.CorrectAnswer = strings.Join(correctAnswers, ", ")

            // Обработка ответа пользователя
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