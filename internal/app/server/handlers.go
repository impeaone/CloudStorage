package server

import (
	minioClient "CloudStorageProject-FileServer/internal/minio"
	logger2 "CloudStorageProject-FileServer/pkg/logger/logger"
	"CloudStorageProject-FileServer/pkg/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// TODO: поверки бакетов надо сделать в функциях всех
func getFileFunc(w http.ResponseWriter, r *http.Request) {
	// пример запроса GET /client/api/v1/get-file?api=apikey&filename=minecraft.png

	logger := r.Context().Value("logger").(*logger2.Log)
	// Если методо не тот
	if r.Method != "GET" {
		logger.Warning(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: user uses not allowed method",
			r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Получаем нужные параметры строки запроса
	api := r.URL.Query().Get("api")
	filename := r.URL.Query().Get("filename")
	// Если названия файла в параметре строки нет
	if filename == "" {
		logger.Warning(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: bad filename parameter",
			r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, "filename is required", http.StatusBadRequest)
		return
	}
	// Достаем minio-пул из контекста
	Minio := r.Context().Value("minio").(*minioClient.MinioClient)

	// Получаем запрошенный файл из minio
	fileMinio, err := Minio.GetOne(api, filename)
	if err != nil {
		logger.Error(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: get minio-file error:%v",
			r.RemoteAddr, r.URL, r.Method, err, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// получаем характеристики файла
	stat, errStat := fileMinio.Stat()
	if errStat != nil {
		logger.Error(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: minio-file error:%v",
			r.RemoteAddr, r.URL, r.Method, errStat, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, errStat.Error(), http.StatusBadRequest)
		return
	}
	// Устанавливаем необходимые заголовки и возвращаем результат
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))
	io.Copy(w, fileMinio)
	return
}

func storeFilesFunc(w http.ResponseWriter, r *http.Request) {
	// Берем логгер из контекста
	logger := r.Context().Value("logger").(*logger2.Log)
	// Если метод не тот
	if r.Method != "POST" {
		logger.Warning(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: user uses not allowed method",
			r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	api := r.URL.Query().Get("api")
	minio := r.Context().Value("minio").(*minioClient.MinioClient)

	// MultipartReader для чтения form-data
	reader, err := r.MultipartReader()
	if err != nil {
		logger.Error(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: multipartReader error:%v",
			r.RemoteAddr, r.URL, r.Method, err, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, "Openfile error", http.StatusInternalServerError)
		return
	}

	// слайсы для загруженных файлов и ошибок
	var uploaded []string
	var errors []string

	// Читаем части multipart формы по очереди
	for {
		part, errNext := reader.NextPart()
		if errNext == io.EOF {
			break
		}
		if errNext != nil {
			logger.Error(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: nextPart error:%v",
				r.RemoteAddr, r.URL, r.Method, errNext, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
			errors = append(errors, fmt.Sprintf("Error reading part: %v", errNext))
			continue
		}

		// Проверяем, что это файл (а не поле формы)
		if part.FileName() == "" {
			part.Close()
			continue
		}

		// Создаем временный файл для партишиона
		tempFile, errTemp := os.CreateTemp("", "upload-*")
		if errTemp != nil {
			logger.Error(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: create temp file error:%v",
				r.RemoteAddr, r.URL, r.Method, errTemp, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
			part.Close()
			errors = append(errors, fmt.Sprintf("Error creating temp file for %s: %v", part.FileName(), errTemp))
			continue
		}
		tempFileName := tempFile.Name()

		// Копируем данные из part во временный файл
		fileSize, errCopy := io.Copy(tempFile, part)
		part.Close()
		tempFile.Close()

		if errCopy != nil {
			logger.Error(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: copy part to temp error:%v",
				r.RemoteAddr, r.URL, r.Method, errCopy, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
			os.Remove(tempFileName)
			errors = append(errors, fmt.Sprintf("Error saving %s: %v", part.FileName(), errCopy))
			continue
		}

		// Теперь открываем временный файл для чтения и загружаем в MinIO
		fileForUpload, errOpen := os.Open(tempFileName)
		if errOpen != nil {
			logger.Error(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: open temp error:%v",
				r.RemoteAddr, r.URL, r.Method, errOpen, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
			os.Remove(tempFileName)
			errors = append(errors, fmt.Sprintf("Error reopening %s: %v", part.FileName(), errOpen))
			continue
		}

		contentType := "application/octet-stream"

		filePartition := models.FileMinio{
			FileName:    part.FileName(),
			Reader:      fileForUpload,
			Size:        fileSize,
			ContentType: contentType,
		}

		uploadErr := minio.CreateOne(api, filePartition)

		// Закрываем и удаляем временный файл
		fileForUpload.Close()
		os.Remove(tempFileName)

		if uploadErr != nil {
			logger.Error(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: upload file to minio error:%v",
				r.RemoteAddr, r.URL, r.Method, uploadErr, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
			errors = append(errors, fmt.Sprintf("Error uploading %s: %v", part.FileName(), uploadErr))
			continue
		}

		uploaded = append(uploaded, part.FileName())
	}

	// Получаем список файлов
	fileList, errList := minio.FilesList(api)
	if errList != nil {
		logger.Error(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: open temp error:%v",
			r.RemoteAddr, r.URL, r.Method, errList, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		fileList = []models.FileWebResponse{}
	}

	// Формируем ответ
	response := models.CreateFileResponse{
		Status:        200,
		Message:       fmt.Sprintf("Successfully uploaded %d files", len(uploaded)),
		NewFiles:      fileList,
		UploadedFiles: uploaded,
	}

	if len(errors) > 0 {
		response.Message = fmt.Sprintf("Uploaded %d files with %d errors", len(uploaded), len(errors))
	}

	w.Header().Set("Content-Type", "application/json")
	bytes, _ := json.Marshal(response)
	w.Write(bytes)
}

func deleteFilesFunc(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*logger2.Log)
	if r.Method != "DELETE" {
		logger.Warning(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: user uses not allowed method",
			r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	api := r.URL.Query().Get("api")
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		logger.Warning(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: bad filename parameter",
			r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, "filename is required", http.StatusBadRequest)
		return
	}
	minio := r.Context().Value("minio").(*minioClient.MinioClient)

	errDelete := minio.Delete(api, filename)
	if errDelete != nil {
		logger.Error(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: delete minio file error: %v",
			r.RemoteAddr, r.URL, r.Method, errDelete, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, "Error", http.StatusNotFound)
		return
	}
	fileList, errList := minio.FilesList(api)
	if errList != nil {
		logger.Error(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: get minio files error: %v",
			r.RemoteAddr, r.URL, r.Method, errList, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, "Error", http.StatusNotFound)
		return
	}

	response := models.CreateFileResponse{
		Status:   200,
		Message:  "success",
		NewFiles: fileList,
	}
	w.Header().Set("Content-Type", "application/json")
	bytes, _ := json.Marshal(response)
	w.Write(bytes)
	return
}

func getFilesListFunc(w http.ResponseWriter, r *http.Request) {
	// пример запроса: POST /client/api/v1/get-files-list?api=api_key
	logger := r.Context().Value("logger").(*logger2.Log)
	if r.Method != "GET" {
		logger.Warning(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: user uses not allowed method",
			r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	api := r.URL.Query().Get("api")
	minio := r.Context().Value("minio").(*minioClient.MinioClient)

	files, err := minio.FilesList(api)
	if err != nil {
		logger.Error(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: get minio files error: %v",
			r.RemoteAddr, r.URL, r.Method, err, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, "Error", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	bytes, _ := json.Marshal(files)
	w.Write(bytes)
	return
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*logger2.Log)
	if r.Method != "GET" {
		logger.Warning(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: user uses not allowed method",
			r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	TemplatePath := r.Context().Value("tmplPath").(string)
	apikey, is := r.Cookie("apikey")
	if is != nil {
		http.ServeFile(w, r, TemplatePath+"/index.html")
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/client/api/v1/storage?api=%s", apikey.Value), http.StatusFound)
	return
}

func storagePage(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*logger2.Log)
	if r.Method != "GET" {
		logger.Warning(fmt.Sprintf("Client: %s; EndPoint: %s; Method: %s; Time: %v; Message: user uses not allowed method",
			r.RemoteAddr, r.URL, r.Method, time.Now().Format("02.01.2006 15:04:05")), logger2.GetPlace())
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	TemplatePath := r.Context().Value("tmplPath").(string)
	apikey := r.URL.Query().Get("api")
	if apikey == "" || apikey == "undefined" {
		http.Redirect(w, r, "/index", http.StatusFound)
		return
	}
	http.ServeFile(w, r, TemplatePath+"/storage.html")
	return
}

func zeroPath(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/index", http.StatusFound)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	health := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		Status:  "OK",
		Message: "Сервер запущен и нормально функционирует",
	}
	w.Header().Set("Content-Type", "application/json")
	bytes, _ := json.Marshal(health)
	w.Write(bytes)
	return
}
