package handlers

import (
	//"net/http"
    "testing"
    "github.com/stretchr/testify/assert"
    "my-auth-app/internal/models"
)

func TestCheckAnswerCorrectness2(t *testing.T) {
    question := &models.Question{
        ID:       1,
        IsTest:   true,
        Ball:     10,
        Answers: []models.Answer{
            {ID: 1, Correct: true},
            {ID: 2, Correct: false},
        },
    }

    t.Run("Correct answer", func(t *testing.T) {
        answers := []models.AnswerUser{{AnswerID: 1}}
        isCorrect, err := checkAnswerCorrectness2(answers, question)
        assert.NoError(t, err)
        assert.True(t, isCorrect)
    })

    t.Run("Incorrect answer", func(t *testing.T) {
        answers := []models.AnswerUser{{AnswerID: 2}}
        isCorrect, err := checkAnswerCorrectness2(answers, question)
        assert.NoError(t, err)
        assert.False(t, isCorrect)
    })
}