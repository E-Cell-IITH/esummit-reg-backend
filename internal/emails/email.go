package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"image/png"
	"log"
	"net/smtp"
	"os"
	"reg/internal/config"
	"reg/internal/model"
	"strings"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
)

var (
	smtpHost = os.Getenv("SMTP_HOST")
	smtpPort = os.Getenv("SMTP_PORT")
	smtpUser = os.Getenv("SMTP_USER")
)

func SendPASSEmail(to string, cc []string, subject string, body []byte, replyto string, qrCodeId string) (bool, error) {
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
	headers["Content-Type"] = "multipart/related; boundary=boundary42"

	// Setup message
	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n--boundary42\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
	msg.Write(body)

	// Attach image.png
	imagePath := "templates/image.png"
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		log.Printf("Failed to read image file: %v\n", err)
		return false, err
	}
	msg.WriteString("\r\n--boundary42\r\n")
	msg.WriteString("Content-Type: image/png\r\n")
	msg.WriteString("Content-Transfer-Encoding: base64\r\n")
	msg.WriteString("Content-ID: <image.png>\r\n\r\n")
	msg.WriteString(base64.StdEncoding.EncodeToString(imageData))

	// Attach QR code
	qrCodeData, err := generateBarcodeBase64(qrCodeId)
	if err != nil {
		log.Printf("Failed to generate QR code: %v\n", err)
		return false, err
	}
	msg.WriteString("\r\n--boundary42\r\n")
	msg.WriteString("Content-Type: image/png\r\n")
	msg.WriteString("Content-Transfer-Encoding: base64\r\n")
	msg.WriteString("Content-ID: <qrcode.png>\r\n\r\n")
	msg.WriteString(qrCodeData)

	msg.WriteString("\r\n--boundary42--")

	// Recipients
	recipients := append([]string{to}, cc...)

	// Sending email
	err = smtp.SendMail(smtpHost+":"+smtpPort, config.SmtpAuth, from, recipients, msg.Bytes())
	if err != nil {
		log.Printf("Failed to send email: %v\n", err)
		config.LogEmails(to, cc, subject, false)
		return false, err
	}
	config.LogEmails(to, cc, subject, true)
	return true, nil
}

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

	htmlContent := string(tmplContent)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.name}}", name)

	return []byte(htmlContent), nil
}

func LoadPurchasedTicketTemplate(name, title, amount string) ([]byte, error) {
	filePath := "templates/tickets_purchased.html"
	tmplContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	htmlContent := string(tmplContent)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.Name}}", name)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.TicketType}}", title)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.Price}}", amount)

	return []byte(htmlContent), nil
}

func LoadPendingTemplate(name, txnId, amount string) ([]byte, error) {
	filePath := "templates/pending.html"
	tmplContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	htmlContent := string(tmplContent)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.Name}}", name)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.TransactionID}}", txnId)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.Amount}}", amount)

	return []byte(htmlContent), nil
}

func generateBarcodeBase64(data string) (string, error) {
	// Generate a Code128 barcode
	barCode, err := code128.Encode(data)
	if err != nil {
		return "", err
	}

	// Scale the barcode to a desired size
	scaledBarCode, err := barcode.Scale(barCode, 1024, 200) // Width x Height
	if err != nil {
		return "", err
	}

	// Encode the barcode to PNG
	var buf bytes.Buffer
	err = png.Encode(&buf, scaledBarCode)
	if err != nil {
		return "", err
	}

	// Convert to Base64
	barBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return barBase64, nil
}

func LoadPassEmailTemplate(name, pass, id string) ([]byte, error) {
	filePath := "templates/pass.html"
	tmplContent, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading template file: %v\n", err)
		return nil, err
	}

	htmlContent := string(tmplContent)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.Name}}", name)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.Pass}}", pass)

	// data, err := generateBarcodeBase64(id)
	// if err != nil {
	// 	fmt.Printf("Failed to generate QR code: %v for id %s\n", err, id)
	// 	return nil, err
	// }
	// htmlContent = strings.ReplaceAll(htmlContent, "{{.QRDATA}}", "data:image/png;base64,"+data)
	htmlContent = strings.ReplaceAll(htmlContent, "{{.ALT}}", id)

	// imageBase64, err := loadImageBase64("templates/image.png")
	// if err != nil {
	// 	fmt.Printf("Failed to load image: %v\n", err)
	// 	return nil, err
	// }
	// htmlContent = strings.ReplaceAll(htmlContent, "{{.ImageBase64}}", imageBase64)

	return []byte(htmlContent), nil
}

// func loadImageBase64(filePath string) (string, error) {
// 	imageData, err := os.ReadFile(filePath)
// 	if err != nil {
// 		fmt.Printf("Error reading image file: %v\n", err)
// 		return "", err
// 	}
// 	base64String := base64.StdEncoding.EncodeToString(imageData)
// 	fmt.Printf("Base64 Image String: %s\n", base64String[:50]) // Print first 50 characters for debugging
// 	return base64String, nil
// }
