---
name: Дорожная карта Atlas
overview: "Комплексная стратегия развития корпоративного мессенджера Atlas на 12 месяцев с декомпозицией по этапам: стабилизация MVP, расширение функционала, масштабирование и вывод на рынок"
todos:
  - id: typing-indicators
    content: Реализовать typing indicators на backend и frontend (заменить TODO в websocket.ts)
    status: completed
  - id: presence-system
    content: Добавить систему presence (online/offline/away) с миграцией БД
    status: completed
  - id: member-events
    content: Завершить обработку member_added/member_removed событий
    status: completed
  - id: file-upload
    content: Интегрировать MinIO для загрузки файлов с drag-and-drop UI
    status: completed
  - id: search-implementation
    content: Реализовать полнотекстовый поиск по сообщениям (PostgreSQL FTS)
    status: completed
  - id: testing-suite
    content: Написать unit и E2E тесты для критических сценариев
    status: completed
  - id: ui-improvements
    content: "Улучшить UI/UX: темы оформления, accessibility, горячие клавиши"
    status: completed
  - id: livekit-integration
    content: Интегрировать LiveKit на frontend для видеозвонков
    status: completed
  - id: reactions
    content: Добавить реакции на сообщения (эмодзи)
    status: completed
  - id: markdown-editor
    content: Внедрить Tiptap для форматирования сообщений
    status: completed
  - id: notifications
    content: Реализовать push и email уведомления
    status: completed
  - id: message-to-task
    content: "Создать киллер-фичу: превращение сообщения в задачу с Kanban доской"
    status: completed
  - id: mobile-app
    content: Разработать Flutter приложение с core функционалом
    status: cancelled
  - id: mobile-push
    content: Настроить FCM для push-уведомлений на мобильных
    status: cancelled
  - id: sso-ldap
    content: Реализовать SSO (SAML/OAuth) и LDAP интеграцию для Enterprise
    status: cancelled
  - id: audit-log
    content: Добавить журнал аудита для Enterprise клиентов
    status: cancelled
  - id: admin-panel
    content: Создать отдельную админ-панель с метриками и управлением
    status: cancelled
  - id: scaling-redis
    content: Настроить Redis Pub/Sub для горизонтального масштабирования WebSocket
    status: cancelled
  - id: monitoring
    content: Внедрить Prometheus + Grafana для мониторинга
    status: cancelled
  - id: ai-summary
    content: Добавить AI саммари встреч через OpenAI API (бета)
    status: cancelled
isProject: false
---

# 🗺 Дорожная карта развития проекта Atlas

## Текущее состояние проекта

**Реализовано:**

- ✅ Базовая архитектура (Go backend + Vue frontend)
- ✅ Аутентификация (JWT)
- ✅ Воркспейсы и каналы
- ✅ Real-time сообщения через WebSocket
- ✅ Треды (вложенные обсуждения)
- ✅ WebSocket на уровне workspace (Discord-подход)
- ✅ Базовая инфраструктура (PostgreSQL, Redis, MinIO, LiveKit)

**Неполные элементы:**

- ⚠️ LiveKit интегрирован в инфраструктуре, но нет клиентской реализации
- ⚠️ Нет мобильных приложений (только заглушка)
- ⚠️ Typing indicators и presence не реализованы (только TODO)
- ⚠️ Member events не обрабатываются полностью (TODO в коде)
- ⚠️ Нет загрузки файлов (MinIO настроен, но функционал отсутствует)
- ⚠️ Отсутствуют Enterprise-функции (аудит, LDAP, SSO)
- ⚠️ Нет поиска по сообщениям
- ⚠️ Отсутствует админ-панель

---

## 🎯 Стратегические направления

### 1. Целевая аудитория: Все сегменты одновременно

**Сегментация:**

- **SMB (Малый/средний бизнес)** - через Freemium модель
- **Enterprise** - через лицензии с кастомизацией и SLA
- **SaaS** - массовый рынок с подпиской
- **On-premise** - гибридная модель для корпораций

### 2. Модели монетизации

#### Freemium (для привлечения пользователей)

- Бесплатно до 50 пользователей
- Ограничения: 10 GB хранилища, история сообщений 90 дней, базовые звонки до 5 участников

#### SaaS подписка

- **Basic:** $5/пользователь/месяц - до 100 пользователей, 50 GB, неограниченная история
- **Pro:** $10/пользователь/месяц - до 500 пользователей, 200 GB, расширенная статистика, гостевой доступ
- **Business:** $20/пользователь/месяц - неограниченно, 1 TB, приоритетная поддержка, SLA 99.9%

#### On-premise лицензии

- **Small:** $5000/год (до 100 пользователей)
- **Medium:** $15000/год (до 500 пользователей)
- **Enterprise:** $50000+/год (неограниченно) + кастомизация

#### Индивидуальные контракты

- Разработка интеграций под заказ
- Выделенный техподдержка
- Обучение сотрудников заказчика

---

## 📅 ЭТАП 1: Стабилизация MVP (Месяцы 1-3)

### Цель этапа

Довести существующий функционал до production-ready состояния, устранить все TODO и критические недоработки.

### 1.1 Завершение базового функционала

#### Typing Indicators (Индикаторы набора)

**Файлы:**

