package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	emails := r.MultipartForm.Value["emails"]
	fmt.Println(emails)
	fmt.Println(fileHeader.Filename)

	// if len(file) != 1 {
	// 	http.Error(w, "Attach only one file", http.StatusBadRequest)
	// 	return
	// }
	if len(emails) == 0 {
		http.Error(w, "No emails are sent", http.StatusBadRequest)
		return
	}

	temp, err := os.CreateTemp("", "file-*"+filepath.Ext(fileHeader.Filename))
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(temp.Name())

	f, err := os.Open(temp.Name())
	if err != nil {
		http.Error(w, "Failed to open temp file", http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(f, file)
	if err != nil {
		http.Error(w, "Failed to copy file", http.StatusInternalServerError)
		return
	}
	temp.Close()

	for _, email := range emails {
		err = sendEmail(email, temp.Name())
		if err != nil {
			http.Error(w, "Failed to send file", http.StatusInternalServerError)
			return
		}
	}

}
