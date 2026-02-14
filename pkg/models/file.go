package models

import (
	"io"
	"time"
)

// FileInfo просто структура для сваггера, на minio.Object или его алиас жалуется
type FileInfo struct {
	Name         string `json:"name" example:"alohadance.png"`
	Size         int64  `json:"size" example:"1024"`
	ContentType  string `json:"content_type" example:"image/png"`
	LastModified string `json:"last_modified" example:"2024-01-01T12:00:00Z"`
	ETag         string `json:"etag" example:"\"33a64df551425fcc55e4d42a148795d9f25f89d4\""`
}
type FileMinio struct {
	FileName    string
	Data        []byte
	Reader      io.Reader // для больших файлов, для стриминга
	ContentType string    // для стриминга
	Size        int64
}

type FileWebResponse struct {
	FileName    string `json:"file_name"`
	FileType    string `json:"file_type"`
	LastModTime string `json:"create_date"`
	FileSize    string `json:"file_size"`
}

type FileResponse struct {
	Status        int               `json:"status"`
	Message       string            `json:"message"`
	NewFiles      []FileWebResponse `json:"new_files"`
	UploadedFiles []string          `json:"uploaded_files"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"time_stamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
}
