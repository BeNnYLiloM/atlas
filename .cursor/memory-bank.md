# Atlas — Memory Bank

## Что это за проект

Корпоративный мессенджер (аналог Slack/Discord) с текстовыми и голосовыми каналами, тредами, задачами, реакциями, поиском, загрузкой файлов. Монорепо: `backend/` (Go) + `frontend/` (Vue 3).

---

## Структура директорий

```
atlas/
├── backend/
│   ├── cmd/server/main.go              # точка входа
│   ├── internal/
│   │   ├── config/config.go            # конфиг из env
│   │   ├── domain/                     # модели
│   │   ├── repository/                 # интерфейсы + postgres
│   │   ├── service/                    # бизнес-логика
│   │   └── transport/
│   │       ├── http/handler/           # Gin handlers
│   │       ├── http/middleware/        # auth, cors
│   │       ├── http/response/          # helpers
│   │       └── ws/                     # WebSocket hub + client
│   ├── migrations/                     # 9 SQL файлов
│   ├── pkg/
│   │   ├── database/postgres.go        # pgx pool
│   │   └── storage/minio.go            # MinIO клиент
│   └── go.mod
├── frontend/
│   └── src/
│       ├── api/                        # 10 axios-клиентов
│       ├── components/
│       │   ├── calls/                  # CallPanel
│       │   ├── chat/                   # сообщения, треды, ввод
│       │   ├── layout/                 # Sidebar, ChannelList
│       │   ├── search/                 # SearchBar, SearchResults
│       │   ├── settings/               # ThemeSwitcher, ShortcutsModal
│       │   ├── tasks/                  # TaskCard, TaskCreationModal
│       │   ├── ui/                     # Modal, Button, Input, Avatar
│       │   └── workspace/              # InviteMemberModal
│       ├── router/index.ts
│       ├── stores/                     # 9 Pinia stores
│       ├── types/index.ts
│       └── views/                      # 5 views
└── deploy/
    ├── docker-compose.yml
    └── configs/
        ├── livekit.yaml               # для Docker (redis: redis:6379)
        └── livekit-native.yaml        # для нативного запуска (redis: localhost:6379)
```

---

## Backend — Доменные модели

### User (`domain/user.go`)
```go
type UserStatus string // "online" | "away" | "offline"

type User struct {
    ID           string     `db:"id"`
    Email        string     `db:"email"`
    PasswordHash string     `db:"password_hash"`
    DisplayName  string     `db:"display_name"`
    AvatarURL    *string    `db:"avatar_url"`
    Status       UserStatus `db:"status"`
    LastSeen     *time.Time `db:"last_seen"`
    CreatedAt    time.Time  `db:"created_at"`
}
```

### Workspace (`domain/workspace.go`)
```go
type WorkspaceMember struct {
    WorkspaceID string // роли: owner/admin/moderator/member/guest
    UserID      string
    Role        string
}
```

### Channel (`domain/channel.go`)
```go
type Channel struct {
    ID          string // тип: text/voice
    WorkspaceID string
    Name        string
    Type        string  // "text" | "voice"
    IsPrivate   bool
}
```

### Message (`domain/message.go`)
```go
type Message struct {
    ID                 string
    ChannelID          string
    UserID             string
    Content            string
    ParentID           *string  // для тредов
    HasAttachments     bool
    SearchVector       string
    CreatedAt          time.Time
    UpdatedAt          time.Time
    // Computed fields
    Author             *User
    ThreadRepliesCount int
    ThreadUnreadCount  int
}
```

### Task (`domain/task.go`)
```go
type TaskStatus   string // "todo" | "in_progress" | "done" | "cancelled"
type TaskPriority string // "low" | "medium" | "high" | "urgent"

type Task struct {
    ID          string
    MessageID   *string  // привязка к сообщению (killer-feature)
    WorkspaceID string
    Title       string
    Description *string
    Status      TaskStatus
    Priority    TaskPriority
    AssigneeID  *string
    ReporterID  *string
    DueDate     *time.Time
    Assignee    *User    // display only
    Reporter    *User    // display only
}
```

### Reaction (`domain/reaction.go`)
```go
type Reaction struct {
    MessageID string
    UserID    string
    Emoji     string
}
type ReactionGroup struct {
    Emoji   string
    Count   int
    UserIDs []string
    Mine    bool  // поставил ли текущий пользователь
}
```

### File (`domain/file.go`)
```go
type File struct {
    ID           string
    MessageID    *string
    UserID       string
    Filename     string
    OriginalName string
    MimeType     string
    SizeBytes    int64
    StoragePath  string
    URL          string  // генерируется динамически (presigned)
}
```

---

