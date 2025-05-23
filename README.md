# Содержание
1. [Описание проекта](#описание-проекта)
2. [Быстрый старт](#быстрый-старт)
3. [Разработка](#разработка)
4. [Тестирование](#тестирование)
5. [API](#api)

---

## 📁 Описание проекта

Сервис предоставляет REST API для:
- Сбора данных с датчиков 🖥️
- Хранения и анализа показаний 📊
- Управления устройствами умного дома 🏠

### Поддерживаемые датчики
| Тип | Примеры | Характеристики |
|------|---------|----------------|
| `ContactClosure` | Датчики дверей, протечек | Дискретные сигналы (0/1) |
| `ADC` | Термометры, гигрометры | Аналоговые значения |

---

## 🚀 Быстрый старт

### Требования
- Go 1.20+
- Docker 20.10+
- PostgreSQL 14+

### Установка
```bash
# 1. Запуск БД
docker-compose up -d

# 2. Миграции
migrate -path=./migrations -database postgres://postgres:postgres@localhost:5432/db?sslmode=disable up

# 3. Запуск сервера
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/db?sslmode=disable"
go run cmd/main.go
```
## 🧪 Тестирование

```bash
# Unit-тесты
go test ./... -race -v
```

## 🧑‍💻 API
Документация доступна после запуска:

-   Swagger UI:  `http://localhost:8080/docs`
    
-   OpenAPI спецификация:  `http://localhost:8080/swagger.json`
