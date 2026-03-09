# Сравнение архитектур WebSocket: Discord vs Slack

## 🎮 Discord подход

### Архитектура
```
Пользователь подключается → Подписка на весь Workspace (Guild)
                          ↓
            Получает ВСЕ события workspace:
            - Новые сообщения в любых каналах
            - Создание/удаление каналов
            - Изменение участников
            - Обновление ролей
            - Typing indicators
            - Presence (статусы онлайн)
```

### Преимущества ✅
1. **Простота подписки** - одна подписка на весь workspace
2. **Мгновенные обновления везде** - видишь activity во всех каналах
3. **Реактивность UI** - списки каналов обновляются автоматически
4. **Notifications/badges** - легко показывать непрочитанные везде
5. **Real-time presence** - кто где печатает, кто онлайн
6. **Меньше запросов** - не нужно переподписываться при смене канала

### Недостатки ❌
1. **Больше трафика** - получаешь события из всех каналов, даже неактивных
2. **Сложнее масштабировать** - большие workspace = много событий
3. **Privacy concerns** - пользователь может "подслушивать" приватные каналы (нужна фильтрация)
4. **Нагрузка на клиент** - нужно фильтровать и обрабатывать все события

### Лучше для:
- Небольших/средних команд (до 1000 человек)
- Активных workspace где важна реактивность
- Приложений где важен awareness (кто что делает)
- Если каналов не очень много (до 100-200)

---

## 💼 Slack подход

### Архитектура
```
Пользователь подключается → Подписка на конкретные ресурсы
                          ↓
            Подписывается выборочно:
            - Текущий открытый канал
            - DM conversations
            - Mentions (@user)
            - Workspace-level события (новые каналы, участники)
```

### Преимущества ✅
1. **Эффективность трафика** - получаешь только нужное
2. **Лучше масштабируется** - большие workspace с сотнями каналов
3. **Privacy by default** - видишь только то, на что подписан
4. **Меньше нагрузка на клиент** - меньше событий для обработки
5. **Гибкость** - можно подписаться на конкретные thread'ы

### Недостатки ❌
1. **Сложность подписок** - нужно управлять множеством подписок
2. **Задержки при переключении** - подписка/отписка при смене канала
3. **Больше запросов** - subscribe/unsubscribe при навигации
4. **Сложнее показывать activity** - нужны отдельные API для непрочитанных
5. **Менее реактивный UI** - новые каналы не появляются автоматически

### Лучше для:
- Больших организаций (1000+ человек)
- Enterprise с сотнями каналов
- Когда privacy критична
- Когда оптимизация трафика важна

---

## 🎯 Рекомендация для Atlas

### Гибридный подход (Best of both worlds)

```
Уровень 1: Workspace-level подписка (обязательная)
├── Создание/удаление каналов
├── Добавление/удаление участников workspace
├── Обновления workspace (имя, настройки)
└── Упоминания (@user, @here)

Уровень 2: Channel-level подписка (опциональная, автоматическая)
├── Новые сообщения в ОТКРЫТЫХ каналах пользователя
├── Typing indicators только в активном канале
└── Thread updates в активных thread'ах

Уровень 3: Direct подписки (по требованию)
├── Конкретный thread
└── Конкретная DM conversation
```

### Реализация

#### Backend структура
```go
type Hub struct {
    // Workspace-level subscriptions
    workspaces map[string]map[string]*Client  // workspaceID -> clientID -> Client
    
    // Channel-level subscriptions (опционально)
    channels map[string]map[string]*Client    // channelID -> clientID -> Client
    
    // User-level subscriptions
    users map[string][]*Client                // userID -> Clients
}

// События
type EventType string
const (
    // Workspace level
    EventWorkspaceUpdated    EventType = "workspace_updated"
    EventChannelCreated      EventType = "channel_created"
    EventChannelDeleted      EventType = "channel_deleted"
    EventMemberAdded         EventType = "member_added"
    EventMemberRemoved       EventType = "member_removed"
    
    // Channel level
    EventMessage             EventType = "message"
    EventMessageUpdated      EventType = "message_updated"
    EventMessageDeleted      EventType = "message_deleted"
    EventTyping              EventType = "typing"
    
    // User level
    EventMention             EventType = "mention"
    EventPresence            EventType = "presence"
)
```

#### Frontend подписка
```typescript
// При входе в workspace
wsStore.subscribeToWorkspace(workspaceId)

// При открытии канала (автоматически)
wsStore.subscribeToChannel(channelId)  // Добавляется к workspace подписке

// При выходе из канала
wsStore.unsubscribeFromChannel(channelId)
```

---

## 📊 Таблица сравнения для Atlas

| Критерий | Discord | Slack | Гибрид |
|----------|---------|-------|--------|
| **Простота реализации** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Масштабируемость** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Трафик** | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Реактивность UI** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Privacy** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **User Experience** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

---

## 💡 Моя рекомендация: **Гибридный подход**

### Почему?

1. **Начальная фаза (MVP)** - у вас небольшие команды
2. **Простота** - проще чем полный Slack подход
3. **Гибкость** - можно легко расширить позже
4. **UX** - пользователи видят активность workspace
5. **Масштабируемость** - можно добавить фильтры позже

### Этапы внедрения

#### Этап 1 (MVP) - Discord-like
- Подписка на workspace
- События: messages, channel_created, channel_deleted
- Простая фильтрация на клиенте по правам доступа

#### Этап 2 - Оптимизация
- Добавить channel-level подписки для typing
- Добавить фильтрацию на сервере
- Ввести rate limiting

#### Этап 3 - Enterprise features
- Selective subscriptions по требованию
- Сжатие событий
- Batch updates

---

## 🚀 Что делать прямо сейчас?

### Для Atlas рекомендую начать с Discord подхода, потому что:

1. ✅ **Быстрее внедрить** - меньше кода
2. ✅ **Лучше UX** - пользователи видят что происходит
3. ✅ **Ваш масштаб** - корпоративные команды обычно до 100-500 человек
4. ✅ **Проще дебажить** - видно все события
5. ✅ **Можно эволюционировать** - добавить Slack features позже

### Критерий выбора для будущего:

- **Workspace < 100 каналов** → Discord подход ✅
- **Workspace > 100 каналов** → Гибридный подход
- **Workspace > 500 каналов** → Slack подход

---

## 🎯 Итоговая рекомендация

**Начните с Discord подхода (workspace-level subscriptions)**

Это даст вам:
- Быструю реализацию (1-2 дня)
- Отличный UX
- Простоту отладки
- Возможность эволюции

Позже, если появятся workspace с сотнями каналов, можете добавить:
- Фильтрацию на сервере по правам
- Опциональные channel subscriptions
- Rate limiting

**Начнём с Discord подхода?** 🚀
