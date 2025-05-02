package database

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "errors"
    "my-auth-app/internal/models"
)

func CreateQuestion(db *sql.DB, question *models.Question) error {
    var answersJSON []byte
    var err error
    
    if question.IsTest {
        answersJSON, err = json.Marshal(question.Answers)
        if err != nil {
            log.Printf("Answers marshaling error: %v", err)
            return fmt.Errorf("ошибка сериализации ответов: %v", err)
        }
    } else {
        answersJSON = []byte("[]")
    }

    err = db.QueryRow(
        `INSERT INTO questions (survey_id, question_text, is_required, is_test, 
        correct_answer, ball, answers) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
        question.SurveyID, question.QuestionText, question.IsRequired,
        question.IsTest, question.CorrectAnswer, question.Ball, answersJSON,
    ).Scan(&question.ID)

    return err
}

func GetQuestions(db *sql.DB, surveyID int) ([]models.Question, error) {
    rows, err := db.Query(
        `SELECT 
            id, 
            survey_id, 
            question_text, 
            ball, 
            is_required, 
            is_test, 
            correct_answer, 
            answers 
         FROM questions 
         WHERE survey_id = $1`,
        surveyID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var questions []models.Question
    for rows.Next() {
        var q models.Question
        var answersJSON []byte
        
        // 2. Добавьте все поля в Scan
        err := rows.Scan(
            &q.ID,
            &q.SurveyID,
            &q.QuestionText,
            &q.Ball,
            &q.IsRequired,
            &q.IsTest,
            &q.CorrectAnswer,
            &answersJSON,
        )
        
        if err != nil {
            return nil, err
        }
        
        // 3. Добавьте обработку answers
        if len(answersJSON) > 0 {
            if err := json.Unmarshal(answersJSON, &q.Answers); err != nil {
                return nil, fmt.Errorf("failed to unmarshal answers: %v", err)
            }
        }
        
        questions = append(questions, q)
    }
    
    if err = rows.Err(); err != nil {
        return nil, err
    }
    
    return questions, nil
}

func GetQuestionWithAnswers(db *sql.DB, questionID int) (*models.Question, error) {
    var q models.Question
    var answersJSON []byte
    err := db.QueryRow(
        `SELECT id, survey_id, question_text, is_required, is_test, 
        correct_answer, ball, answers FROM questions WHERE id = $1`,
        questionID,
    ).Scan(
        &q.ID, &q.SurveyID, &q.QuestionText, &q.IsRequired,
        &q.IsTest, &q.CorrectAnswer, &q.Ball, &answersJSON,
    )
    if err != nil {
        return nil, err
    }
    if err := json.Unmarshal(answersJSON, &q.Answers); err != nil {
        return nil, err
    }
    return &q, nil
}

// GetQuestionIDsForSurvey возвращает список ID вопросов опроса
func GetQuestionIDsForSurvey(db *sql.DB, surveyID int) ([]int, error) {
    rows, err := db.Query("SELECT id FROM questions WHERE survey_id = $1", surveyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var ids []int
    for rows.Next() {
        var id int
        if err := rows.Scan(&id); err != nil {
            return nil, err
        }
        ids = append(ids, id)
    }
    return ids, nil
}

func DeleteQuestion(db *sql.DB, questionID, userID int) error {
    var surveyID int
    if err := db.QueryRow("SELECT survey_id FROM questions WHERE id = $1", questionID).Scan(&surveyID); err != nil {
        return err
    }
    ownerID, err := GetSurveyOwnerID(db, surveyID)
    if err != nil {
        return err
    }
    if ownerID != userID {
        return errors.New("недостаточно прав для удаления вопроса")
    }
    _, err = db.Exec("DELETE FROM questions WHERE id = $1", questionID)
    return err
}

// DeleteQuestionsExcept удаляет вопросы не входящие в список ID
func DeleteQuestionsExcept(tx *sql.Tx, surveyID int, keepIDs []int) error {
    query := "DELETE FROM questions WHERE survey_id = $1"
    args := []interface{}{surveyID}
    
    if len(keepIDs) > 0 {
        query += " AND id NOT IN ("
        for i, id := range keepIDs {
            if i > 0 {
                query += ","
            }
            query += fmt.Sprintf("$%d", i+2)
            args = append(args, id)
        }
        query += ")"
    }

    _, err := tx.Exec(query, args...)
    return err
}

func UpdateQuestion(db *sql.DB, question *models.Question) error {
    answersJSON, err := json.Marshal(question.Answers)
    if err != nil {
        return fmt.Errorf("ошибка сериализации ответов: %v", err)
    }

    _, err = db.Exec(`
        UPDATE questions SET
            question_text = $1,
            is_required = $2,
            is_test = $3,
            correct_answer = $4,
            ball = $5,
            answers = $6
        WHERE id = $7`,
        question.QuestionText,
        question.IsRequired,
        question.IsTest,
        question.CorrectAnswer,
        question.Ball,
        answersJSON,
        question.ID,
    )
    return err
}