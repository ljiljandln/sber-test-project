# sber-test-project
Простое REST API для создания и управления задачами на день. Реализованы CRUD-операции, фильтрация по дате и статусу (выполнено/не выполнено), пагинация и Swagger-документация.

Клонирование   
```bash
git clone https://github.com/ljiljandln/sber-test-project.git
cd sber-test-project/todo-api
```

## 🚀 Запуск приложения

1. Запуск контейнера:
```bash
make up
```
или
```bash
docker-compose up --build
```
2. Остановка контейнера
```bash
make down
```
3. Запуск тестов
```bash
make test
```

приложение доступно по адресу http://localhost:8081,
сваггер-документация - http://localhost:8081/swagger/index.html