## Backend — Конфиг (`internal/config/config.go`)

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    MinIO    MinIOConfig
    LiveKit  LiveKitConfig
    JWT      JWTConfig
}

type LiveKitConfig struct {
    Host      string  // env: LIVEKIT_HOST (default: localhost:7880)
    URL       string  // env: LIVEKIT_URL  (default: ws://localhost:7880)
    APIKey    string  // env: LIVEKIT_API_KEY
    APISecret string  // env: LIVEKIT_API_SECRET
}

type JWTConfig struct {
    Secret     string // env: JWT_SECRET
    ExpireHour int    // 24
}
```

---

## Backend — HTTP Routes

Все роуты под `/api/v1`. Защищённые (`authMiddleware`) идут через `protected` группу.

```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
GET    /api/v1/auth/me                    [auth]
GET    /api/v1/users/search?email=        [auth]

POST   /api/v1/workspaces                 [auth]
GET    /api/v1/workspaces                 [auth]
GET    /api/v1/workspaces/:id             [auth]
DELETE /api/v1/workspaces/:id             [auth]
GET    /api/v1/workspaces/:id/members     [auth]
POST   /api/v1/workspaces/:id/members     [auth]
GET    /api/v1/workspaces/:id/channels    [auth]

POST   /api/v1/channels                   [auth]
GET    /api/v1/channels/:id               [auth]
PUT    /api/v1/channels/:id               [auth]
DELETE /api/v1/channels/:id               [auth]
POST   /api/v1/channels/:id/read          [auth]
GET    /api/v1/channels/:id/messages      [auth]

POST   /api/v1/messages                   [auth]
PUT    /api/v1/messages/:id               [auth]
DELETE /api/v1/messages/:id               [auth]
GET    /api/v1/messages/:id/thread        [auth]
POST   /api/v1/messages/:id/thread/read   [auth]
GET    /api/v1/messages/:id/thread/unread [auth]

POST   /api/v1/messages/:id/reactions     [auth]
DELETE /api/v1/messages/:id/reactions/:emoji [auth]
GET    /api/v1/messages/:id/reactions     [auth]

POST   /api/v1/tasks                      [auth]
GET    /api/v1/tasks?workspace_id=&status= [auth]
PATCH  /api/v1/tasks/:id                  [auth]
DELETE /api/v1/tasks/:id                  [auth]

POST   /api/v1/calls/join                 [auth]

POST   /api/v1/files/upload               [auth, если MinIO доступен]
GET    /api/v1/files/:id                  [auth, если MinIO доступен]
DELETE /api/v1/files/:id                  [auth, если MinIO доступен]

GET    /api/v1/search?q=&workspace_id=    [auth]

GET    /ws?token=                         # WebSocket (не в /api/v1)
GET    /health
```

### Важно: извлечение userID в handlers
```go
// ПРАВИЛЬНО (все handlers):
userID := middleware.GetUserID(c)

// НЕПРАВИЛЬНО (исторически было, уже исправлено):
userID := c.GetString("user_id")  // ключ "userID", не "user_id"!
```

---

## Backend — WebSocket

### Формат сообщений
```json
// Входящие (клиент → сервер):
{"event": "subscribe_workspace", "data": {"workspace_id": "uuid"}}
{"event": "subscribe", "data": {"channel_id": "uuid"}}
{"event": "typing", "data": {"channel_id": "uuid", "typing": true}}

