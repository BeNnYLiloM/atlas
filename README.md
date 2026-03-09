# Atlas (Corp Messenger) - Техническая Документация

## 1. Обзор проекта

Atlas — это корпоративный мессенджер с двумя моделями развёртывания:
- **SaaS** — облачная версия (как Slack/Discord)
- **On-premise** — развёртывание в закрытом контуре (Self-hosted)

Ключевые особенности: текстовые чаты, треды, иерархия воркспейсов и качественная аудио/видео связь.

**Клиенты:**
- Веб-приложение (Vue 3)
- Десктоп (Electron/Tauri — в планах)
- Мобильные приложения iOS/Android (Flutter — в планах)

## 2. Технологический стек

| Уровень | Технология | Обоснование |
|---------|------------|-------------|
| Backend | Go (Golang) 1.21+ | Высокая производительность, работа с WebSockets (10k+ соединений) |
| Frontend | Vue 3 + Vite + Pinia | Реактивность, скорость разработки, Composition API |
| Styling | Tailwind CSS | Ускоренная вёрстка, Dark/Light mode |
| Database | PostgreSQL 15+ | Надёжное хранение реляционных данных |
| Cache | Redis 7 | Хранение сессий, pub/sub для чатов |
| Files | MinIO | S3-совместимое хранилище (локальное) |
| Voice/Video | LiveKit | SFU сервер для WebRTC (конференции) |
| Mobile | Flutter | Единая кодовая база для iOS и Android |

## 3. Структура проекта (Monorepo)

```
atlas/
├── deploy/                       # Инфраструктура
│   ├── docker-compose.yml        # Production
│   ├── docker-compose.dev.yml    # Локальная разработка
│   └── configs/                  # Конфиги (livekit.yaml, nginx.conf)
│
├── backend/                      # Серверная часть (Go)
│   ├── cmd/
│   │   └── server/
│   │       └── main.go           # Точка входа
│   ├── internal/
│   │   ├── config/               # Загрузка конфигурации
│   │   ├── domain/               # Сущности (User, Message, Channel)
│   │   ├── repository/           # Работа с БД (PostgreSQL)
│   │   ├── service/              # Бизнес-логика
│   │   └── transport/
│   │       ├── http/             # REST API хендлеры
│   │       └── ws/               # WebSocket хендлеры
│   ├── migrations/               # SQL миграции
│   └── pkg/                      # Общие пакеты (logger, validator)
│
├── frontend/                     # Клиентская часть (Vue 3)
│   └── src/
│       ├── api/                  # HTTP/WebSocket клиенты
│       ├── assets/               # Статика (изображения, шрифты)
│       ├── components/           # Vue компоненты
│       ├── composables/          # Vue composables
│       ├── stores/               # Pinia stores
│       └── views/                # Страницы
│
├── mobile/                       # Мобильное приложение (Flutter)
│   └── .gitkeep
│
├── docs/                         # Документация
│   └── api/                      # OpenAPI/Swagger спецификации
│
├── scripts/                      # Скрипты автоматизации
│
├── .env.example                  # Шаблон переменных окружения
├── Makefile                      # Команды сборки/запуска
├── README.md
└── SPECIFICATION.md              # Техническое задание
```

## 4. Схема Базы Данных (MVP)

SQL-схема для первой версии (миграция `001_init.sql`).

```sql
-- 1. Пользователи
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 2. Воркспейсы (Область видимости компании)
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    owner_id UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 3. Участники воркспейса
CREATE TABLE workspace_members (
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member', -- admin, moderator, member
    PRIMARY KEY (workspace_id, user_id)
);

-- 4. Каналы
CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(20) DEFAULT 'text', -- text, voice
    is_private BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Сообщения
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID REFERENCES channels(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    content TEXT,
    parent_id UUID REFERENCES messages(id), -- Для тредов
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE
);
```

## 5. API Endpoints (Черновик)

### Auth
- `POST /api/v1/auth/register` — Регистрация
- `POST /api/v1/auth/login` — Авторизация

### Chat
- `GET /api/v1/workspaces` — Список воркспейсов
- `GET /api/v1/channels/{id}/messages` — Сообщения канала

### WebSocket
- `WS /ws/connect` — Соединение для real-time событий

## 6. Быстрый старт

### Запуск инфраструктуры

```bash
# Поднять инфраструктуру (PostgreSQL, Redis, MinIO, LiveKit)
make dev-up

# Запустить backend (в отдельном терминале)
make backend

# Запустить frontend (в отдельном терминале)
make frontend
```

**Порты:**
- Backend API: http://localhost:8080
- Frontend: http://localhost:3000
- Adminer (БД): http://localhost:8081
- MinIO Console: http://localhost:9001

### Просмотр базы данных

#### Через Adminer (Web UI)

1. Откройте http://localhost:8081
2. Введите данные для подключения:
   - Система: **PostgreSQL**
   - Сервер: **atlas-postgres**
   - Пользователь: **atlas**
   - Пароль: **atlas**
   - База данных: **atlas**

#### Через командную строку

```powershell
# Запустить интерактивный скрипт
pwsh scripts/db-view.ps1

# Или напрямую через Docker
docker exec atlas-postgres psql -U atlas -d atlas -c "SELECT * FROM users;"
```

Подробнее: [scripts/db-commands.md](scripts/db-commands.md)

### Остановка инфраструктуры

```bash
make dev-down
```