- `[backend/internal/transport/ws/client.go](backend/internal/transport/ws/client.go)` - добавить обработку события `typing`
- `[frontend/src/stores/websocket.ts](frontend/src/stores/websocket.ts)` - реализовать логику из TODO (строки 67, 72)
- `[frontend/src/stores/channels.ts](frontend/src/stores/channels.ts)` - хранение состояния typing по каналам

**Реализация:**

```typescript
// Frontend: отправка typing события при вводе
const typingTimeout = ref<NodeJS.Timeout>()
function handleTyping() {
  wsStore.send('typing', { channel_id: currentChannelId.value, typing: true })
  clearTimeout(typingTimeout.value)
  typingTimeout.value = setTimeout(() => {
    wsStore.send('typing', { channel_id: currentChannelId.value, typing: false })
  }, 3000)
}
```

**Backend:**

```go
// ws/client.go
case "typing":
    var data struct {
        ChannelID string `json:"channel_id"`
        Typing    bool   `json:"typing"`
    }
    json.Unmarshal(msg.Data, &data)
    
    // Broadcast только в channel
    c.hub.Broadcast(data.ChannelID, "typing", map[string]interface{}{
        "channel_id": data.ChannelID,
        "user_id": c.UserID,
        "user": c.User, // Нужно добавить User в Client
        "typing": data.Typing,
    }, c.UserID)
```

#### Presence (Статусы online/offline/away)

**Файлы:**

- `[backend/internal/domain/user.go](backend/internal/domain/user.go)` - добавить поле `Status` и `LastSeen`
- Новая миграция `004_user_presence.sql`
- `[backend/internal/transport/ws/hub.go](backend/internal/transport/ws/hub.go)` - broadcast presence при подключении/отключении
- `[frontend/src/stores/workspace.ts](frontend/src/stores/workspace.ts)` - хранение presence участников

**Схема БД:**

```sql
ALTER TABLE users 
ADD COLUMN status VARCHAR(20) DEFAULT 'offline',
ADD COLUMN last_seen TIMESTAMP WITH TIME ZONE;

CREATE INDEX idx_users_status ON users(status);
```

#### Member Events (Полная обработка)

**Файлы:**

- `[frontend/src/stores/websocket.ts](frontend/src/stores/websocket.ts)` - заменить TODO (строки 131, 137) на реальную логику
- `[frontend/src/stores/workspace.ts](frontend/src/stores/workspace.ts)` - методы `addMember`, `removeMember`, `updateMemberRole`

**Компоненты UI:**

- Новый компонент `WorkspaceMembersList.vue` - список участников с ролями
- Обновить `[frontend/src/components/workspace/InviteMemberModal.vue](frontend/src/components/workspace/InviteMemberModal.vue)` - показывать результат приглашения

### 1.2 Загрузка и хранение файлов

#### Backend: Интеграция MinIO

**Новые файлы:**

- `backend/internal/service/file.go` - сервис для работы с файлами
- `backend/internal/transport/http/handler/file.go` - API для загрузки/скачивания
- `backend/pkg/storage/minio.go` - клиент MinIO

**API endpoints:**

```
POST   /api/v1/files/upload           - Загрузка файла
GET    /api/v1/files/:id/download     - Скачивание файла
DELETE /api/v1/files/:id              - Удаление файла
GET    /api/v1/files/:id/thumbnail    - Превью для изображений
```

**Обновление БД:**

```sql
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100),
    size_bytes BIGINT,
    storage_path TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

ALTER TABLE messages 
ADD COLUMN has_attachments BOOLEAN DEFAULT FALSE;
```

#### Frontend: Upload UI

**Новые компоненты:**

- `frontend/src/components/chat/FileUpload.vue` - drag-and-drop область
- `frontend/src/components/chat/FilePreview.vue` - превью файлов перед отправкой
- `frontend/src/api/files.ts` - API клиент для файлов

**Интеграция:**

- Обновить `[frontend/src/components/chat/MessageInput.vue](frontend/src/components/chat/MessageInput.vue)` - кнопка прикрепления файлов
- Обновить `[frontend/src/components/chat/MessageItem.vue](frontend/src/components/chat/MessageItem.vue)` - отображение файлов

**Ограничения Freemium:**

```typescript
const FILE_LIMITS = {
  free: {
    maxFileSize: 10 * 1024 * 1024, // 10 MB
    allowedTypes: ['image/*', 'application/pdf', 'text/*']
  },
  pro: {
    maxFileSize: 100 * 1024 * 1024, // 100 MB
    allowedTypes: ['*']
  }
}
```

### 1.3 Поиск по сообщениям

#### Backend: PostgreSQL Full-Text Search

**Новая миграция:**

```sql
-- Добавляем tsvector колонку для полнотекстового поиска
ALTER TABLE messages 
ADD COLUMN search_vector tsvector;

-- Создаем GIN индекс для быстрого поиска
CREATE INDEX idx_messages_search ON messages USING GIN(search_vector);

-- Триггер для автоматического обновления search_vector
CREATE OR REPLACE FUNCTION messages_search_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_vector := to_tsvector('russian', COALESCE(NEW.content, ''));
  RETURN NEW;
END
$$ LANGUAGE plpgsql;

CREATE TRIGGER messages_search_update 
BEFORE INSERT OR UPDATE ON messages
FOR EACH ROW EXECUTE FUNCTION messages_search_trigger();
```

**Новые файлы:**

