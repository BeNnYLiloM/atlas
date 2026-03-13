# Atlas — Memory Bank

## Философия проекта

**Это не MVP. Код пишется сразу в production-качество.**

- Нет "сделаем потом", нет "для MVP сойдёт"
- Каждое изменение — как будто это уже работает под реальными пользователями
- Безопасность, надёжность, наблюдаемость — с первого коммита
- Никаких memory leak, глобальных `log.Printf`, наивных реализаций

---

## Что это за проект

Корпоративный мессенджер (аналог Slack/Discord) с текстовыми и голосовыми каналами, тредами, задачами, реакциями, поиском, загрузкой файлов. Монорепо: `backend/` (Go) + `frontend/` (Vue 3).

---

## Структура директорий

```
atlas/
├── backend/
│   ├── cmd/server/main.go              # точка входа
│   ├── cmd/debugrefresh/main.go        # УДАЛИТЬ — debug-скрипт, не нужен в репо
│   ├── internal/
│   │   ├── config/config.go            # конфиг из env
│   │   ├── domain/                     # модели
│   │   ├── repository/                 # интерфейсы + postgres
│   │   ├── service/                    # бизнес-логика
│   │   └── transport/
│   │       ├── http/handler/           # Gin handlers
│   │       ├── http/middleware/        # auth, cors, rate_limit
│   │       ├── http/response/          # helpers
│   │       └── ws/                     # WebSocket hub + client
│   ├── migrations/                     # 19 SQL файлов (000001–000019)
│   ├── pkg/
│   │   ├── database/postgres.go        # pgx pool
│   │   └── storage/minio.go            # MinIO клиент
│   └── go.mod
├── frontend/
│   └── src/
│       ├── api/
│       │   ├── client.ts               # axios + interceptors (in-memory token, авто-refresh)
│       │   ├── session.ts              # управление access/refresh токенами
│       │   ├── auth.ts
│       │   ├── users.ts                # searchByEmail, updateMe, uploadAvatar
│       │   └── ...
│       ├── components/
│       │   ├── calls/                  # CallPanel
│       │   ├── chat/                   # сообщения, треды, ввод
│       │   ├── layout/                 # Sidebar, ChannelList
│       │   ├── search/                 # SearchBar, SearchResults
│       │   ├── settings/               # ThemeSwitcher (карточный UI), ShortcutsModal
│       │   ├── tasks/                  # TaskCard, TaskCreationModal
│       │   ├── ui/                     # Modal, Button, Input, Avatar
│       │   └── workspace/              # InviteMemberModal, UserSettingsModal, WorkspaceSettingsModal
│       ├── router/index.ts
│       ├── stores/                     # Pinia stores
│       ├── types/index.ts
│       └── views/                      # views
└── deploy/
    ├── docker-compose.yml              # dev
    ├── docker-compose.coolify.yml      # prod (Coolify)
    └── configs/
        ├── livekit.yaml               # для Docker (redis: redis:6379)
        └── livekit.coolify.yaml        # для прода
```

---

## Система авторизации (после рефакторинга e3f8c5d)

### Архитектура токенов
- **Access token** — JWT, TTL 15 минут, хранится in-memory (НЕ localStorage — уязвим к XSS)
- **Refresh token** — случайные 32 байта → SHA-256 → в БД; отдаётся в httpOnly cookie
- **Ротация сессий** — каждый refresh создаёт новую запись в `auth_sessions`, старая помечается `revoked_at` и `replaced_by_session_id`
- **Обнаружение reuse-атаки** — если revoked-токен используется повторно, вся family сессий отзывается

### JWT Claims
```go
type Claims struct {
    UserID    string `json:"user_id"`
    Email     string `json:"email"`
    SessionID string `json:"session_id"`  // позволяет инвалидировать конкретные токены
    jwt.RegisteredClaims  // iss, aud, exp, iat
}
```
- Проверяются `iss` (atlas), `aud` (atlas-web), `alg` (только HMAC — закрыта атака `alg: none`)

