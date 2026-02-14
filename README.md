# CloudStorage - Файловое хранилище
Сервис для хранения файлов с веб-интерфейсом и REST API.

### Требования
- Docker
- Docker Compose

## 🚀 Запуск
```bash
git clone https://github.com/impeaone/CloudStorage
cd CloudStorage 
docker-compose up -d
```

## 📡 Что запускается
После запуска доступны:
<div align="center">

| Сервис        | URL                     | Логин/Пароль      | Назначение            |
|---------------|-------------------------|-------------------|-----------------------|
| fileserver    | http://localhost:11682  | -                 | Основное приложение   | 
| MinIO Console | http://localhost:9001   | user/password     | Управление хранилищем | 
| Grafana       | http://localhost:3000   | admin/admin       | Мониторинг            | 
| Adminer       | http://localhost:8080   | postgres/postgres | Управление БД         | 
| RedisInsight  | http://localhost:8001   | -                 | Просмотр Redis        | 
| Prometheus    | http://localhost:9090   | -                 | Метрики               |
</div>

## 🔌 Основные эндпоинты
### Приложение (порт 11682)
### Файлы
```text
GET     /client/api/v1/get-file        # Получить файл
POST    /client/api/v1/upload-files    # Загрузить файл
GET     /client/api/v1/get-files-list  # Получить список файлов конкретного пользователя
DELETE  /client/api/v1/delete-file     # Удалить файл
```
### Web UI
```text
GET     /index                    # Страница входа
GET     /client/api/v1/storage/   # Страница с хранилищем
```
### Метрики (порт 11680)
```text
GET /metrics
```

## ⚙️ Конфигурация
#### Для запуска необходимо в корне создать .env файл
```text
NUM_CPU=4
SERVER_PORT=11682
MINIO_ENDPOINT=minio:9000
MINIO_EXAMPLE_BUCKET=test
MINIO_ROOT_USER=user
MINIO_ROOT_PASSWORD=password
MINIO_USER_SSL=false
SERVER_PORT=11682
SERVER_IP=0.0.0.0
PG_USER=postgres
PG_PASSWORD=postgres
PG_HOST=postgres
PG_PORT=5432
PG_DATABASE=storage
TEST_API_NEEDED=true
TEST_API_KEY=test
TEST_API_EMAIL=test@test.test
CloudStorage_LOGGER=INFO
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
METRICS_SERVER_PORT=11680
METRICS_SERVER_IP=0.0.0.0
```
## 📚 Документация
### Swagger UI
#### После запуска сервера доступна по адресу:
```text 
http://localhost:11682/swagger/index.html
```
### OpenAPI спецификации
- JSON: ```http://localhost:11682/swagger/doc.json```
- YAML: ```http://localhost:11682/swagger/doc.yaml```

## 🖼️ Демонстрация Web UI

| Главная страница | Хранилище |
|------------------|-----------|
| <img src="images/index.png" width="600"> | <img src="images/storage.png" width="600"> |