- `backend/internal/repository/postgres/search.go` - репозиторий для поиска
- `backend/internal/service/search.go` - логика поиска с фильтрацией
- `backend/internal/transport/http/handler/search.go` - API эндпоинты

**API:**

```go
GET /api/v1/search?q=текст&workspace_id=uuid&channel_id=uuid&user_id=uuid&from=date&to=date
```

#### Frontend: Search UI

**Новые компоненты:**

- `frontend/src/components/search/SearchBar.vue` - панель поиска (Cmd+K / Ctrl+K)
- `frontend/src/components/search/SearchResults.vue` - результаты с группировкой
- `frontend/src/stores/search.ts` - Pinia store для поиска

### 1.4 Тестирование и багфиксинг

#### Unit-тесты (критичные области)

**Backend:**

- `backend/internal/transport/ws/hub_test.go` - тесты Hub подписок и broadcast
- `backend/internal/service/message_test.go` - создание, обновление, удаление
- `backend/internal/service/auth_test.go` - JWT генерация, валидация

**Frontend:**

- `frontend/src/stores/__tests__/websocket.spec.ts` - тесты WebSocket store
- `frontend/src/stores/__tests__/messages.spec.ts` - обработка сообщений

#### E2E тесты (критические сценарии)

**Инструмент:** Playwright

**Сценарии:**

1. Регистрация → создание workspace → создание канала → отправка сообщения
2. Два пользователя: один пишет → другой получает real-time
3. Создание треда → получение уведомлений
4. Загрузка файла → скачивание другим пользователем

#### Performance тесты

**Инструмент:** k6

**Сценарии:**

- 1000 одновременных WebSocket соединений
- 100 сообщений/секунду broadcast
- Нагрузка на поиск: 50 запросов/секунду

### 1.5 Улучшение UI/UX

#### Адаптивная верстка

- Мобильная версия веб-приложения (responsive)
- Hamburger меню для боковых панелей
- Оптимизация для планшетов

#### Темы оформления

**Файлы:**

- `frontend/src/assets/themes/dark.css`
- `frontend/src/assets/themes/light.css`
- `frontend/src/assets/themes/high-contrast.css` (для accessibility)

**Компонент:**

- `frontend/src/components/settings/ThemeSwitcher.vue`

#### Accessibility (A11y)

- ARIA-метки для всех интерактивных элементов
- Навигация с клавиатуры (Tab, Enter, Esc)
- Screen reader поддержка
- Высокая контрастность для слабовидящих

#### Горячие клавиши

```typescript
const SHORTCUTS = {
  'Ctrl+K': 'Открыть поиск',
  'Ctrl+/': 'Список горячих клавиш',
  'Esc': 'Закрыть модалку/поиск',
  'Ctrl+Enter': 'Отправить сообщение',
  'Ctrl+Shift+M': 'Переключить mute',
  'Alt+↑/↓': 'Переключение между каналами'
}
```

---

## 📅 ЭТАП 2: Расширение функционала (Месяцы 4-6)

### Цель этапа

Добавить ключевые функции, которые выделят Atlas среди конкурентов и откроют возможность монетизации.

### 2.1 Видеозвонки и аудио-комнаты (LiveKit)

#### Backend: LiveKit интеграция

**Новые файлы:**

- `backend/internal/service/livekit.go` - генерация токенов для LiveKit
- `backend/internal/transport/http/handler/calls.go` - API для звонков

**Зависимости:**

```bash
go get github.com/livekit/protocol/livekit
go get github.com/livekit/server-sdk-go
```

**API endpoints:**

```go
POST   /api/v1/calls/create           - Создать комнату
POST   /api/v1/calls/:room_id/join    - Получить токен для присоединения
DELETE /api/v1/calls/:room_id         - Завершить звонок
GET    /api/v1/calls/active           - Список активных звонков
```

**Обновление БД:**

```sql
CREATE TABLE calls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID REFERENCES channels(id),
    livekit_room_id VARCHAR(255) UNIQUE NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    started_by UUID REFERENCES users(id)
);

CREATE TABLE call_participants (
    call_id UUID REFERENCES calls(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (call_id, user_id, joined_at)
);
```

#### Frontend: LiveKit клиент

**Зависимости:**

```json
{
  "dependencies": {
    "livekit-client": "^2.0.0",
    "@livekit/components-vue": "^2.0.0"
  }
}
```

**Новые компоненты:**

- `frontend/src/components/calls/CallModal.vue` - окно звонка
- `frontend/src/components/calls/VideoGrid.vue` - сетка участников
- `frontend/src/components/calls/AudioControls.vue` - управление микрофоном/камерой
- `frontend/src/components/calls/ScreenShare.vue` - демонстрация экрана
- `frontend/src/stores/calls.ts` - управление звонками

**Интеграция:**

- Кнопка "Начать звонок" в `[frontend/src/components/chat/ChatHeader.vue](frontend/src/components/chat/ChatHeader.vue)`
- Индикатор активного звонка в сайдбаре

#### Аудио-комнаты (Huddles)

**Особенности:**

- Постоянные голосовые каналы (как Discord voice channels)
- Входим/выходим без подтверждения
- "Push to Talk" режим (опционально)

**Новый тип канала:**

```sql
-- В migrations
ALTER TABLE channels 
ADD COLUMN voice_settings JSONB DEFAULT '{
  "push_to_talk": false,
  "auto_mute": true,
  "quality": "high"
}'::jsonb;
```

