package utils

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"os"
)

var (
	NotZip        = "./errors/NotZip.jpg"
	BadReq        = "./errors/BadRequest.jpg"
	InternalError = "./errors/InternalServerError.jpg"
)

// Common utility to handle errors and send responses
func SendError(w http.ResponseWriter, message string, status int) {
	var img []byte
	var err error
	switch status {
	case http.NotAppropriateFile:
		img, err = openFile(NotZip)
		status = http.StatusBadRequest
	case http.StatusBadRequest:
		img, err = openFile(BadReq)
	case http.StatusInternalServerError:
		img, err = openFile(InternalError)
	default:
		http.Error(w, message, status)
		return
	}

	if err != nil {
		http.Error(w, message, status)
		slog.Error(err.Error())
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	w.Write(img)
	slog.Error(message)
}

func openFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, file)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
