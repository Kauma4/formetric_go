package database

import (
	"database/sql"
	//"encoding/json"
   // "fmt"
   // "log"
    "errors"
  //  "my-auth-app/internal/models"
	// "golang.org/x/crypto/bcrypt"  
    _ "github.com/lib/pq" 
)

func Connect() (*sql.DB, error) {
    connStr := "user=admin dbname=db_formetric password=root sslmode=disable"
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }

    err = db.Ping() 
    if err != nil {
        return nil, err
    }

    err = createTables(db)
    if err != nil {
        return nil, err
    }

    return db, nil
}

func createTables(db *sql.DB) error {
    createTablesQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        email TEXT,
        full_name TEXT,
        avatar_url TEXT,
        phone_number TEXT,
        date_of_birth DATE,
        location TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        deleted_at TIMESTAMP DEFAULT NULL
    );

    CREATE TABLE IF NOT EXISTS surveys (
        id SERIAL PRIMARY KEY,
        is_private BOOLEAN NOT NULL DEFAULT FALSE,
        title TEXT NOT NULL,
        description TEXT,
        created_by INTEGER NOT NULL,
        max_ball INT DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS questions (
        id SERIAL PRIMARY KEY,
        survey_id INTEGER NOT NULL,
        is_required BOOLEAN NOT NULL DEFAULT FALSE,
        is_test BOOLEAN NOT NULL DEFAULT FALSE,
        correct_answer TEXT,
        question_text TEXT NOT NULL,
        ball INTEGER NOT NULL DEFAULT 0,
        answers JSONB NOT NULL DEFAULT '[]'::JSONB,  -- Храним ответы в JSONB
        FOREIGN KEY (survey_id) REFERENCES surveys(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS answers_users (
        id SERIAL PRIMARY KEY,
        user_id INTEGER NOT NULL,
        question_id INTEGER NOT NULL,
        answer_id INTEGER,
        answer_user TEXT,
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
        FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS results (
    id SERIAL PRIMARY KEY,
    survey_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    total_ball INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (survey_id) REFERENCES surveys(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT unique_survey_user UNIQUE (survey_id, user_id)
    );
    `

    _, err := db.Exec(createTablesQuery)
    if err != nil {
        return errors.New("Ошибка создания таблиц: " + err.Error())
    }

    return nil
}

/*
func CreateSurvey(db *sql.DB, survey *models.Survey) error {
    var surveyID int
    err := db.QueryRow(
        "INSERT INTO surveys (title, description, created_by) VALUES ($1, $2, $3) RETURNING id",
        survey.Title, survey.Description, survey.CreatedBy,
    ).Scan(&surveyID)
    if err != nil {
        return err
    }

    survey.ID = surveyID // Сохраняем ID в структуру
    return nil
}

func GetSurveys(db *sql.DB) ([]models.Survey, error) {
    rows, err := db.Query("SELECT id, title, description, created_by FROM surveys")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var surveys []models.Survey
    for rows.Next() {
        var survey models.Survey
        if err := rows.Scan(&survey.ID, &survey.Title, &survey.Description, &survey.CreatedBy); err != nil {
            return nil, err
        }
        surveys = append(surveys, survey)
    }

    return surveys, nil
}

func GetSurveyOwnerID(db *sql.DB, surveyID int) (int, error) {
    var surveyOwnerID int
    err := db.QueryRow("SELECT created_by FROM surveys WHERE id = $1", surveyID).Scan(&surveyOwnerID)
    if err != nil {
        if err == sql.ErrNoRows {
            return 0, errors.New("опрос не найден")
        }
        return 0, err
    }
    return surveyOwnerID, nil
}

func CreateQuestion(db *sql.DB, question *models.Question) error {
    // Логирование перед выполнением запроса
    log.Printf("Creating question in DB: Text='%s', Ball=%d, IsTest=%v", 
        question.QuestionText, question.Ball, question.IsTest)

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

    // Логируем SQL-запрос
    log.Printf("Executing SQL with params: survey_id=%d, text='%s', ball=%d", 
        question.SurveyID, question.QuestionText, question.Ball)

    err = db.QueryRow(`
        INSERT INTO questions 
        (survey_id, question_text, is_required, is_test, correct_answer, ball, answers) 
        VALUES ($1, $2, $3, $4, $5, $6, $7) 
        RETURNING id`,
        question.SurveyID,
        question.QuestionText,
        question.IsRequired,
        question.IsTest,
        question.CorrectAnswer,
        question.Ball,
        answersJSON,
    ).Scan(&question.ID)

    if err != nil {
        log.Printf("DB query error: %v", err)
        return fmt.Errorf("ошибка создания вопроса: %v", err)
    }
    
    log.Printf("Question created with ID: %d", question.ID)
    return nil
}

func GetQuestions(db *sql.DB, surveyID int) ([]models.Question, error) {
    rows, err := db.Query(`
        SELECT id, survey_id, question_text, ball, answers 
        FROM questions 
        WHERE survey_id = $1
    `, surveyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var questions []models.Question
    for rows.Next() {
        var q models.Question
        var answersJSON []byte
        
        err := rows.Scan(&q.ID, &q.SurveyID, &q.QuestionText, &q.Ball, &answersJSON)
        if err != nil {
            return nil, err
        }
        
        // Десериализация JSONB-поля answers
        if err := json.Unmarshal(answersJSON, &q.Answers); err != nil {
            return nil, err
        }
        
        questions = append(questions, q)
    }
    
    return questions, nil
}

func CreateAnswerUser(db *sql.DB, answerUser *models.AnswerUser) error {
    _, err := db.Exec(
        "INSERT INTO answers_users (user_id, question_id, answer_id, answer_user) VALUES ($1, $2, $3, $4)",
        answerUser.UserID, answerUser.QuestionID, answerUser.AnswerID, answerUser.AnswerText,
    )
    if err != nil {
        log.Println("Ошибка при сохранении ответа пользователя:", err)
        return err
    }
    
    return nil
}

func CreateUser(db *sql.DB, user *models.User) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        return errors.New("failed to hash password")
    }

    _, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, string(hashedPassword))
    if err != nil {
        return errors.New("failed to register user")
    }

    return nil
}

func GetUserByUsername(db *sql.DB, username string) (*models.User, error) {
    var user models.User
    err := db.QueryRow("SELECT id, username, password FROM users WHERE username = $1", username).Scan(&user.ID, &user.Username, &user.Password)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, errors.New("database error")
    }

    return &user, nil
}

// GetUserByID возвращает пользователя без чувствительных данных
func GetUserByID(db *sql.DB, userID int) (*models.User, error) {
    var user models.User
    err := db.QueryRow("SELECT id, username, password, email, full_name, avatar_url, phone_number, date_of_birth, location  FROM users WHERE id = $1", userID).Scan(
        &user.ID, &user.Username, &user.Password, &user.Email, &user.FullName, &user.AvatarURL, &user.PhoneNumber, &user.DateOfBirth, &user.Location,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}
// Обновление данных пользователя
func UpdateUser(db *sql.DB, user *models.User) error {
    _, err := db.Exec(`
        UPDATE users 
        SET 
            username = $1, 
            password = $2, 
            email = $3, 
            full_name = $4, 
            avatar_url = $5, 
            phone_number = $6, 
            date_of_birth = $7, 
            location = $8
        WHERE id = $9
    `, user.Username, user.Password, user.Email, user.FullName, user.AvatarURL, user.PhoneNumber, user.DateOfBirth, user.Location, user.ID)
    
    return err
}

// UpdateUserResult обновляет результаты пользователя
func UpdateUserResult(db *sql.DB, surveyID, userID, ballToAdd int) error {
    _, err := db.Exec(`
        INSERT INTO results (survey_id, user_id, total_ball)
        VALUES ($1, $2, $3)
        ON CONFLICT (survey_id, user_id) 
        DO UPDATE SET total_ball = results.total_ball + $3`,
        surveyID, userID, ballToAdd)
    
    return err
}

// Получаем survey_id по question_id
func GetSurveyIDByQuestionID(db *sql.DB, questionID int) (int, error) {
    var surveyID int
    query := "SELECT survey_id FROM questions WHERE id = $1"  // SQL запрос

    err := db.QueryRow(query, questionID).Scan(&surveyID)  // Выполняем запрос и сканируем результат в переменную
    if err != nil {
        if err == sql.ErrNoRows {
            return 0, fmt.Errorf("вопрос с таким ID не найден")
        }
        return 0, fmt.Errorf("ошибка при получении survey_id: %v", err)
    }
    return surveyID, nil
}

// GetQuestionWithAnswers получает вопрос вместе с вариантами ответов
func GetQuestionWithAnswers(db *sql.DB, questionID int) (*models.Question, error) {
    var question models.Question
    var answersJSON []byte

    err := db.QueryRow(`
        SELECT id, survey_id, question_text, is_required, is_test, 
               correct_answer, ball, answers
        FROM questions 
        WHERE id = $1`, questionID).Scan(
        &question.ID, &question.SurveyID, &question.QuestionText,
        &question.IsRequired, &question.IsTest, &question.CorrectAnswer,
        &question.Ball, &answersJSON)

    if err != nil {
        return nil, err
    }

    // Десериализуем JSON с вариантами ответов
    if err := json.Unmarshal(answersJSON, &question.Answers); err != nil {
        return nil, err
    }

    return &question, nil
}

// DeleteSurvey удаляет опрос и все связанные вопросы (каскадное удаление)
func DeleteSurvey(db *sql.DB, surveyID, userID int) error {
    // Проверяем, что опрос принадлежит пользователю
    ownerID, err := GetSurveyOwnerID(db, surveyID)
    if err != nil {
        return err
    }
    if ownerID != userID {
        return errors.New("недостаточно прав для удаления опроса")
    }

    _, err = db.Exec("DELETE FROM surveys WHERE id = $1", surveyID)
    return err
}

// DeleteQuestion удаляет вопрос
func DeleteQuestion(db *sql.DB, questionID, userID int) error {
    // Проверяем права через survey_owner
    var surveyID int
    err := db.QueryRow("SELECT survey_id FROM questions WHERE id = $1", questionID).Scan(&surveyID)
    if err != nil {
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



func GetUserSimpleSurveyAnswers(db *sql.DB, userID, surveyID int) ([]models.UserAnswerSimple, error) {
    query := `
        SELECT 
            q.question_text,
            q.answers,
            q.correct_answer,
            q.is_test,
            au.answer_id,
            au.answer_user
        FROM 
            answers_users au
        JOIN 
            questions q ON au.question_id = q.id
        WHERE 
            au.user_id = $1 AND q.survey_id = $2
    `
    
    rows, err := db.Query(query, userID, surveyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var answers []models.UserAnswerSimple
    for rows.Next() {
        var answer models.UserAnswerSimple
        var answersJSON []byte
        var isTest bool
        var answerID int
        var answerText sql.NullString
        
        err := rows.Scan(
            &answer.QuestionText,
            &answersJSON,
            &answer.CorrectAnswer,
            &isTest,
            &answerID,
            &answerText,
        )
        if err != nil {
            return nil, err
        }

        // Обработка ответа пользователя
        if isTest {
            // Для тестовых вопросов
            var questionAnswers []models.Answer
            if err := json.Unmarshal(answersJSON, &questionAnswers); err != nil {
                return nil, err
            }
            
            if answerID > 0 && answerID <= len(questionAnswers) {
                answer.UserAnswer = questionAnswers[answerID-1].Text
                isCorrect := questionAnswers[answerID-1].Correct
                answer.IsCorrect = &isCorrect
                
                // Находим правильный ответ для тестового вопроса
                for _, a := range questionAnswers {
                    if a.Correct {
                        answer.CorrectAnswer = a.Text
                        break
                    }
                }
            }
        } else {
            // Для текстовых вопросов
            if answerText.Valid {
                answer.UserAnswer = answerText.String
            }
            // correct_answer уже заполнен из БД
        }

        answers = append(answers, answer)
    }

    return answers, nil
}

// UserExists проверяет существование пользователя
func UserExists(db *sql.DB, userID int) (bool, error) {
    var exists bool
    err := db.QueryRow(
        "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND deleted_at IS NULL)", 
        userID,
    ).Scan(&exists)
    return exists, err
}

// SoftDeleteUser выполняет "мягкое" удаление пользователя
func SoftDeleteUser(tx *sql.Tx, userID int) (int64, error) {
    query := `
        UPDATE users 
        SET 
            deleted_at = NOW(),
            email = NULL,
            phone_number = NULL,
            username = CONCAT(username, '_deleted_', EXTRACT(EPOCH FROM NOW()))
        WHERE id = $1 AND deleted_at IS NULL
    `
    result, err := tx.Exec(query, userID)
    if err != nil {
        return 0, fmt.Errorf("ошибка soft delete: %v", err)
    }
    return result.RowsAffected()
}

*/