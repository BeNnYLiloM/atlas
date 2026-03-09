# Deploy через Coolify

## Что уже подготовлено

- `deploy/docker-compose.coolify.yml` — полный стек для Coolify
- `backend/Dockerfile` — сборка и запуск Go backend
- `frontend/Dockerfile` — сборка Vue frontend и раздача через встроенный nginx
- `deploy/.env.coolify.example` — пример переменных окружения

## Что создать в Coolify

1. Подними новый сервер и установи Coolify по официальной инструкции.
2. Подключи Git-репозиторий.
3. Создай ресурс типа `Docker Compose`.
4. В качестве compose-файла укажи `deploy/docker-compose.coolify.yml`.
5. Заполни переменные окружения по примеру `deploy/.env.coolify.example`.

## Какие домены задать

- `frontend` — основной домен приложения, например `app.example.com`
- `backend` — API и WebSocket, например `api.example.com`
- `minio` — публичный домен файлов, например `files.example.com`
- `livekit` — домен звонков, например `livekit.example.com`

`VITE_API_URL`, `VITE_WS_URL`, `MINIO_PUBLIC_URL` и `LIVEKIT_URL` должны совпадать с этими доменами.

## Что важно для звонков

- Для сервиса `livekit` нужно открыть на сервере `TCP 7881` и `UDP 50000-50100`.
- Значения `LIVEKIT_API_KEY` и `LIVEKIT_API_SECRET` должны совпадать с `keys` в `deploy/configs/livekit.coolify.yaml`.
- HTTPS для `frontend`, `backend`, `minio` и `livekit` лучше выпускать через Coolify.

## Дополнительно по MinIO

- Консоль MinIO будет доступна на `http://SERVER_IP:9001`.
- Публичные ссылки на файлы будут строиться из `MINIO_PUBLIC_URL`.

## Что получится после деплоя

- frontend будет доступен по домену приложения
- backend поднимет миграции сам при старте
- MinIO будет отдавать публичные ссылки на загруженные файлы
- LiveKit будет работать как отдельный self-hosted сервис
