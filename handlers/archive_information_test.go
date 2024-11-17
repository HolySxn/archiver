package handlers

import (
	"archive/zip"
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// func TestFormArchive(t *testing.T) {
// 	// Create a temporary file to simulate the uploaded file
// 	tempFile, err := os.CreateTemp("", "test-*.txt")
// 	if err != nil {
// 		t.Fatalf("Failed to create temp file: %v", err)
// 	}
// 	defer os.Remove(tempFile.Name())

// 	// Write some content to the temp file
// 	_, err = tempFile.Write([]byte("This is a test file"))
// 	if err != nil {
// 		t.Fatalf("Failed to write to temp file: %v", err)
// 	}
// 	tempFile.Close()

// 	// Create a multipart form file field with the temp file
// 	body := &bytes.Buffer{}
// 	writer := multipart.NewWriter(body)
// 	fileWriter, err := writer.CreateFormFile("files", filepath.Base(tempFile.Name()))
// 	if err != nil {
// 		t.Fatalf("Failed to create form file: %v", err)
// 	}

// 	file, err := os.Open(tempFile.Name())
// 	if err != nil {
// 		t.Fatalf("Failed to open temp file: %v", err)
// 	}
// 	defer file.Close()

// 	_, err = io.Copy(fileWriter, file)
// 	if err != nil {
// 		t.Fatalf("Failed to copy file content: %v", err)
// 	}
// 	writer.Close()

// 	// Create a request to test the handler
// 	req := httptest.NewRequest("POST", "http://localhost:8080/api/archive/form-archive", body)
// 	req.Header.Set("Content-Type", writer.FormDataContentType())

// 	// Record the response
// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(handlers.FormArchive)
// 	handler.ServeHTTP(rr, req)

// 	// Check the status code
// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
// 	}

// 	// Check the response body
// 	responseBody := rr.Body.String()
// 	if responseBody == "" {
// 		t.Errorf("Handler returned empty response body")
// 	}

// 	// Additional checks (e.g., verify that the response is a zip file)
// 	contentType := rr.Header().Get("Content-Type")
// 	if contentType != "application/zip" {
// 		t.Errorf("Handler returned wrong content type: got %v want application/zip", contentType)
// 	}
// }

func TestArchiveInformation(t *testing.T) {
	tempZip := "./test.zip"
	createTestZip(tempZip)
	defer os.Remove(tempZip)

	body, contentType, err := CreateMultipartRequest(tempZip, "file")
	if err != nil {
		t.Fatalf("Failed to create multipart request: %v", err)
	}

	req := httptest.NewRequest("Post", "/api/archive/information", body)
	req.Header.Set("Content-Type", contentType)

	w := httptest.NewRecorder()

	ArchiveInformation(w, req)

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.StatusCode)
		msg, err := io.ReadAll(res.Body)
		if err != nil {
			return
		}
		t.Error(string(msg))
	}

}

func CreateMultipartRequest(filepath string, fieldname string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(filepath)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	part, err := writer.CreateFormFile(fieldname, filepath)
	if err != nil {
		return nil, "", err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, "", err
	}

	err = writer.Close()
	if err != nil {
		return nil, "", err

	}

	return body, writer.FormDataContentType(), nil

}

func createTestZip(filename string) {
	tempFile, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	txtFile, err := os.CreateTemp("", "test.txt")
	if err != nil {
		panic(err)
	}
	defer os.Remove(txtFile.Name())

	_, err = txtFile.Write([]byte("hello world!"))
	if err != nil {
		panic(err)
	}

	f1, err := os.Open(txtFile.Name())
	if err != nil {
		panic(err)
	}
	defer f1.Close()

	w1, err := zipWriter.Create("txt/test.txt")
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(w1, f1)
	if err != nil {
		panic(err)
	}
}