### JWTConfig (актуальный)
```go
type JWTConfig struct {
    Secret                string
    AccessTokenTTLMinutes int    // env: JWT_ACCESS_TTL_MINUTES, default: 15
    RefreshTokenTTLDays   int    // env: JWT_REFRESH_TTL_DAYS, default: 14
    Issuer                string // env: JWT_ISSUER, default: "atlas"
    Audience              string // env: JWT_AUDIENCE, default: "atlas-web"
    RefreshCookieName     string // env: JWT_REFRESH_COOKIE_NAME
    RefreshCookieDomain   string // env: JWT_REFRESH_COOKIE_DOMAIN
    RefreshCookieSecure   bool   // env: JWT_REFRESH_COOKIE_SECURE (auto: production mode)
}
```

### AuthService (сигнатуры после рефакторинга)
```go
func NewAuthService(userRepo, sessionRepo, jwtConfig) *AuthService
func (s) Register(ctx, input UserCreate, meta AuthSessionMetadata) (*User, *TokenPair, string, error)
func (s) Login(ctx, input UserLogin, meta AuthSessionMetadata) (*User, *TokenPair, string, error)
func (s) Refresh(ctx, refreshToken string, meta AuthSessionMetadata) (*TokenPair, string, error)
func (s) Logout(ctx, refreshToken string) error
func (s) LogoutAll(ctx, userID string) error
func (s) UpdateProfile(ctx, userID string, input UserUpdate) (*User, error)
```

### TokenPair (актуальный — без refresh_token!)
```go
type TokenPair struct {
    AccessToken string `json:"access_token"`
    ExpiresAt   int64  `json:"expires_at"`
    // RefreshToken УБРАН — теперь только в httpOnly cookie
}
```

### AuthSessionMetadata
```go
type AuthSessionMetadata struct {
    UserAgent string
    IPAddress string
}
```

### Frontend — session.ts
```typescript
// Access token хранится в ref<string | null> (in-memory, не localStorage)
// refreshSkewSeconds = 30 — упреждающее обновление за 30с до истечения
// refreshPromise — дедупликация параллельных запросов refresh
export function getAccessToken(): string | null
export function applyAuthTokens(tokens: AuthTokens): void
export async function refreshAccessToken(): Promise<string | null>
export async function ensureAccessToken(): Promise<string | null>
```

### Frontend — api/client.ts
```typescript
// 401 interceptor:
// 1. Не retry для /auth/login, /auth/register, /auth/refresh, /auth/logout
// 2. При 401 → refreshAccessToken() → retry с новым токеном
// 3. Если refresh не удался → clearAccessToken() → redirect /login
// _retry флаг предотвращает бесконечные циклы
```

---

## Доменные модели

### User
```go
type User struct {
    ID           string
    Email        string
    PasswordHash string
    DisplayName  string
    AvatarURL    *string
    Status       UserStatus  // "online" | "away" | "offline"
    LastSeen     *time.Time
    CreatedAt    time.Time
}

type UserUpdate struct {
    DisplayName *string `json:"display_name" validate:"omitempty,min=2,max=100"`
    AvatarURL   *string `json:"avatar_url"`
}
```

### AuthSession (`domain/auth_session.go`)
```go
type AuthSession struct {
    ID                  string
    FamilyID            string     // для группового отзыва при reuse-атаке
    UserID              string
    RefreshTokenHash    string     // SHA-256 от сырого токена
    UserAgent           string
    IPAddress           string
    CreatedAt           time.Time
    ExpiresAt           time.Time
    LastUsedAt          *time.Time
    RevokedAt           *time.Time
    ReplacedBySessionID *string
}
```

### Task
```go
type Task struct {
    ID          string
    MessageID   *string  // killer-feature: привязка к сообщению
    WorkspaceID string
    Title, Description, Status, Priority string
    AssigneeID, ReporterID *string
    DueDate     *time.Time
    Assignee, Reporter *User
}
```

---

## Backend — HTTP Routes (актуальные)