// Исходящие (сервер → клиент):
{"type": "message", "payload": {...}}
{"type": "presence", "payload": {"user_id": "uuid", "status": "online"}}
{"type": "typing", "payload": {"user_id": "uuid", "channel_id": "uuid", "typing": true}}
```

### Все WS события (тип → обработчик в frontend)
| Тип | Когда |
|---|---|
| `message` | новое сообщение |
| `thread_reply` | ответ в тред |
| `message_updated` | редактирование |
| `message_deleted` | удаление |
| `channel_created` | новый канал |
| `channel_updated` | изменение канала |
| `channel_deleted` | удаление канала |
| `member_added` | новый участник workspace |
| `member_removed` | участник удалён |
| `member_updated` | смена роли |
| `typing` | индикатор печати |
| `presence` | статус online/offline |
| `reaction_added` | добавлена реакция |
| `reaction_removed` | убрана реакция |
| `read_state_update` | прочитано |

### Hub — структуры
```go
type Hub struct {
    workspaces map[string]map[string]*Client  // workspaceID → clientID → Client
    channels   map[string]map[string]*Client  // channelID → clientID → Client
    users      map[string][]*Client           // userID → []Client (мультитаб)
    mu         sync.RWMutex
}
```

---

## Backend — Сервисы: ключевые особенности

### AuthService
- bcrypt для паролей
- JWT HS256, access token 24h
- `GetUserByID` используется в `CallsHandler` для получения DisplayName

### WorkspaceService
- При создании workspace автоматически создаётся канал "general"
- `AddMember` — только owner/admin могут добавлять

### MessageService
- Проверяет членство пользователя в канале перед созданием
- Загружает `Author` и `ThreadRepliesCount` для каждого сообщения
- Update/Delete — только автор (или модератор+ для Delete)

### ReactionService
```go
// Использует интерфейс для избежания циклического импорта
type ReactionBroadcaster interface {
    BroadcastToWorkspace(workspaceID, event string, data interface{}, excludeUserID string)
}
```

### LiveKitService
```go
// Стабильное имя комнаты — все входят в одну комнату
func (s *LiveKitService) CreateRoomName(channelID string) string {
    return fmt.Sprintf("channel-%s", channelID)  // без UUID!
}
// URL берётся из конфига (не строится с wss://)
URL: s.cfg.URL  // ws://localhost:7880 для dev
```

---

## Frontend — Stores

### `auth` store
```typescript
// Ключи в localStorage:
// "atlas_token" — JWT токен
// Redirect на /login при 401 (в axios interceptor)
```

### `calls` store (критически важный)
```typescript
// SDK: livekit-client@1.15.13 (v1, совместим с сервером v1.9.x)
// НЕ использовать v2 SDK — сервер v1.9.11 не поддерживает /rtc/v2

// Имена методов в v1 SDK:
// room.participants (Map) — не remoteParticipants!
// p.name || p.identity — имя участника

// Аудио треки ТРЕБУЮТ attach к DOM:
const el = track.attach()
getAudioContainer().appendChild(el)

// Discord-логика:
toggleVoiceChannel(channelId) // войти или выйти
isInChannel(channelId): bool  // проверка текущего канала
```

### `websocket` store
```typescript
// Reconnect: exponential backoff, max 5 попыток
// Подписка на workspace отправляется при connect
// wsStore.subscribeToWorkspace(workspaceId) — вызывается из AppView при смене workspace
```

### `channels` store
```typescript
// typingUsers: Record<channelId, Set<userId>>
setUserTyping(channelId, userId, isTyping)
getTypingUsers(channelId): string[]
```

### `workspace` store
```typescript
// presenceMap: Record<userId, status>
// membersMap: Record<workspaceId, WorkspaceMember[]>
setPresence(userId, status)
getPresence(userId): string
addMember / removeMember / updateMemberRole
```

---

## Frontend — Компоненты: что где

### Голосовые каналы
- `ChannelList.vue` — клик по voice → `callsStore.toggleVoiceChannel(id)`
- `CallPanel.vue` — встроен в `Sidebar.vue` (не floating!), показывает имя канала + mute + выйти
- Участники канала отображаются прямо под каналом в сайдбаре

### Сообщения
- `MessageInput.vue` — отправка по Enter (Shift+Enter = перенос), drag&drop файлов
- `MessageItem.vue` — hover → кнопки действий (edit/delete/react/create task)
- `ReactionBar.vue` — QUICK_EMOJIS = 10 эмодзи, toggle реакций
- `TypingIndicator.vue` — фильтрует self из списка typing users

### Layout
- `AppView.vue` — root для авт. пользователей, инициализирует WS, следит за workspace
- Тема (dark/light/high-contrast) через `data-theme` атрибут на `<html>`

---

## Frontend — API клиент

```typescript
// baseURL: http://localhost:8080/api/v1
// Токен: "atlas_token" из localStorage → "Bearer {token}"
// 401 → router.push('/login') + localStorage.removeItem('atlas_token')
```

---

## База данных — Миграции

| Файл | Ключевое |
|---|---|
| `001_init.sql` | users, workspaces, workspace_members, channels, messages |
| `002_channel_members.sql` | channel_members + 2 триггера для автодобавления |
| `003_thread_read_state.sql` | channel_members.thread_last_reads JSONB |
| `004_user_presence.sql` | users.status, users.last_seen |
| `005_files.sql` | files, messages.has_attachments |
| `006_search.sql` | messages.search_vector, GIN index, trigger (russian dictionary) |
| `007_reactions.sql` | message_reactions, UNIQUE(message_id, user_id, emoji) |
| `008_tasks.sql` | tasks (с message_id FK для killer-feature) |

### Применение новых миграций (PowerShell):
```powershell
# НЕ использовать < для pipe в PowerShell!
Get-Content backend/migrations/00X_name.sql | docker exec -i atlas-postgres psql -U atlas -d atlas
```

---

## Docker Compose

```yaml
# Файл: deploy/docker-compose.yml
services:
  postgres:   # порт 5433:5432 (!) — не стандартный 5432
  redis:      # порт 6379:6379
  minio:      # порт 9000 (API), 9001 (console)
  livekit:    # порт 7880 (HTTP/WS), 7881 (TCP RTC), 7882/udp
