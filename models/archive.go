package models

type Archive struct {
	FileName    string  `json:"filename"`
	ArchiveSize float64 `json:"archive_size"`
	TotalSize   float64 `json:"total_size"`
	TotalFiles  int     `json:"total_files"`
	Files       []File  `json:"files"`
}

type File struct {
	FilePath string  `json:"file_path"`
	Size     float64 `json:"size"`
	MimeType string  `json:"mimetype"`
}