```
POST   /api/v1/auth/register
POST   /api/v1/auth/login
POST   /api/v1/auth/refresh
POST   /api/v1/auth/logout
POST   /api/v1/auth/logout-all           [auth]
GET    /api/v1/auth/me                   [auth]
PATCH  /api/v1/auth/me                   [auth] — UpdateMe (display_name, avatar_url)
POST   /api/v1/auth/me/avatar            [auth] — UploadAvatar (multipart)
GET    /api/v1/users/search?email=       [auth]

POST   /api/v1/workspaces                [auth]
GET    /api/v1/workspaces                [auth]
GET    /api/v1/workspaces/:id            [auth]
DELETE /api/v1/workspaces/:id            [auth]
GET    /api/v1/workspaces/:id/members    [auth]
POST   /api/v1/workspaces/:id/members    [auth]
GET    /api/v1/workspaces/:id/channels   [auth]

POST   /api/v1/channels                  [auth]
GET    /api/v1/channels/:id              [auth]
PUT    /api/v1/channels/:id              [auth]
DELETE /api/v1/channels/:id              [auth]
POST   /api/v1/channels/:id/read         [auth]
GET    /api/v1/channels/:id/messages     [auth]

POST   /api/v1/messages                  [auth]
PUT    /api/v1/messages/:id              [auth]
DELETE /api/v1/messages/:id              [auth]
GET    /api/v1/messages/:id/thread       [auth]
POST   /api/v1/messages/:id/thread/read  [auth]
GET    /api/v1/messages/:id/thread/unread [auth]

POST   /api/v1/messages/:id/reactions    [auth]
DELETE /api/v1/messages/:id/reactions/:emoji [auth]
GET    /api/v1/messages/:id/reactions    [auth]

POST   /api/v1/tasks                     [auth]
GET    /api/v1/tasks?workspace_id=&status= [auth]
PATCH  /api/v1/tasks/:id                 [auth]
DELETE /api/v1/tasks/:id                 [auth]

POST   /api/v1/calls/join                [auth]

POST   /api/v1/files/upload              [auth]
GET    /api/v1/files/:id                 [auth]
DELETE /api/v1/files/:id                 [auth]

GET    /api/v1/search?q=&workspace_id=   [auth]

GET    /ws                               # WebSocket (Bearer в query ?token= или header)
GET    /health
```

### Важно: извлечение userID
```go
userID := middleware.GetUserID(c)  // ПРАВИЛЬНО
// c.GetString("user_id") — НЕПРАВИЛЬНО (исторический баг, исправлен)
```

---

## Middleware

### CORS
```go
// allowlist-based, динамический origin, Vary: Origin
// Access-Control-Allow-Credentials: true (нужно для httpOnly cookie)
// Настраивается через CORS_ALLOWED_ORIGINS env (CSV)
CORS(cfg.Server.AllowedOrigins)
```

### Rate Limiter
```go
// ВНИМАНИЕ: текущая реализация в middleware/rate_limit.go имеет MEMORY LEAK
// buckets никогда не очищаются — нужно заменить на golang.org/x/time/rate
// или добавить периодический cleanup
NewRateLimiter(10, time.Minute)  // login: 10 req/min
NewRateLimiter(5, time.Minute)   // register: 5 req/min
```

### AuthMiddleware
```go
// Bearer token из Authorization header
// Проверяет ValidateToken → Claims → записывает userID в gin.Context
userID := middleware.GetUserID(c)
```

---

## Права доступа (service/access.go)

```go
// Централизованная проверка — используется в channel, message, task сервисах
func ensureWorkspaceMember(ctx, workspaceRepo, workspaceID, userID) (*WorkspaceMember, error)
func getAccessibleChannel(ctx, channelRepo, workspaceRepo, roleRepo, permRepo, channelID, userID) (*Channel, *WorkspaceMember, error)
// Логика: membership → роль (owner/admin проходят без доп. проверок) → ChannelPermission для приватных
```

---

## Архитектурные правила (production-quality)

### Go backend
- Dependency injection через конструкторы, никаких глобальных переменных
- Logger передаётся через DI, не `log.Printf` в сервисах
- Сервисы не знают о HTTP/cookie — это слой transport
- `fileService` в `AuthHandler` — ТЕХНИЧЕСКИЙ ДОЛГ, нужно вынести в отдельный `ProfileHandler` или `UserHandler`
- Cookie-getters (`RefreshCookieName()` и др.) в `AuthService` — ТЕХНИЧЕСКИЙ ДОЛГ, handler должен читать cfg напрямую
- Swagger-аннотации были удалены — если используется swaggo, нужно восстановить
- `cmd/debugrefresh/` — УДАЛИТЬ

