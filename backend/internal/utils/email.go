package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

type EmailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func getEmailConfig() EmailConfig {
	return EmailConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
	}
}

func SendEmail(to, subject, body string) error {
	config := getEmailConfig()
	if config.Host == "" || config.Port == "" || config.Username == "" || config.Password == "" || config.From == "" {
		return fmt.Errorf("SMTP configuration is incomplete")
	}
	auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	msg := []byte(fmt.Sprintf(
		"To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s\r\n",
		to, subject, body,
	))
	return smtp.SendMail(addr, auth, config.From, []string{to}, msg)
}

func SendVerificationEmail(to, username string) error {
	subject := "Подтверждение регистрации"
	body := fmt.Sprintf(
		"Здравствуйте, %s!\n\nВаш аккаунт успешно зарегистрирован.\nС уважением,\nКоманда приложения",
		username,
	)
	return SendEmail(to, subject, body)
}

func SendTokenEmail(to, username, token string) error {
	subject := "Ваш токен для входа"
	body := fmt.Sprintf(
		"Здравствуйте, %s!\n\nВаш токен для доступа:\n%s\n\nС уважением,\nКоманда приложения",
		username, token,
	)
	return SendEmail(to, subject, body)
}