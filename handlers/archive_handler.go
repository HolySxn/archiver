package handlers

import (
	"archive/zip"
	"archiver/models"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

// type ArchiveHandler struct{}

// func NewArchiveHandler() *ArchiveHandler {
// 	return &ArchiveHandler{}
// }

func ArchiveInformation(w http.ResponseWriter, r *http.Request) {
	// Prase file size
	err := r.ParseMultipartForm(30 << 20) // 30 mb
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Retrieve file
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to retrieve file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Check file type
	if filepath.Ext(fileHeader.Filename) != ".zip" {
		f, err := os.Open("./data/notZip.jpg")
		if err != nil {
			http.Error(w, "Not appropriate file type", http.StatusInternalServerError)
			return
		}
		defer f.Close()

		img, err := io.ReadAll(f)
		if err != nil {
			http.Error(w, "Not appropriate file type", http.StatusInternalServerError)
			return
		}

		w.Write(img)
	}

	// Create temp file to save zip file
	tempFile, err := os.CreateTemp("", fileHeader.Filename)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to create tempFile", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.ReadFrom(file)
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tempFile.Close()

	archive, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		http.Error(w, fmt.Sprintf("This is not a valid archive! (%s)", err.Error()), http.StatusBadRequest)
		return
	}
	defer archive.Close()

	var files []models.File
	var totalSize float64
	for _, f := range archive.File {
		mimeType := mime.TypeByExtension(filepath.Ext(f.Name))
		if mimeType == "" {
			continue
		}
		fileSize := float64(f.UncompressedSize64)
		totalSize += fileSize
		files = append(files, models.File{
			FilePath: f.Name,
			Size:     fileSize,
			MimeType: mimeType,
		})
	}

	response := models.Archive{
		FileName:    fileHeader.Filename,
		ArchiveSize: float64(fileHeader.Size),
		TotalSize:   totalSize,
		TotalFiles:  len(files),
		Files:       files,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func FormArchive(w http.ResponseWriter, r *http.Request) {
	// Prase file size
	err := r.ParseMultipartForm(30 << 20) // 30 mb
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]

	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	// Create temp zip to save files
	tempFile, err := os.CreateTemp("", "archive.zip")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to create tempFile", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())

	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	for _, fileHeader := range files {
		mimeType := mime.TypeByExtension(filepath.Ext(fileHeader.Filename))
		if mimeType != "application/vnd.openxmlformats-officedocument.wordprocessingml.document" && mimeType != "application/xml" && mimeType != "image/jpeg" && mimeType != "image/png" {
			http.Error(w, fmt.Sprintf("Unsupported file type: %s", fileHeader.Filename), http.StatusBadRequest)
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}

		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		fileWriter, err := zipWriter.Create(fileHeader.Filename)
		if err != nil {
			http.Error(w, "Error creating file writer", http.StatusInternalServerError)
			return
		}

		_, err = fileWriter.Write(data)
		if err != nil {
			http.Error(w, "Error writing file", http.StatusInternalServerError)
			return
		}

		file.Close()

	}

	tempFile.Seek(0, 0)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"archive.zip\"")
	w.WriteHeader(200)
	_, err = io.Copy(w, tempFile)
	if err != nil {
		http.Error(w, "Error sending archive", http.StatusInternalServerError)
	}
}