### 2.2 Реакции и эмодзи

#### Backend

**Обновление БД:**

```sql
CREATE TABLE message_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    emoji VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(message_id, user_id, emoji)
);

CREATE INDEX idx_message_reactions ON message_reactions(message_id);
```

**Новые файлы:**

- `backend/internal/repository/postgres/reaction.go`
- `backend/internal/service/reaction.go`
- `backend/internal/transport/http/handler/reaction.go`

**API:**

```go
POST   /api/v1/messages/:id/reactions   - Добавить реакцию
DELETE /api/v1/messages/:id/reactions/:emoji - Удалить реакцию
GET    /api/v1/messages/:id/reactions   - Получить все реакции
```

**WebSocket события:**

```go
// При добавлении реакции
h.wsHub.BroadcastToWorkspace(workspaceID, "reaction_added", map[string]interface{}{
    "message_id": messageID,
    "user_id": userID,
    "emoji": emoji,
})
```

#### Frontend

**Новые компоненты:**

- `frontend/src/components/chat/EmojiPicker.vue` - выбор эмодзи
- `frontend/src/components/chat/ReactionBar.vue` - отображение реакций под сообщением

**Библиотека:**

```bash
npm install emoji-picker-vue
```

### 2.3 Markdown и форматирование

#### Редактор сообщений

**Библиотека:** Tiptap (расширяемый редактор)

```bash
npm install @tiptap/vue-3 @tiptap/starter-kit @tiptap/extension-link @tiptap/extension-code-block-lowlight
```

**Поддерживаемые элементы:**

- **Жирный**, *курсив*, ~~зачеркнутый~~
- Списки (маркированные, нумерованные)
- Цитаты
- Ссылки (автоматическое определение)
- Код (inline и блоки с подсветкой синтаксиса)
- Упоминания @user

**Компонент:**

- Обновить `[frontend/src/components/chat/MessageInput.vue](frontend/src/components/chat/MessageInput.vue)`

**Backend обработка:**

- Валидация и санитизация HTML (библиотека `bluemonday`)
- Парсинг @mentions для отправки уведомлений

### 2.4 Уведомления (Notifications)

#### Push-уведомления (Web Push)

**Backend:**

- Интеграция с Web Push API
- Хранение subscription в БД

```sql
CREATE TABLE push_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    keys JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Frontend:**

- Запрос разрешения на уведомления
- Service Worker для фоновых уведомлений

#### Email-уведомления

**Backend:**

- Интеграция SMTP (или SendGrid/Mailgun)
- Настройки пользователя: какие события присылать на email

**Типы уведомлений:**

- Упоминание (@mention)
- Ответ в треде
- Приглашение в workspace/канал
- Пропущенный звонок

**Настройки:**

```sql
CREATE TABLE notification_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    mentions BOOLEAN DEFAULT TRUE,
    threads BOOLEAN DEFAULT TRUE,
    invites BOOLEAN DEFAULT TRUE,
    calls BOOLEAN DEFAULT TRUE,
    email_enabled BOOLEAN DEFAULT FALSE,
    email_frequency VARCHAR(20) DEFAULT 'instant' -- instant, hourly, daily
);
```

### 2.5 Message-to-Task (Киллер-фича)

#### Концепция

Превращение любого сообщения в задачу (тикет) с назначением ответственного и дедлайном.

#### Backend

**Новая таблица:**

```sql
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID REFERENCES messages(id),
    workspace_id UUID REFERENCES workspaces(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) DEFAULT 'todo', -- todo, in_progress, done, cancelled
    priority VARCHAR(20) DEFAULT 'medium', -- low, medium, high, urgent
    assignee_id UUID REFERENCES users(id),
    reporter_id UUID REFERENCES users(id),
    due_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tasks_assignee ON tasks(assignee_id);
