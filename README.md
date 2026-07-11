# Файлы для итогового задания

В директории `tests` находятся тесты для проверки API, которое должно быть реализовано в веб-сервере.

Директория `web` содержит файлы фронтенда.

# Описание проекта 
- Создание, редактирование, удаление задач
- Повторяющиеся задачи (ежедневно, еженедельно, ежемесячно, ежегодно)
- Отметка задачи
- Аутентификация через JWT токены
- Хранение данных через SQLite
- Готов к Докеру

# Для запуска 
- установите все пакеты = go mod download
- запуск с паролем export TODO_PASSWORD=12345 go run main.go 
- тесты ОБЯЗАТЕЛЬНО после запуска в терминале запускать :

    go test -run ^TestApp$ ./tests
    go test -run ^TestDB$ ./tests
    go test -run ^TestNextDate$ ./tests
    go test -run ^TestAddTask$ ./tests 
    go test -run ^TestTasks$ ./tests
    go test -run ^TestEditTask$ ./tests
    Или же просто проверить все тесты одновременно go test ./tests или go test ./tests -v

- http://localhost:7540 откройте после в браузере

# Для сборки образа Docker 
- docker build -t scheduler-app .
- docker run -d -p 7540:7540 scheduler scheduler-app

# Задание со звездочкой
- Ежемесячные и еженедельные правила повторения 
