# Руководство по деплою Go Backend на Vercel

## Шаг 1: Создание и выполнение SQL скрипта

### 1.1. Выполнение скрипта в v0
1. В интерфейсе v0 найди файл `scripts/001_create_users_table.sql`
2. Нажми кнопку "Run" для выполнения скрипта
3. Дождись подтверждения успешного выполнения

### 1.2. Проверка таблицы
После выполнения скрипта таблица `users` будет создана со следующими элементами:
- Все поля (id, username, email, password_hash, и т.д.)
- Индексы на email, username, trainer_id
- Триггер для автоматического обновления updated_at

## Шаг 2: Структура Go проекта для Vercel

Vercel поддерживает Go через Serverless Functions. Структура должна быть следующей:

\`\`\`
project/
├── api/
│   ├── register.go      # POST /api/register
│   ├── login.go         # POST /api/login
│   ├── profile.go       # GET /api/profile
│   └── refresh.go       # POST /api/refresh
├── pkg/
│   ├── auth/
│   │   ├── jwt.go       # JWT утилиты
│   │   └── password.go  # Хеширование паролей
│   └── db/
│       └── supabase.go  # Подключение к Supabase
├── vercel.json
└── go.mod
\`\`\`

## Шаг 3: Конфигурация vercel.json

Создай файл `vercel.json` в корне проекта:

\`\`\`json
{
  "functions": {
    "api/**/*.go": {
      "runtime": "go1.x"
    }
  },
  "env": {
    "SUPABASE_URL": "@supabase_url",
    "SUPABASE_SERVICE_ROLE_KEY": "@supabase_service_role_key",
    "JWT_SECRET": "@jwt_secret",
    "POSTGRES_URL": "@postgres_url"
  }
}
\`\`\`

## Шаг 4: Создание Go модулей

### 4.1. Инициализация go.mod
\`\`\`bash
go mod init github.com/yourusername/fitness-app
\`\`\`

### 4.2. Основные зависимости
\`\`\`bash
go get github.com/golang-jwt/jwt/v5
go get github.com/lib/pq
go get golang.org/x/crypto/bcrypt
\`\`\`

## Шаг 5: Пример API эндпоинта (Register)

\`\`\`go
// api/register.go
package api

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "os"
    
    _ "github.com/lib/pq"
    "golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
    Username  string `json:"username"`
    Email     string `json:"email"`
    Password  string `json:"password"`
    FirstName string `json:"first_name,omitempty"`
    LastName  string `json:"last_name,omitempty"`
    Phone     string `json:"phone,omitempty"`
    IsTrainer bool   `json:"is_trainer,omitempty"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
    // CORS headers
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    
    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }
    
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate required fields
    if req.Username == "" || req.Email == "" || req.Password == "" {
        http.Error(w, "Username, email and password are required", http.StatusBadRequest)
        return
    }
    
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Error hashing password", http.StatusInternalServerError)
        return
    }
    
    // Connect to database
    db, err := sql.Open("postgres", os.Getenv("POSTGRES_URL"))
    if err != nil {
        http.Error(w, "Database connection error", http.StatusInternalServerError)
        return
    }
    defer db.Close()
    
    // Insert user
    var userID string
    query := `
        INSERT INTO users (username, email, password_hash, first_name, last_name, phone, is_trainer)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `
    err = db.QueryRow(query, req.Username, req.Email, string(hashedPassword), 
        req.FirstName, req.LastName, req.Phone, req.IsTrainer).Scan(&userID)
    
    if err != nil {
        http.Error(w, "Error creating user: "+err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Return success response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "user_id": userID,
        "message": "User registered successfully",
    })
}
\`\`\`

## Шаг 6: Деплой на Vercel

### 6.1. Через Vercel CLI
\`\`\`bash
# Установка Vercel CLI
npm i -g vercel

# Логин
vercel login

# Деплой
vercel
\`\`\`

### 6.2. Через GitHub (рекомендуется)
1. Создай GitHub репозиторий
2. Запушь код в репозиторий
3. Зайди на vercel.com
4. Нажми "Import Project"
5. Выбери свой GitHub репозиторий
6. Vercel автоматически определит Go проект
7. Добавь environment variables:
   - `JWT_SECRET` (создай свой секретный ключ)
   - Переменные из Supabase уже доступны в этом проекте v0

## Шаг 7: Настройка Environment Variables на Vercel

В настройках проекта на Vercel добавь:
- `JWT_SECRET` - для подписи JWT токенов (минимум 32 символа)
- Остальные переменные (SUPABASE_URL, POSTGRES_URL и т.д.) уже доступны из интеграции

## Шаг 8: Тестирование API

### 8.1. Регистрация пользователя
\`\`\`bash
curl -X POST https://your-project.vercel.app/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "securepassword123",
    "first_name": "Test",
    "last_name": "User",
    "is_trainer": false
  }'
\`\`\`

### 8.2. Ожидаемый ответ
\`\`\`json
{
  "success": true,
  "user_id": "uuid-here",
  "message": "User registered successfully"
}
\`\`\`

### 8.3. Логин пользователя
\`\`\`bash
curl -X POST https://your-project.vercel.app/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "securepassword123"
  }'
\`\`\`

### 8.4. Проверка профиля (с JWT токеном)
\`\`\`bash
curl -X GET https://your-project.vercel.app/api/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
\`\`\`

## Шаг 9: Проверка базы данных

После регистрации проверь данные в Supabase:
1. Зайди в Supabase Dashboard
2. Выбери Table Editor
3. Открой таблицу `users`
4. Проверь, что новый пользователь добавлен с хешированным паролем

## Важные замечания

### Безопасность
- Никогда не храни пароли в открытом виде
- Используй bcrypt для хеширования (уже в примере)
- JWT_SECRET должен быть длинным и случайным
- Используй HTTPS (Vercel предоставляет автоматически)

### CORS
- Настрой правильные CORS заголовки для твоего фронтенда
- В продакшене замени `*` на конкретный домен

### Rate Limiting
- Добавь rate limiting для защиты от brute-force атак
- Vercel предоставляет базовые лимиты по умолчанию

### Логирование
- Используй `log.Printf()` для отладки
- Логи доступны в Vercel Dashboard → Logs

## Следующие шаги

1. Добавь endpoint для login с генерацией JWT
2. Создай middleware для проверки JWT токенов
3. Добавь refresh token механизм
4. Реализуй эндпоинты для обновления профиля
5. Добавь эндпоинты для связи тренер-клиент

## Полезные ссылки

- [Vercel Go Runtime](https://vercel.com/docs/functions/runtimes/go)
- [Supabase Postgres](https://supabase.com/docs/guides/database)
- [JWT для Go](https://github.com/golang-jwt/jwt)
- [bcrypt для Go](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