CREATE INDEX idx_tasks_workspace ON tasks(workspace_id);
CREATE INDEX idx_tasks_status ON tasks(status);
```

**Новые файлы:**

- `backend/internal/domain/task.go`
- `backend/internal/repository/postgres/task.go`
- `backend/internal/service/task.go`
- `backend/internal/transport/http/handler/task.go`

**API:**

```go
POST   /api/v1/tasks                  - Создать задачу из сообщения
GET    /api/v1/tasks                  - Список задач (с фильтрами)
PATCH  /api/v1/tasks/:id/status       - Изменить статус
PATCH  /api/v1/tasks/:id/assign       - Назначить ответственного
DELETE /api/v1/tasks/:id              - Удалить задачу
```

#### Frontend

**Новые компоненты:**

- `frontend/src/components/tasks/TaskCreationModal.vue` - создание задачи из сообщения
- `frontend/src/components/tasks/TaskBoard.vue` - Kanban-доска задач
- `frontend/src/components/tasks/TaskCard.vue` - карточка задачи
- `frontend/src/views/TasksView.vue` - отдельная страница с задачами

**Интеграция:**

- Контекстное меню в `[frontend/src/components/chat/MessageItem.vue](frontend/src/components/chat/MessageItem.vue)` - кнопка "Создать задачу"
- Бейдж на сообщении, если из него создана задача

---

## 📅 ЭТАП 3: Мобильные приложения (Месяцы 7-9)

### Цель этапа

Разработать нативные мобильные приложения для iOS и Android, обеспечить паритет функций с веб-версией.

### 3.1 Настройка Flutter проекта

#### Инициализация

```bash
cd mobile
flutter create atlas_mobile
cd atlas_mobile
```

#### Структура проекта

```
mobile/atlas_mobile/
├── lib/
│   ├── core/
│   │   ├── api/              # HTTP и WebSocket клиенты
│   │   ├── models/           # Dart модели (User, Message, Channel)
│   │   └── utils/            # Helpers
│   ├── features/
│   │   ├── auth/
│   │   │   ├── screens/      # LoginScreen, RegisterScreen
│   │   │   ├── widgets/      # Компоненты
│   │   │   └── providers/    # State management (Riverpod)
│   │   ├── chat/
│   │   ├── calls/
│   │   └── workspace/
│   ├── main.dart
│   └── routes.dart
├── android/
├── ios/
└── pubspec.yaml
```

#### Зависимости

```yaml
dependencies:
  flutter:
    sdk: flutter
  
  # State management
  flutter_riverpod: ^2.4.0
  
  # Networking
  dio: ^5.3.2
  web_socket_channel: ^2.4.0
  
  # Local storage
  hive: ^2.2.3
  hive_flutter: ^1.1.0
  shared_preferences: ^2.2.1
  
  # LiveKit
  livekit_client: ^2.0.0
  
  # UI
  flutter_svg: ^2.0.7
  cached_network_image: ^3.3.0
  flutter_markdown: ^0.6.17
  emoji_picker_flutter: ^1.6.1
  
  # Notifications
  firebase_messaging: ^14.6.8
  flutter_local_notifications: ^15.1.1
  
  # File handling
  file_picker: ^5.5.0
  image_picker: ^1.0.4
```

### 3.2 Core функционал

#### Аутентификация

**Файлы:**

- `lib/core/api/auth_api.dart` - HTTP клиент для auth
- `lib/features/auth/providers/auth_provider.dart` - состояние авторизации
- `lib/features/auth/screens/login_screen.dart`
- `lib/features/auth/screens/register_screen.dart`

**Хранение токена:**

```dart
// Использование secure storage для токена
final storage = FlutterSecureStorage();
await storage.write(key: 'atlas_token', value: token);
```

#### WebSocket соединение

**Файл:** `lib/core/api/websocket_service.dart`

```dart
class WebSocketService {
  late WebSocketChannel channel;
  final StreamController<WSEvent> eventStream;
  
  void connect(String token) {
    final uri = Uri.parse('ws://api.example.com/ws?token=$token');
    channel = WebSocketChannel.connect(uri);
    
    channel.stream.listen((message) {
      final event = WSEvent.fromJson(jsonDecode(message));
      eventStream.add(event);
    });
  }
  
  void send(String event, Map<String, dynamic> data) {
    channel.sink.add(jsonEncode({'event': event, 'data': data}));
  }
}
```

#### Чаты и сообщения

**Экраны:**

- `lib/features/chat/screens/channel_list_screen.dart` - список каналов
- `lib/features/chat/screens/chat_screen.dart` - окно чата
- `lib/features/chat/screens/thread_screen.dart` - тред

**Widgets:**

- `lib/features/chat/widgets/message_item.dart` - элемент сообщения
- `lib/features/chat/widgets/message_input.dart` - поле ввода
- `lib/features/chat/widgets/typing_indicator.dart`

### 3.3 Мобильные специфичные функции

#### Push-уведомления (Firebase Cloud Messaging)

**Настройка:**

1. Создать Firebase проект
2. Добавить `google-services.json` (Android) и `GoogleService-Info.plist` (iOS)
3. Настроить backend для отправки FCM

**Файл:** `lib/core/services/notification_service.dart`

```dart
class NotificationService {
  final FirebaseMessaging messaging = FirebaseMessaging.instance;
  
  Future<void> initialize() async {
    // Запросить разрешение
    NotificationSettings settings = await messaging.requestPermission();
    
    // Получить FCM токен
    String? token = await messaging.getToken();
    // Отправить token на backend
    
    // Обработка уведомлений
    FirebaseMessaging.onMessage.listen((RemoteMessage message) {
      _showLocalNotification(message);
    });
  }
}
```

#### Загрузка файлов с камеры

**Файл:** `lib/features/chat/widgets/attachment_picker.dart`

```dart
Future<void> pickImage(ImageSource source) async {
  final picker = ImagePicker();
  final XFile? image = await picker.pickImage(source: source);
  
  if (image != null) {
    // Сжать изображение
    final compressed = await compressImage(File(image.path));
    // Загрузить на сервер
    await uploadFile(compressed);
  }
}
```

#### Оффлайн режим

**Библиотека:** Hive (локальная БД)

```dart
// Кэширование сообщений
final box = await Hive.openBox<Message>('messages');
box.put(message.id, message);

