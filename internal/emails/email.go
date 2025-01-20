package email

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
	"reg/internal/config"
	"reg/internal/model"
	"strings"
)

var (
	smtpHost = os.Getenv("SMTP_HOST")
	smtpPort = os.Getenv("SMTP_PORT")
	smtpUser = os.Getenv("SMTP_USER")
)

func SendEmail(to string, cc []string, subject string, body []byte, replyto string) (bool, error) {
	fromName := "E-Summit x E-Cell IIT Hyderabad"
	from := smtpUser
	// Setup headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", fromName, from)
	if replyto != "" {
		headers["Reply-To"] = replyto
	}
	headers["To"] = to
	if len(cc) > 0 {
		headers["Cc"] = strings.Join(cc, ",")
	}
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"utf-8\""

	// Setup message
	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.Write(body)

	// Recipients
	recipients := append([]string{to}, cc...)

	// Sending email
	err := smtp.SendMail(smtpHost+":"+smtpPort, config.SmtpAuth, from, recipients, msg.Bytes())
	if err != nil {
		log.Printf("Failed to send email: %v\n", err)
		config.LogEmails(to, cc, subject, false)
		return false, err
	}
	config.LogEmails(to, cc, subject, true)
	log.Println("Email sent successfully!")
	return true, nil
}

func LoadOtpVerificationsTemplate(otp string) ([]byte, error) {
	filePath := "templates/otp.html"
	template, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Replace placeholders in the template with actual data
	htmlContent := string(template)
	htmlContent = strings.ReplaceAll(htmlContent, "[OTP]", otp)

	return []byte(htmlContent), nil
}

func LoadRegistrationTemplate(data model.RegistrationRequest) ([]byte, error) {
	filePath := "templates/register.html"
	tmplContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("registration").Parse(string(tmplContent))
	if err != nil {
		return nil, err
	}

	var renderedContent bytes.Buffer

	// Execute the template with the registration data and write to the buffer
	err = tmpl.Execute(&renderedContent, data)
	if err != nil {
		return nil, err
	}

	// Return the rendered HTML content as a byte slice
	return renderedContent.Bytes(), nil
}

func LoadSignUpVerificationTemplate(name string) ([]byte, error) {
	filePath := "templates/signup.html"
	tmplContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("signup").Parse(string(tmplContent))
	if err != nil {
		return nil, err
	}

	var renderedContent bytes.Buffer

	// Execute the template with the registration data and write to the buffer
	err = tmpl.Execute(&renderedContent, name)
	if err != nil {
		return nil, err
	}

	// Return the rendered HTML content as a byte slice
	return renderedContent.Bytes(), nil
}
