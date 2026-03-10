package tools

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
)

func FormatFileSize(sizeBytes int64) string {
	const (
		KB = 1024.0
		MB = 1024.0 * 1024.0
		GB = 1024.0 * 1024.0 * 1024.0
		TB = 1024.0 * 1024.0 * 1024.0 * 1024.0
	)

	size := float64(sizeBytes)

	switch {
	case sizeBytes >= TB:
		return fmt.Sprintf("%.2f TB", size/TB)
	case sizeBytes >= GB:
		return fmt.Sprintf("%.2f GB", size/GB)
	case sizeBytes >= MB:
		return fmt.Sprintf("%.2f MB", size/MB)
	case sizeBytes >= KB:
		return fmt.Sprintf("%.2f KB", size/KB)
	default:
		return fmt.Sprintf("%d B", sizeBytes)
	}
}

// GetPlace - получает файл от куда вызвана функция была
func GetPlace() string {
	_, file, line, _ := runtime.Caller(1)
	split := strings.Split(file, "/")
	StartFile := split[len(split)-1]
	place := StartFile + ":" + strconv.Itoa(line)
	return place
}