// При отсутствии интернета - читаем из кэша
final cachedMessages = box.values.toList();
```

#### Голосовые сообщения

**Библиотека:** `record` для записи, `just_audio` для воспроизведения

```dart
Future<void> recordVoiceMessage() async {
  final recorder = Record();
  await recorder.start();
  // ... запись
  await recorder.stop();
  // Загрузить как файл
}
```

### 3.4 LiveKit на мобильных

**Файл:** `lib/features/calls/screens/call_screen.dart`

```dart
class CallScreen extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final room = ref.watch(callProvider).room;
    
    return Scaffold(
      body: room != null
        ? Stack(
            children: [
              ParticipantGrid(room: room),
              CallControls(room: room),
            ],
          )
        : CircularProgressIndicator(),
    );
  }
}
```

### 3.5 Публикация в сторы

#### App Store (iOS)

1. Регистрация в Apple Developer Program ($99/год)
2. Создание App ID, provisioning profiles
3. Подготовка иконок, скриншотов
4. Тестирование через TestFlight
5. Отправка на ревью

#### Google Play (Android)

1. Регистрация в Google Play Console ($25 одноразово)
2. Подготовка иконок, скриншотов, описания
3. Создание signed APK/AAB
4. Internal/Closed testing
5. Публикация в Production

---

## 📅 ЭТАП 4: Enterprise функции и масштабирование (Месяцы 10-12)

### Цель этапа

Подготовить продукт для корпоративных клиентов, обеспечить масштабируемость и безопасность на уровне Enterprise.

### 4.1 Безопасность и комплаенс

#### SSO (Single Sign-On)

**Протоколы:** SAML 2.0, OAuth 2.0, OpenID Connect

**Backend:**

```bash
go get github.com/crewjam/saml
go get golang.org/x/oauth2
```

**Новые файлы:**

- `backend/internal/service/sso.go`
- `backend/internal/transport/http/handler/sso.go`

**Конфигурация:**

```sql
CREATE TABLE sso_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id),
    provider_type VARCHAR(20) NOT NULL, -- saml, oauth, oidc
    issuer TEXT NOT NULL,
    metadata_url TEXT,
    client_id TEXT,
    client_secret TEXT,
    enabled BOOLEAN DEFAULT TRUE
);
```

#### LDAP/Active Directory интеграция

**Библиотека:** `github.com/go-ldap/ldap`

**Функционал:**

- Синхронизация пользователей из LDAP
- Автоматическое создание workspace членов
- Синхронизация групп → каналов

**Файлы:**

- `backend/internal/service/ldap.go`
- `backend/cmd/ldap-sync/main.go` - отдельная утилита для синхронизации

#### Audit Log (Журнал аудита)

**Таблица:**

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id),
    user_id UUID REFERENCES users(id),
    action VARCHAR(100) NOT NULL, -- user_login, message_deleted, channel_created
    resource_type VARCHAR(50),     -- message, channel, user
    resource_id UUID,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_workspace ON audit_logs(workspace_id, created_at);
CREATE INDEX idx_audit_user ON audit_logs(user_id, created_at);
CREATE INDEX idx_audit_action ON audit_logs(action);
```

**Middleware:**

```go
func AuditMiddleware(auditSvc *service.AuditService) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next() // Выполняем запрос
        
        // Логируем после выполнения
        auditSvc.Log(context.Background(), &domain.AuditLog{
            UserID: getUserID(c),
            Action: c.Request.Method + " " + c.Request.URL.Path,
            IPAddress: c.ClientIP(),
            UserAgent: c.Request.UserAgent(),
        })
    }
}
```

#### Шифрование данных

**At rest:**

- Шифрование БД на уровне диска (LUKS, BitLocker)
- Шифрование файлов в MinIO (SSE-C)

**In transit:**

- Обязательный TLS 1.3
- Certificate pinning в мобильных приложениях

**End-to-End шифрование (опционально для Enterprise+):**

- Использование libsodium/NaCl
- Ключи хранятся только у клиентов
- Backend не имеет доступа к расшифрованным данным

### 4.2 Админ-панель

#### Backend API

**Новые эндпоинты:**

```go
// Управление пользователями
GET    /api/v1/admin/users
POST   /api/v1/admin/users/:id/suspend
DELETE /api/v1/admin/users/:id

// Управление workspace
GET    /api/v1/admin/workspaces
PATCH  /api/v1/admin/workspaces/:id/limits

// Статистика
GET    /api/v1/admin/stats/users
GET    /api/v1/admin/stats/messages
GET    /api/v1/admin/stats/storage

// Audit logs
GET    /api/v1/admin/audit
```

#### Frontend: Отдельное приложение

**Новая папка:** `admin-panel/`

```
admin-panel/
├── src/
│   ├── pages/
│   │   ├── Dashboard.vue       # Главная страница с метриками
│   │   ├── Users.vue           # Управление пользователями
│   │   ├── Workspaces.vue      # Управление workspace
│   │   ├── AuditLogs.vue       # Журнал аудита
│   │   └── Settings.vue        # Глобальные настройки
│   ├── components/
│   │   ├── charts/             # Графики (Chart.js)
│   │   └── tables/             # Таблицы с фильтрами
│   └── App.vue
├── package.json
└── vite.config.ts
```

**Библиотеки:**

```json
{
  "dependencies": {
    "vue": "^3.4.21",
    "vue-router": "^4.3.0",
    "chart.js": "^4.4.0",
    "vue-chartjs": "^5.2.0"
  }
}
```

### 4.3 Масштабирование архитектуры

#### Горизонтальное масштабирование backend

**Проблема:** Множество серверов → нужна синхронизация WebSocket

**Решение:** Redis Pub/Sub