```

### Запуск всего:
```powershell
cd deploy
docker-compose up -d
```

### Порты PostgreSQL
- В Docker: `5433:5432` (внешний 5433 → внутренний 5432)
- Строка подключения backend: `DB_HOST=localhost DB_PORT=5433`

---

## LiveKit — Конфигурация

### Для локальной разработки (deploy/configs/livekit.yaml)
```yaml
port: 7880
rtc:
  tcp_port: 7881
  node_ip: 127.0.0.1     # браузер подключается через localhost
  port_range_start: 50000
  port_range_end: 60000
redis:
  address: redis:6379    # имя сервиса в Docker сети
keys:
  devkey: secret_replace_in_production_min_32_chars!!
```

### Для прода
- Заменить `node_ip` на публичный IP сервера
- Или `use_external_ip: true`
- Сгенерировать длинный секрет: `openssl rand -hex 32`

### Совместимость версий (ВАЖНО!)
- Сервер: `livekit/livekit-server:v1.9.11`
- SDK: `livekit-client@1.15.13` (v1.x — НЕ v2!)
- SDK v2 пробует `/rtc/v2` → 404 → retry → работает, но с ошибками в консоли

---

## Запуск для разработки

```powershell
# 1. Docker сервисы
cd deploy
docker-compose up -d

# 2. Backend
cd backend
go run ./cmd/server/...
# Слушает: http://localhost:8080

# 3. Frontend
cd frontend
npm run dev
# Слушает: http://localhost:5173
```

---

## Известные проблемы и решения

### Docker Desktop зависает при старте LiveKit
- Перезапустить Docker Desktop
- Убедиться что порт 7880 свободен перед `docker-compose up`
- ```powershell
  $pids = (Get-NetTCPConnection -LocalPort 7880 -ErrorAction SilentlyContinue).OwningProcess
  $pids | ForEach-Object { Stop-Process -Id $_ -Force }
  ```

### Порт 8080 занят
```powershell
$pids = (Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue).OwningProcess | Where-Object { $_ -ne 0 -and $_ -ne 4 } | Select-Object -Unique
$pids | ForEach-Object { Stop-Process -Id $_ -Force }
```

### PowerShell pipe вместо `<`
```powershell
# НЕ работает в PowerShell:
psql < file.sql

# Работает:
Get-Content file.sql | docker exec -i atlas-postgres psql -U atlas -d atlas
```

### Cyclic import в Go (service ↔ ws)
- `ReactionService` использует интерфейс `ReactionBroadcaster` вместо `*ws.Hub`
- Паттерн для всех новых сервисов, которым нужен broadcast

### Нет звука в голосовых каналах
- LiveKit v1 требует явного attach треков к DOM
- Аудио элементы помещаются в скрытый `<div id="livekit-audio-container">`
- Attach при `TrackSubscribed` + при входе в активную комнату (`room.participants`)

---

## Зависимости

### Backend (ключевые)
```
github.com/gin-gonic/gin          v1.9.1
github.com/golang-jwt/jwt/v5      v5.2.0
github.com/google/uuid            v1.6.0
github.com/gorilla/websocket      v1.5.4
github.com/jackc/pgx/v5           v5.5.1
golang.org/x/crypto               v0.47.0
github.com/livekit/protocol       v1.44.0
github.com/minio/minio-go/v7      v7.0.98
```

### Frontend (ключевые)
```
vue                  ^3.4.21
pinia                ^2.1.7
vue-router           ^4.3.0
axios                ^1.6.8
livekit-client       ^1.15.13   # v1! не v2
@vueuse/core         ^10.9.0
tailwindcss          ^3.4.1
typescript           ~5.4.0
vite                 ^5.2.0
```

---

## Горячие клавиши (ShortcutsModal)

| Клавиша | Действие |
|---|---|
| `Ctrl+K` | Открыть поиск |
| `Ctrl+/` | Показать горячие клавиши |
| `Alt+↑/↓` | Переключение текстовых каналов |
| `Shift+Enter` | Новая строка в сообщении |
| `Esc` | Закрыть модальное окно |

---

## Темы

Файлы в `frontend/src/assets/themes/`:
- `dark.css` — тёмная (по умолчанию)
- `light.css` — светлая
- `high-contrast.css` — высокий контраст

Применяются через `data-theme` на `<html>`, сохраняются в localStorage `atlas_theme`.