### Frontend
- `UserUpdate` interface использует `snake_case` (для прямого JSON-маппинга) — технический долг, лучше `camelCase` + маппинг при отправке
- `extractApiError` в `users.ts` — нужно вынести в `api/utils.ts` (используется во всех api-модулях)
- Токены — in-memory only, никогда localStorage

---

## WebSocket

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

### Все WS события
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

---

## Frontend — Stores

### `auth` store
```typescript
// user: ref<User | null>
// updateProfile(data: UserUpdate): Promise<User>  — обновляет user реактивно
// uploadAvatar(file: File): Promise<User>          — обновляет user реактивно
// logout() — async, ждёт API
// logoutAll() — завершает все сессии
```

### `calls` store
```typescript
// SDK: livekit-client@1.15.13 (v1, совместим с сервером v1.9.x)
// НЕ использовать v2 SDK — сервер v1.9.11 не поддерживает /rtc/v2
// room.participants (Map) — не remoteParticipants!
// Аудио треки ТРЕБУЮТ attach к DOM:
const el = track.attach()
getAudioContainer().appendChild(el)
```

### `websocket` store
```typescript
// Reconnect: exponential backoff, max 5 попыток
// wsStore.subscribeToWorkspace(workspaceId) — вызывается из AppView при смене workspace
```

---

## База данных — Миграции

| Файл | Ключевое |
|---|---|
| `000001_init.sql` | users, workspaces, workspace_members, channels, messages |
| `000002_channel_members.sql` | channel_members + 2 триггера |
| `000003_thread_read_state.sql` | channel_members.thread_last_reads JSONB |
| `000004_user_presence.sql` | users.status, users.last_seen |
| `000005_files.sql` | files, messages.has_attachments |
| `000006_search.sql` | messages.search_vector, GIN index, trigger (russian) |
| `000007_reactions.sql` | message_reactions, UNIQUE(message_id, user_id, emoji) |
| `000008_tasks.sql` | tasks (с message_id FK) |
| `000019_auth_sessions.sql` | auth_sessions (id, family_id, user_id, refresh_token_hash, ...) |

### Применение миграций (PowerShell):
```powershell
# НЕ использовать < в PowerShell!
Get-Content backend/migrations/00X_name.sql | docker exec -i atlas-postgres psql -U atlas -d atlas
```

---

## Docker Compose

```yaml
# dev: deploy/docker-compose.yml
services:
  postgres:   # порт 5433:5432 (!) — не стандартный
  redis:      # 6379:6379
  minio:      # 9000 (API), 9001 (console)
  livekit:    # 7880 (HTTP/WS), 7881 (TCP RTC), 7882/udp
```

---

## Зависимости

### Backend
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

### Frontend
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

## Известные технические долги (приоритет исправления)

| Проблема | Файл | Приоритет |
|---|---|---|
| Rate limiter — memory leak (buckets не чистятся) | `middleware/rate_limit.go` | Высокий |
| `fileService` в `AuthHandler` — нарушение SRP | `handler/auth.go` | Средний |
| Cookie-getters в `AuthService` — утечка в транспорт | `service/auth.go` | Средний |
| `extractApiError` дублируется | `api/users.ts` | Средний |
| `UserUpdate` snake_case в TypeScript | `types/index.ts` | Низкий |
| `cmd/debugrefresh/` — удалить | `cmd/debugrefresh/` | Низкий |
| Swagger-аннотации удалены | `handler/auth.go` | Низкий |

---

## Горячие клавиши

| Клавиша | Действие |
|---|---|
| `Ctrl+K` | Открыть поиск |
| `Ctrl+/` | Показать горячие клавиши |
| `Alt+↑/↓` | Переключение каналов |
| `Shift+Enter` | Новая строка |
| `Esc` | Закрыть модальное окно |

---

## PowerShell — частые проблемы

```powershell
# pipe вместо <:
Get-Content file.sql | docker exec -i atlas-postgres psql -U atlas -d atlas

# Убить процесс на порту:
(Get-NetTCPConnection -LocalPort 8080 -EA SilentlyContinue).OwningProcess | Where-Object { $_ -ne 0 -and $_ -ne 4 } | Select-Object -Unique | ForEach-Object { Stop-Process -Id $_ -Force }
```