```go
// backend/internal/transport/ws/redis_hub.go
type RedisHub struct {
    *Hub
    redisClient *redis.Client
}

func (h *RedisHub) BroadcastToWorkspace(workspaceID, event string, data interface{}) {
    // Публикуем в Redis
    message := BroadcastMessage{
        WorkspaceID: workspaceID,
        Event: event,
        Data: data,
    }
    
    payload, _ := json.Marshal(message)
    h.redisClient.Publish(ctx, "broadcast:"+workspaceID, payload)
}

func (h *RedisHub) SubscribeToRedis() {
    pubsub := h.redisClient.PSubscribe(ctx, "broadcast:*")
    
    for msg := range pubsub.Channel() {
        var broadcastMsg BroadcastMessage
        json.Unmarshal([]byte(msg.Payload), &broadcastMsg)
        
        // Отправляем только локальным клиентам
        h.Hub.handleBroadcast(&broadcastMsg)
    }
}
```

**Обновление docker-compose:**

```yaml
services:
  backend:
    deploy:
      replicas: 3  # Несколько инстансов
    environment:
                                                                                 - REDIS_PUBSUB_ENABLED=true
  
  nginx:
    # Балансировка нагрузки
    depends_on:
                                                                                 - backend
```

#### Database масштабирование

**Read replicas:**

```yaml
services:
  postgres-primary:
    image: postgres:15
    
  postgres-replica:
    image: postgres:15
    environment:
                                                                                 - POSTGRES_MASTER_HOST=postgres-primary
```

**Connection pooling:**

```go
// backend/pkg/database/postgres.go
pool, err := pgxpool.New(ctx, "postgres://...?pool_max_conns=20")
```

#### CDN для статики

**Интеграция CloudFlare/Cloudinary:**

- Раздача изображений через CDN
- Автоматическое сжатие и оптимизация
- WebP конвертация

### 4.4 Мониторинг и наблюдаемость

#### Метрики (Prometheus + Grafana)

**Backend инструментация:**

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    wsConnections = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "atlas_websocket_connections",
    })
    
    messagesTotal = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "atlas_messages_total",
    })
)

func init() {
    prometheus.MustRegister(wsConnections, messagesTotal)
}
```

**Docker-compose добавить:**

```yaml
services:
  prometheus:
    image: prom/prometheus
    volumes:
                                                                                 - ./deploy/configs/prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
                                                                                 - "9090:9090"
  
  grafana:
    image: grafana/grafana
    ports:
                                                                                 - "3001:3000"
    depends_on:
                                                                                 - prometheus
```

#### Логирование (ELK Stack)

```yaml
services:
  elasticsearch:
    image: elasticsearch:8.9.0
  
  logstash:
    image: logstash:8.9.0
  
  kibana:
    image: kibana:8.9.0
    ports:
                                                                                 - "5601:5601"
```

**Backend структурированное логирование:**

```go
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
logger.Info("Message sent",
    zap.String("message_id", msg.ID),
    zap.String("channel_id", msg.ChannelID),
    zap.String("user_id", msg.UserID),
)
```

#### Tracing (Jaeger)

**Библиотека:** OpenTelemetry

```go
import "go.opentelemetry.io/otel"

tracer := otel.Tracer("atlas-backend")
ctx, span := tracer.Start(ctx, "message.create")
defer span.End()
```

### 4.5 AI-функции (бета)

#### AI Summary встреч

**Интеграция:** OpenAI GPT-4 / Claude API

**Функционал:**

1. Запись звонка (LiveKit recording)
2. Транскрипция аудио → текст (Whisper API)
3. Генерация саммари через GPT-4
4. Отправка результата в канал

**Backend:**

```go
// backend/internal/service/ai_summary.go
func (s *AISummaryService) GenerateSummary(callID string) (*Summary, error) {
    // 1. Получить запись звонка
    recording := s.livekit.GetRecording(callID)
    
    // 2. Транскрибировать
    transcript := s.whisper.Transcribe(recording.AudioFile)
    
    // 3. Генерировать саммари
    prompt := fmt.Sprintf("Summarize this meeting transcript:\n\n%s", transcript)
    summary := s.openai.Complete(prompt)
    
    return &Summary{
        CallID: callID,
        Transcript: transcript,
        Summary: summary,
        KeyPoints: extractKeyPoints(summary),
        ActionItems: extractActionItems(summary),
    }, nil
}
```

**UI:**

- Автоматическое сообщение после звонка с саммари
- Кнопка "Сгенерировать саммари вручную"

#### Smart Search (семантический поиск)

**Технология:** Vector embeddings + PostgreSQL pgvector

**Обновление БД:**

```sql
CREATE EXTENSION vector;

ALTER TABLE messages 
ADD COLUMN embedding vector(1536); -- OpenAI ada-002 размер

