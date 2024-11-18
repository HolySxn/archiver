package handlers

import (
	"archive/zip"
	"archiver/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

var (
	OPEN_XML = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	XML      = "application/xml"
	PNG      = "image/png"
	JPEG     = "image/jpeg"
	PDF      = "application/pdf"
)

// Common utility to handle errors and send responses
func sendError(w http.ResponseWriter, message string, status int) {
	http.Error(w, message, status)
	slog.Error(message)
}

func processZipFile(w http.ResponseWriter, r *http.Request) ([]models.File, float64, error) {
	// Parse multipart form
	err := r.ParseMultipartForm(0)
	if err != nil {
		sendError(w, "Failed to parse multipart form", http.StatusInternalServerError)
		return nil, 0, err
	}

	// Retrieve file
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		sendError(w, "Failed to retrieve file", http.StatusInternalServerError)
		return nil, 0, err
	}
	defer file.Close()

	// Check file type
	if filepath.Ext(fileHeader.Filename) != ".zip" {
		f, err := os.Open("./data/notZip.jpg")
		if err != nil {
			http.Error(w, "Not appropriate file type", http.StatusInternalServerError)
			return nil, 0, err
		}
		defer f.Close()

		img, err := io.ReadAll(f)
		if err != nil {
			http.Error(w, "Not appropriate file type", http.StatusInternalServerError)
			return nil, 0, err
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(img)
		return nil, 0, err
	}

	// Create temp file to save zip file
	tempFile, err := os.CreateTemp("", fileHeader.Filename)
	if err != nil {
		http.Error(w, "Failed to create tempFile", http.StatusInternalServerError)
		return nil, 0, err
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.ReadFrom(file)
	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return nil, 0, err
	}
	tempFile.Close()

	archive, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		http.Error(w, fmt.Sprintf("This is not a valid archive! (%s)", err.Error()), http.StatusBadRequest)
		return nil, 0, err
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

	return files, totalSize, nil
}

func ArchiveInformation(w http.ResponseWriter, r *http.Request) {

	files, totalSize, err := processZipFile(w, r)
	if err != nil {
		se
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
	err := r.ParseMultipartForm(0)
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
	tempFile, err := os.CreateTemp("", "archive-*.zip")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	zipWriter := zip.NewWriter(tempFile)

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Error opening file", http.StatusInternalServerError)
			return
		}

		buffer := make([]byte, 512)
		_, err = file.Read(buffer)
		if err != nil && err != io.EOF {
			file.Close()
			http.Error(w, "Error reading file header", http.StatusInternalServerError)
			return
		}

		mimeType, err := GetMimeType(file)
		if err != nil {
			file.Close()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		file.Seek(0, 0)

		if mimeType != OPEN_XML && mimeType != XML && mimeType != JPEG && mimeType != PNG {
			file.Close()
			http.Error(w, fmt.Sprintf("Unsupported file type: %s", fileHeader.Filename), http.StatusBadRequest)
			return
		}

		fileWriter, err := zipWriter.Create(fileHeader.Filename)
		if err != nil {
			file.Close()
			http.Error(w, "Error creating file writer", http.StatusInternalServerError)
			return
		}

		buffer = make([]byte, 1024*1024) // 1 MB buffer size
		for {
			n, err := file.Read(buffer)
			if n > 0 {
				_, writeErr := fileWriter.Write(buffer[:n])
				if writeErr != nil {
					file.Close()
					http.Error(w, "Error writing file to ZIP", http.StatusInternalServerError)
					return
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				file.Close()
				http.Error(w, "Error reading file", http.StatusInternalServerError)
				return
			}
		}
		file.Close()
		log.Printf("Added file to ZIP: %s\n", fileHeader.Filename)
	}

	// if err := zipWriter.Close(); err != nil {
	// 	http.Error(w, "Error finalizing ZIP archive", http.StatusInternalServerError)
	// 	return
	// }

	tempFile.Seek(0, 0)
	stat, _ := tempFile.Stat()

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"archive.zip\"")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	w.WriteHeader(http.StatusCreated)

	_, err = io.Copy(w, tempFile)
	if err != nil {
		http.Error(w, "Error sending archive", http.StatusInternalServerError)
	}
}

// Function to get MIME type of the file
func GetMimeType(file io.Reader) (string, error) {
	buf := make([]byte, 512)
	_, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	mimeType := http.DetectContentType(buf)
	return mimeType, nil
}
