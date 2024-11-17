package services

import (
	"path/filepath"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	RootPath   = filepath.Join(filepath.Dir(b), "../")
)

// func SaveFile(file multipart.File, handler *multipart.FileHeader) (string, error){

// }
