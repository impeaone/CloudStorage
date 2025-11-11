package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func getFilesFunc(w http.ResponseWriter, r *http.Request) {
	// пример запроса GET /client/get-file?api=api-key&filename=minecraft.png
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	api := r.URL.Query().Get("api")
	filename := r.URL.Query().Get("filename")
	if !validateAPI(api) {
		//logger
		http.Error(w, "Invalid API", http.StatusBadRequest)
	}
	if !validateAPItoFile(api, filename) {
		//logger
		http.Error(w, "File not found", http.StatusBadRequest)
	}

	file, err := os.Open(fmt.Sprintf("%s/%s", FilesDirectory, filename))
	if err != nil {
		//logger
		http.Error(w, "File not found", http.StatusNotFound)
	}
	defer file.Close()

	_, errCopy := io.Copy(w, file)
	if errCopy != nil {
		//logger
		http.Error(w, errCopy.Error(), http.StatusInternalServerError)
	}
	//logger info

}

// TODO: додэлать
func storeFilesFunc(w http.ResponseWriter, r *http.Request) {
	// пример запроса: POST /client/upload-files
	// все файлы в теле запроса
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	api := r.FormValue("apikey")

	//TODO: поход в бд, сохранение файла,
	//TODO: создание папки если ее нет
	apiDIR := api
	// получаем файлы из формы
	files := r.MultipartForm.File["files"]
	if !validateAPI(api) {
		http.Error(w, "Invalid API", http.StatusBadRequest)
	}

	for _, fileHend := range files {
		file, err := fileHend.Open()
		if err != nil {
			//logger
			continue
		}
		dst, errC := os.Create(fmt.Sprintf("%s/%s/%s", FilesDirectory, apiDIR, fileHend.Filename))
		if errC != nil {
			continue
		}

		io.Copy(dst, file)

		//logger
		file.Close()
		dst.Close()
	}

}
