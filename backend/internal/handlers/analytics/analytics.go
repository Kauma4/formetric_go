package analytics

import (
	"bytes"
	"fmt"
	"os"
	"io"
	"log"
	"net/http"
	//"net/url" 
	"strconv"
	"strings"
	"time"
	"database/sql"
	"my-auth-app/internal/database"
	"my-auth-app/internal/models"
	"my-auth-app/internal/jwt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)
// RegisterAnalyticsRoutes регистрирует маршруты для аналитики
func RegisterAnalyticsRoutes(r *gin.Engine, db *sql.DB) {
	authGroup := r.Group("/surveys")
	authGroup.Use(AuthMiddleware())
	{
		authGroup.GET("/:id/analytics", getSurveyAnalytics(db))
		authGroup.GET("/:id/analytics/pdf", getSurveyAnalyticsPDF(db))
	}
}

// getSurveyAnalytics возвращает JSON с аналитикой по опросу
func getSurveyAnalytics(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		surveyID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID опроса"})
			return
		}

		userID, err := getUserIDFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			return
		}

		ownerID, err := database.GetSurveyOwnerID(db, surveyID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if ownerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещён"})
			return
		}

		analytics, err := database.GetSurveyAnalytics(db, surveyID)
		if err != nil {
			log.Printf("Ошибка получения аналитики для опроса %d: %v", surveyID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
			return
		}

		c.JSON(http.StatusOK, analytics)
	}
}

// getSurveyAnalyticsPDF возвращает аналитику в формате PDF
func getSurveyAnalyticsPDF(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		surveyID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID опроса"})
			return
		}

		userID, err := getUserIDFromToken(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			return
		}

		ownerID, err := database.GetSurveyOwnerID(db, surveyID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if ownerID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещён"})
			return
		}

		analytics, err := database.GetSurveyAnalytics(db, surveyID)
		if err != nil {
			log.Printf("Ошибка получения аналитики для опроса %d: %v", surveyID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
			return
		}

		// Генерируем HTML-документ
		log.Printf("Генерация HTML для опроса %d", surveyID)
		htmlContent := generateHTMLDocument(analytics)

		// Сохраняем HTML-код для отладки
		debugFile := fmt.Sprintf("debug_survey_%d.html", surveyID)
		if err := os.WriteFile(debugFile, []byte(htmlContent), 0644); err != nil {
			log.Printf("Ошибка сохранения HTML-файла %s: %v", debugFile, err)
		} else {
			log.Printf("HTML-код сохранён в %s", debugFile)
		}

		// Компилируем HTML в PDF через DocRaptor API
		log.Printf("Отправка запроса к DocRaptor для опроса %d", surveyID)
		pdfContent, err := compileHTMLToPDF(htmlContent)
		if err != nil {
			log.Printf("Ошибка генерации PDF для опроса %d: %v", surveyID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации PDF"})
			return
		}

		log.Printf("PDF успешно сгенерирован для опроса %d", surveyID)
		// Отправляем PDF клиенту
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=survey_%d_analytics.pdf", surveyID))
		c.Data(http.StatusOK, "application/pdf", pdfContent)
	}
}

// compileHTMLToPDF отправляет HTML-код в DocRaptor API и получает PDF
func compileHTMLToPDF(htmlContent string) ([]byte, error) {
	const apiKey = "o12_26jojjnlbbjXF01E" // Замените на ваш API-ключ DocRaptor
	urlStr := "https://api.docraptor.com/docs"
	payload := bytes.NewBufferString(fmt.Sprintf(`{
		"test": true,
		"document_content": %q,
		"type": "pdf",
		"javascript": true,
		"prince_options": {
			"media": "print"
		}
	}`, htmlContent))

	req, err := http.NewRequest("POST", urlStr, payload)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка создания HTTP-запроса")
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(apiKey, "")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка отправки запроса к DocRaptor")
	}
	defer resp.Body.Close()

	// Сохраняем ответ для отладки
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ошибка чтения ответа")
	}
	debugFile := fmt.Sprintf("debug_docraptor_response_%d.txt", time.Now().Unix())
	if err := os.WriteFile(debugFile, body, 0644); err != nil {
		log.Printf("Ошибка сохранения ответа DocRaptor в %s: %v", debugFile, err)
	} else {
		log.Printf("Ответ DocRaptor сохранён в %s", debugFile)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("DocRaptor вернул ошибку: %s, тело: %s", resp.Status, string(body))
	}

	if len(body) < 4 || string(body[:4]) != "%PDF" {
		return nil, errors.Errorf("получен некорректный PDF-файл, первые 100 символов: %s", string(body[:min(100, len(body))]))
	}

	return body, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateHTMLDocument создаёт HTML-документ для аналитики
