package handlers_test

import (
	"archiver/handlers"
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendEmail(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/mail/file", nil)
		rr := httptest.NewRecorder()

		handlers.SendEmail(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}

		expected := "Failed to parse multipart form"
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("invalid file type", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		fileWriter, _ := writer.CreateFormFile("file", "test.txt")
		fileWriter.Write([]byte("this is a test file"))

		writer.WriteField("emails", "test@example.com")
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/mail/file", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		handlers.SendEmail(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}

		expected := "Not appropriate file type"
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("invalid email addresses", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// Add a dummy file
		fileWriter, _ := writer.CreateFormFile("file", "test.docx")
		fileWriter.Write([]byte("this is a test file"))

		// Add invalid emails
		writer.WriteField("emails", "invalid-email")
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/mail/file", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		handlers.SendEmail(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}

		expected := "Invalid email address"
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})
}