CREATE INDEX idx_messages_embedding ON messages USING ivfflat (embedding vector_cosine_ops);
```

**Процесс:**

1. При создании сообщения → генерировать embedding
2. Поиск: запрос → embedding → cosine similarity

---

## 📅 Развитие и продвижение

### 5.1 MVP → Production (Месяцы 1-3)

#### Шаги запуска

1. **Бета-тестирование**
  - Закрытая бета для 5-10 компаний
                                                                  - Сбор фидбека через встроенную форму
                                                                  - Исправление критических багов
2. **Документация**
  - API документация (Swagger/OpenAPI)
                                                                  - Руководство администратора (on-premise установка)
                                                                  - Пользовательская документация (Getting Started)
3. **Landing Page**
  - Описание продукта
                                                                  - Демо-видео
                                                                  - Pricing таблица
                                                                  - Форма заявки на бета-доступ
4. **Юридическая подготовка**
  - Terms of Service
                                                                  - Privacy Policy
                                                                  - GDPR compliance (если планируется Европа)
                                                                  - Договор оферты

### 5.2 Go-to-Market стратегия

#### Канал 1: Freemium SaaS

**Платформа:** atlas.io (гипотетический домен)

**Воронка:**

1. Регистрация без кредитки (low friction)
2. Onboarding: создание workspace → приглашение коллег → первое сообщение
3. Триггеры апгрейда:
  - Достижение 50 пользователей
                                                                  - Заполнение 10 GB хранилища
                                                                  - Потребность в расширенных звонках

**Маркетинг:**

- Контент-маркетинг: статьи на VC.ru, Habr
- SEO: целевые запросы "корпоративный мессенджер", "аналог Slack"
- Реферальная программа: пригласи друга → получи месяц Pro бесплатно

#### Канал 2: On-premise для Enterprise

**Целевые клиенты:**

- Банки и финансовые организации
- Госструктуры
- Промышленные компании с закрытым контуром

**Продажи:**

- Прямые продажи (outbound)
- Участие в тендерах
- Партнерства с системными интеграторами

**Демонстрация:**

- Виртуальная демо-среда (sandbox)
- Пилотный проект на 1-3 месяца
- POC (Proof of Concept) с кастомной интеграцией

#### Канал 3: Маркетплейсы

- **G2, Capterra** - листинг с отзывами
- **AWS Marketplace** - продажа on-premise через AWS
- **Microsoft AppSource** - интеграция с Teams

### 5.3 Партнерства и интеграции

#### Интеграции (Месяцы 6-12)

**Приоритет 1 (Must Have):**

- GitHub/GitLab - уведомления о коммитах, PR
- Jira/Trello - синхронизация задач
- Google Drive/Dropbox - вставка файлов
- Календари (Google Calendar, Outlook) - планирование встреч

**Приоритет 2 (Should Have):**

- CI/CD (Jenkins, CircleCI) - статусы сборок
- Monitoring (Grafana, Datadog) - алерты
- CRM (Salesforce, Битрикс24)

**Технология:** Webhooks + OAuth 2.0

**UI:** Marketplace интеграций в админ-панели

#### OEM партнерства

- Продажа white-label версии корпорациям
- Интеграция в существующие корпоративные порталы

### 5.4 Community и Open Source

#### Open Source Core

**Модель:** Open Core (ядро открыто, Enterprise функции проприетарные)

**Репозиторий:** GitHub

- MIT License для core
- Proprietary license для Enterprise модуля

**Преимущества:**

- Доверие сообщества
- Контрибуции от разработчиков
- Бесплатный маркетинг

#### Community

- Discord-сервер для пользователей
- Форум (Discourse)
- Еженедельные демо/Q&A сессии
- Хакатоны для разработчиков интеграций

---

## 🎯 Метрики успеха

### Технические метрики

- Uptime: 99.9% (SaaS)
- Latency: < 100ms для сообщений (p95)
- WebSocket connections: 10,000+ одновременно
- Database queries: < 50ms (p95)

### Бизнес-метрики

**Месяц 3:**

- 100+ активных workspace (бета)
- 1,000+ зарегистрированных пользователей

**Месяц 6:**

- 1,000+ активных workspace
- 10,000+ пользователей
- 5% конверсия free → paid

**Месяц 12:**

- 10,000+ активных workspace
- 100,000+ пользователей
- 50+ Enterprise клиентов
- MRR: $50,000+

---

## 🚧 Риски и митигация

### Технические риски

| Риск | Вероятность | Влияние | Митигация |

|------|------------|---------|-----------|

| Проблемы масштабирования WebSocket | Средняя | Высокое | Ранние нагрузочные тесты, Redis Pub/Sub |

| Сложность LiveKit интеграции | Средняя | Среднее | Тщательное изучение документации, POC |

| Потеря данных | Низкая | Критическое | Регулярные бэкапы, репликация БД |

### Бизнес риски

| Риск | Вероятность | Влияние | Митигация |

|------|------------|---------|-----------|

| Низкая конверсия free → paid | Высокая | Высокое | A/B тесты, улучшение onboarding |

| Сильные конкуренты (Slack, Teams) | Высокая | Среднее | Фокус на уникальные фичи (Message-to-Task, on-premise) |

| Медленное принятие рынком | Средняя | Высокое | Агрессивный маркетинг, реферальная программа |

---

## 📚 Заключение

Эта дорожная карта охватывает развитие Atlas на 12 месяцев:

- **Месяцы 1-3:** Стабилизация MVP, устранение технического долга
- **Месяцы 4-6:** Ключевые функции (видеозвонки, задачи, поиск)
- **Месяцы 7-9:** Мобильные приложения
- **Месяцы 10-12:** Enterprise функции и масштабирование

**Главные приоритеты для solo-разработчика:**

1. Завершить базовый функционал (typing, presence, файлы)
2. Запустить стабильную бета-версию
3. Привлечь первых пользователей через freemium
4. По мере роста - нанять команду для ускорения развития

**Конкурентные преимущества:**

- On-premiseFirst с простым развертыванием
- Message-to-Task киллер-фича
- Качественные звонки через LiveKit
- Открытый core для доверия сообщества