func generateHTMLDocument(analytics *models.SurveyAnalytics) string {
	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <title>Аналитика опроса: ` + htmlEscape(analytics.Survey.Title) + `</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 2cm; }
        h1 { text-align: center; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #ccc; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .section { margin-top: 20px; }
    </style>
</head>
<body>
    <h1>Аналитика опроса: ` + htmlEscape(analytics.Survey.Title) + `</h1>
    <div class="section">
        <h2>Общая информация</h2>
        <table>
            <tr><th>Название</th><td>` + htmlEscape(analytics.Survey.Title) + `</td></tr>
            <tr><th>Описание</th><td>` + htmlEscape(analytics.Survey.Description) + `</td></tr>
            <tr><th>Создан</th><td>` + analytics.Survey.CreatedAt.Format("02.01.2006") + `</td></tr>
            <tr><th>Максимальный балл</th><td>` + fmt.Sprintf("%d", analytics.Survey.MaxBall) + `</td></tr>
            <tr><th>Количество ответов</th><td>` + fmt.Sprintf("%d", analytics.TotalResponses) + `</td></tr>
            <tr><th>Средний балл</th><td>` + fmt.Sprintf("%.2f", analytics.AverageScore) + `</td></tr>
        </table>
    </div>
    <div class="section">
        <h2>Участники</h2>
        <table>
            <tr><th>Имя пользователя</th><th>Дата</th><th>Баллы</th><th>% от максимума</th></tr>
`)

	for _, p := range analytics.Participants {
		percent := 0.0
		if analytics.Survey.MaxBall > 0 {
			percent = float64(p.TotalBall) / float64(analytics.Survey.MaxBall) * 100
		}
		sb.WriteString(fmt.Sprintf(
			`<tr><td>%s</td><td>%s</td><td>%d</td><td>%.2f%%</td></tr>`,
			htmlEscape(p.Username),
			p.Date.Format("02.01.2006 15:04"),
			p.TotalBall,
			percent,
		))
	}
	sb.WriteString(`
        </table>
    </div>
    <div class="section">
        <h2>Статистика по вопросам</h2>
`)

	for i, qa := range analytics.Questions {
		sb.WriteString(fmt.Sprintf(`
        <h3>Вопрос %d: %s</h3>
        <p><strong>Тип:</strong> %s</p>
        <p><strong>Баллы за вопрос:</strong> %d</p>
`, i+1, htmlEscape(qa.Question.QuestionText), qa.Question.GetQuestionType(), qa.Question.Ball))

		if qa.Question.IsTest {
			// Извлекаем правильный ответ из AnswerStats
			correctAnswer := "Не указан"
			for _, stat := range qa.AnswerStats {
				if stat.IsCorrect {
					correctAnswer = htmlEscape(stat.AnswerText)
					break
				}
			}
			sb.WriteString(fmt.Sprintf(`
        <p><strong>Правильный ответ:</strong> %s</p>
        <p><strong>Процент правильных ответов:</strong> %.2f%%</p>
        <table>
            <tr><th>Вариант ответа</th><th>Количество</th><th>Правильный</th></tr>
`, correctAnswer, qa.CorrectRate))
			for _, stat := range qa.AnswerStats {
				sb.WriteString(fmt.Sprintf(
					`<tr><td>%s</td><td>%d</td><td>%v</td></tr>`,
					htmlEscape(stat.AnswerText),
					stat.Count,
					stat.IsCorrect,
				))
			}
			sb.WriteString(`</table>`)
		} else {
			sb.WriteString(`<p><strong>Текстовые ответы:</strong></p>`)
			if len(qa.TextResponses) > 0 {
				sb.WriteString(`
        <table>
            <tr><th>Пользователь</th><th>Ответ</th></tr>
`)
				for _, resp := range qa.TextResponses {
					sb.WriteString(fmt.Sprintf(
						`<tr><td>%s</td><td>%s</td></tr>`,
						htmlEscape(resp.Username),
						htmlEscape(resp.Answer),
					))
				}
				sb.WriteString(`</table>`)
			} else {
				sb.WriteString(`<p>Нет ответов.</p>`)
			}
		}
	}

	sb.WriteString(`
</body>
</html>
`)

	return sb.String()
}

// htmlEscape экранирует специальные символы для HTML
func htmlEscape(s string) string {
	replacements := map[string]string{
		"&":  "&amp;",
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&#39;",
	}
	for k, v := range replacements {
		s = strings.ReplaceAll(s, k, v)
	}
	return s
}

// getUserIDFromToken извлекает userID из контекста
func getUserIDFromToken(c *gin.Context) (int, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, errors.New("user ID not found in context")
	}
	id, ok := userID.(int)
	if !ok {
		return 0, errors.New("invalid user ID type")
	}
	return id, nil
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Получаем токен из заголовка
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
            return
        }

        // Удаляем префикс "Bearer " если есть
        if strings.HasPrefix(tokenString, "Bearer ") {
            tokenString = strings.TrimPrefix(tokenString, "Bearer ")
        }

        // Извлекаем ID пользователя из токена
        userID, err := jwt.ExtractUserIDFromJWT(tokenString)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Неверный токен"})
            return
        }

        // Сохраняем ID пользователя в контекст
        c.Set("userID", userID)
        c.Next()
    }
}