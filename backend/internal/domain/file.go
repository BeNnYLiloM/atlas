package domain

import "time"

type File struct {
	ID           string    `json:"id" db:"id"`
	MessageID    *string   `json:"message_id" db:"message_id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Filename     string    `json:"filename" db:"filename"`
	OriginalName string    `json:"original_name" db:"original_name"`
	MimeType     string    `json:"mime_type" db:"mime_type"`
	SizeBytes    int64     `json:"size_bytes" db:"size_bytes"`
	StoragePath  string    `json:"-" db:"storage_path"`
	URL          string    `json:"url" db:"-"` // Генерируется динамически
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type FileUploadResult struct {
	File *File  `json:"file"`
	URL  string `json:"url"`
}
