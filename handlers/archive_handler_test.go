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

func TestArchiveInformation(t *testing.T) {
	tempZip := "./test.zip"
	createTestZip(tempZip)
	defer os.Remove(tempZip)

	body, contentType, err := CreateMultipartRequest(tempZip, "file")
	if err != nil {
		t.Fatalf("Failed to create multipart request: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/archive/information", body)
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

func TestFormArchive(t *testing.T) {
	body, contentType, err := CreateMultipartRequest("../data/notZip.jpg", "files")
	if err != nil {
		t.Fatalf("Failed to create multipart request: %v", err)
	}
	
	// Create a request to test the handler
	req := httptest.NewRequest("POST", "/api/archive/files", body)
	req.Header.Set("Content-Type", contentType)

	// Record the response
	w := httptest.NewRecorder()
	FormArchive(w, req)

	res := w.Result()
	defer res.Body.Close()

	// Check the status code
	if res.StatusCode != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", res.StatusCode, http.StatusCreated)
		msg, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Error reading body: %v", err)
		}
		t.Error(string(msg))
	}

	// Check the response body
	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}
	if len(responseBody) == 0 {
		t.Errorf("Handler returned empty response body")
	}
	// Additional checks (e.g., verify that the response is a zip file)
	contentType = res.Header.Get("Content-Type")
	if contentType != "application/zip" {
		t.Errorf("Handler returned wrong content type: got %v want application/zip", contentType)
	}

	// os.WriteFile("debug-archive.zip", responseBody, 0644)
	//defer os.Remove("debug-archive.zip")

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
