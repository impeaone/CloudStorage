package server

func validateAPI(api string) bool {
	if api == "" {
		return false
	}
	//TODO: подключение к бд проверка, что api вообще есть и может быть использован для хранилища
	return true
}

// На сервер приходит запрос на получение файла
func validateAPItoFile(api, filename string) bool {
	//TODO:  тут тоже поход в бд, проверка, что файл относится к api ключу этому
	return true
}
