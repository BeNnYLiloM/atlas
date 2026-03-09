# Команды для работы с БД Atlas

## Просмотр через Adminer (Web UI)

**URL:** http://localhost:8081

**Данные для входа:**
- Система: PostgreSQL
- Сервер: `atlas-postgres` (или `localhost:5433`)
- Пользователь: `atlas`
- Пароль: `atlas`
- База данных: `atlas`

## Команды через терминал

### Базовые команды

```powershell
# Список всех таблиц
docker exec atlas-postgres psql -U atlas -d atlas -c "\dt"

# Структура таблицы users
docker exec atlas-postgres psql -U atlas -d atlas -c "\d users"
```

### Просмотр данных

```powershell
# Все пользователи
docker exec atlas-postgres psql -U atlas -d atlas -c "SELECT * FROM users;"

# Пользователи с форматированием
docker exec atlas-postgres psql -U atlas -d atlas -c "SELECT id, email, display_name, created_at FROM users ORDER BY created_at DESC;"

# Все workspace
docker exec atlas-postgres psql -U atlas -d atlas -c "SELECT * FROM workspaces;"

# Все каналы
docker exec atlas-postgres psql -U atlas -d atlas -c "SELECT * FROM channels;"

# Последние 10 сообщений
docker exec atlas-postgres psql -U atlas -d atlas -c "SELECT m.*, u.display_name FROM messages m LEFT JOIN users u ON m.user_id = u.id ORDER BY m.created_at DESC LIMIT 10;"

# Подсчет записей
docker exec atlas-postgres psql -U atlas -d atlas -c "SELECT 'users' as table, COUNT(*) FROM users UNION SELECT 'workspaces', COUNT(*) FROM workspaces UNION SELECT 'channels', COUNT(*) FROM channels UNION SELECT 'messages', COUNT(*) FROM messages;"
```

### Интерактивная консоль

```powershell
# Войти в интерактивную psql консоль
docker exec -it atlas-postgres psql -U atlas -d atlas
```

В интерактивной консоли можно использовать:
- `\dt` - список таблиц
- `\d table_name` - структура таблицы
- `\q` - выход
- Любые SQL запросы

### Быстрый просмотр через скрипт

```powershell
# Запустить интерактивный скрипт просмотра БД
pwsh scripts/db-view.ps1
```

## Полезные SQL запросы

### Пользователи с их workspace

```sql
SELECT 
    u.email,
    u.display_name,
    w.name as workspace,
    wm.role
FROM users u
LEFT JOIN workspace_members wm ON u.id = wm.user_id
LEFT JOIN workspaces w ON wm.workspace_id = w.id;
```

### Статистика по каналам

```sql
SELECT 
    w.name as workspace,
    c.name as channel,
    c.type,
    COUNT(m.id) as message_count
FROM channels c
LEFT JOIN workspaces w ON c.workspace_id = w.id
LEFT JOIN messages m ON c.id = m.channel_id
GROUP BY w.name, c.name, c.type
ORDER BY message_count DESC;
```

### Активность пользователей

```sql
SELECT 
    u.display_name,
    u.email,
    COUNT(m.id) as message_count
FROM users u
LEFT JOIN messages m ON u.id = m.user_id
GROUP BY u.id, u.display_name, u.email
ORDER BY message_count DESC;
```

## Подключение через внешние инструменты

Если хотите использовать внешние инструменты (DataGrip, DBeaver, pgAdmin):

**Параметры подключения:**
- Host: `localhost`
- Port: `5433` (внешний порт)
- Database: `atlas`
- Username: `atlas`
- Password: `atlas`
- SSL Mode: `disable`

## Резервное копирование и восстановление

```powershell
# Создать дамп БД
docker exec atlas-postgres pg_dump -U atlas atlas > backup.sql

# Восстановить из дампа
docker exec -i atlas-postgres psql -U atlas atlas < backup.sql
```
