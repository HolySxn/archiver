package handlers

import (
	"archive/zip"
	"archiver/internal/utils"
	"archiver/models"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

var (
	OPEN_XML = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	XML      = "application/xml"
	PNG      = "image/png"
	JPEG     = "image/jpeg"
	PDF      = "application/pdf"
)

// Process zip file and return list of files, total size and error
func processZipFile(w http.ResponseWriter, r *http.Request) ([]models.File, float64, string, error) {
	// Parse multipart form
	err := r.ParseMultipartForm(0)
	if err != nil {
		utils.SendError(w, "Failed to parse multipart form", http.StatusInternalServerError)
		return nil, 0, "", err
	}

	// Retrieve file
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		utils.SendError(w, "Failed to retrieve file", http.StatusBadRequest)
		return nil, 0, "", err
	}
	defer file.Close()

	// Check file type
	if filepath.Ext(fileHeader.Filename) != ".zip" {
		utils.SendError(w, "Unsupported file type", http.NotAppropriateFile)
		return nil, 0, "", errors.New("unsupported file type")
	}

	// Create temp file to save zip file
	tempFile, err := os.CreateTemp("", fileHeader.Filename)
	if err != nil {
		utils.SendError(w, "Failed to create tempFile", http.StatusBadRequest)
		return nil, 0, "", err
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.ReadFrom(file)
	if err != nil {
		utils.SendError(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return nil, 0, "", err
	}
	tempFile.Close()

	archive, err := zip.OpenReader(tempFile.Name())
	if err != nil {
		utils.SendError(w, fmt.Sprintf("This is not a valid archive! (%s)", err.Error()), http.StatusBadRequest)
		return nil, 0, "", err
	}
	defer archive.Close()

	var files []models.File
	var totalSize float64
	for _, f := range archive.File {
		if f.CompressedSize64 == 0 {
			continue
		}
		mimeType := GetMimeType(f.Name)
		fileSize := float64(f.UncompressedSize64)
		totalSize += fileSize
		files = append(files, models.File{
			FilePath: f.Name,
			Size:     fileSize,
			MimeType: mimeType,
		})
	}

	return files, totalSize, fileHeader.Filename, nil
}

// Handle function to get archive information
func ArchiveInformation(w http.ResponseWriter, r *http.Request) {
	files, totalSize, archiveName, err := processZipFile(w, r)
	if err != nil {
		return
	}

	response := models.Archive{
		FileName:    archiveName,
		ArchiveSize: float64(r.ContentLength),
		TotalSize:   totalSize,
		TotalFiles:  len(files),
		Files:       files,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// FormArchive handles file uploads, compresses them into a zip file, and sends back the archive
func FormArchive(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(0)
	if err != nil {
		utils.SendError(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files[]"]
	if len(files) == 0 {
		utils.SendError(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	// Create temp zip to save files
	tempFile, err := os.CreateTemp("", "archive-*.zip")
	if err != nil {
		utils.SendError(w, "Failed to create temp file", http.StatusInternalServerError)
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
			utils.SendError(w, "Error opening file", http.StatusInternalServerError)
			return
		}

		mimeType := GetMimeType(fileHeader.Filename)

		if mimeType != OPEN_XML && mimeType != XML && mimeType != JPEG && mimeType != PNG {
			file.Close()
			utils.SendError(w, fmt.Sprintf("Unsupported file type: %s, type:%s", fileHeader.Filename, mimeType), http.StatusBadRequest)
			return
		}

		fileWriter, err := zipWriter.Create(fileHeader.Filename)
		if err != nil {
			file.Close()
			utils.SendError(w, "Error creating file writer", http.StatusInternalServerError)
			return
		}

		buffer := make([]byte, 1024*1024) // 1 MB buffer size
		for {
			n, err := file.Read(buffer)
			if n > 0 {
				_, writeErr := fileWriter.Write(buffer[:n])
				if writeErr != nil {
					file.Close()
					utils.SendError(w, "Error writing file to ZIP", http.StatusInternalServerError)
					return
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				file.Close()
				utils.SendError(w, "Error reading file", http.StatusInternalServerError)
				return
			}
		}
		file.Close()
		log.Printf("Added file to ZIP: %s\n", fileHeader.Filename)
	}

	if err := zipWriter.Close(); err != nil {
		utils.SendError(w, "Error finalizing ZIP archive", http.StatusInternalServerError)
		return
	}

	tempFile.Seek(0, 0)
	stat, _ := tempFile.Stat()

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\"archive.zip\"")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	w.WriteHeader(http.StatusCreated)

	_, err = io.Copy(w, tempFile)
	if err != nil {
		utils.SendError(w, "Error sending archive", http.StatusInternalServerError)
	}
}

// Function to get MIME type of the file
func GetMimeType(filename string) string {
	mimeType := mime.TypeByExtension(path.Ext(filename))
	if mimeType == "text/xml; charset=utf-8" {
		mimeType = "application/xml"
	}
	return mimeType
}
