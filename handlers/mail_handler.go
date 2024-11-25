package handlers

import (
	"archiver/internal/utils"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/gomail.v2"
)

func sendEmail(to string, attachment string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("FROM_EMAIL"))
	m.SetHeader("To", to)
	m.SetHeader("Subject", "A new file")
	m.SetBody("text/html", "<h3>Knock Knock. A new file has been sent to you</h3>")
	m.Attach(attachment)

	d := gomail.NewDialer(os.Getenv("HOST"), 465, os.Getenv("FROM_EMAIL"), os.Getenv("FROM_EMAIL_PASSWORD"))
	return d.DialAndSend(m)
}

func SendEmail(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(0) // 30 mb
	if err != nil {
		utils.SendError(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		utils.SendError(w, "Failed to retrieve file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	mimeType := GetMimeType(fileHeader.Filename)
	if mimeType != OPEN_XML && mimeType != PDF {
		utils.SendError(w, "Not appropriate file type", http.StatusBadRequest)
		return
	}

	emails := r.FormValue("emails")

	emailsList := strings.Split(emails, ",")
	for i, email := range emailsList {
		email = strings.TrimSpace(email)
		if !utils.IsValidEmail(email) {
			utils.SendError(w, fmt.Sprintf("Invalid email address: %s", email), http.StatusBadRequest)
			return
		}
		emailsList[i] = email
	}

	if len(emails) == 0 {
		utils.SendError(w, "No emails are sent", http.StatusBadRequest)
		return
	}

	temp, err := os.CreateTemp("", "file-*"+filepath.Ext(fileHeader.Filename))
	if err != nil {
		utils.SendError(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(temp.Name())

	_, err = io.Copy(temp, file)
	if err != nil {
		utils.SendError(w, "Failed to copy file", http.StatusInternalServerError)
		return
	}
	temp.Close()

	var wg sync.WaitGroup
	errors := make(chan error, len(emails))

	for _, email := range emailsList {
		wg.Add(1)
		go func(email string) {
			defer wg.Done()
			if err := sendEmail(email, temp.Name()); err != nil {
				errors <- fmt.Errorf("failed to send email to %s: %v", email, err)
			}
		}(email)
	}

	wg.Wait()
	close(errors)

	// Check if any errors occurred
	var failedEmails []string
	for err := range errors {
		slog.Error(err.Error())
		failedEmails = append(failedEmails, err.Error())
	}

	if len(failedEmails) > 0 {
		utils.SendError(w, fmt.Sprintf("Some emails failed: %v", failedEmails), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("file has been sent"))
}
