package database

import (
	"database/sql"
    "errors"
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
        multipleAnswers BOOLEAN NOT NULL DEFAULT FALSE,
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