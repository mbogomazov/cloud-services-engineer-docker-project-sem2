# Momo Store — Docker

Магазин пельменей: бэкенд (Go) + фронтенд (Vue.js). Контейнеризация через Docker и Docker Compose.

## Архитектура

```
        :80                  backend-net
браузер ────▶ frontend(nginx) ────────▶ backend(Go API):8081
              SPA + reverse proxy
```

- **frontend** — статическая Vue-сборка на nginx; проксирует API (`/products`, `/categories`, `/orders`, `/auth`, `/health`, `/metrics`, `/csrf`) на бэкенд.
- **backend** — Go API (chi), `:8081`, in-memory store, `/health` + `/metrics`.

| Сервис   | Внешний порт | В контейнере |
|----------|--------------|--------------|
| frontend | `80`         | `8080`       |
| backend  | `8081`       | `8081`       |

> nginx непривилегированный → не может слушать `<1024`, поэтому `8080` внутри, Compose публикует как `80`.

## Запуск

```bash
docker compose up -d --build   # сборка + запуск
docker compose ps              # статус + healthcheck
docker compose down            # остановка
```

- приложение — http://localhost/
- API — http://localhost:8081/products · healthcheck — http://localhost:8081/health

Отдельные образы: `docker build -t docker-project-backend ./backend` (аналогично frontend).

## Конфигурация

**Env (Compose):** `DOCKER_USER` (`local`), `FRONTEND_PORT` (`80`), `BACKEND_PORT` (`8081`), `VUE_APP_API_URL` (`""` = через прокси).
**Build-args:** backend `GO_VERSION` (`1.17`); frontend `NODE_VERSION` (`16`), `VUE_APP_API_URL`, `PUBLIC_PATH` (`/`).

```bash
FRONTEND_PORT=8080 BACKEND_PORT=9090 docker compose up -d
```

## Оптимизация образов

Multi-stage: инструменты сборки (Go toolchain, Node/npm) в финал не попадают.

| Образ    | Финальный базовый образ              | Размер |
|----------|--------------------------------------|--------|
| backend  | `gcr.io/distroless/static:nonroot`   | ~24 MB |
| frontend | `nginxinc/nginx-unprivileged:alpine` | ~78 MB |

- backend — статическая сборка (`CGO_ENABLED=0`, `-ldflags="-w -s"`, `-trimpath`), distroless без shell/пакетов.
- frontend — в финал копируется только `dist`.
- Кэш слоёв: сначала манифесты (`go.mod`/`go.sum`, `package*.json`) → зависимости, затем код.
- `.dockerignore` исключает `.git`, тесты, `node_modules`, `dist`.

## Устойчивость

- **healthchecks** обоих сервисов (backend — бинарник `/healthcheck`, т.к. в distroless нет curl/wget; frontend — `wget`).
- **depends_on** `condition: service_healthy`.
- **restart: unless-stopped**.
- Две bridge-сети: `frontend-net`, `backend-net`.

**Volumes.** Приложение stateless — данные хранятся в памяти (in-memory store, без БД), поэтому персистентные volumes не нужны. Запись на диск исключена намеренно: корневая ФС `read_only`, а временные пути nginx вынесены в `tmpfs`. При добавлении БД достаточно подключить named volume к её сервису.

## Масштабирование и балансировка

Бэкенд stateless. nginx балансирует реплики через DNS Docker (`resolver 127.0.0.11` + runtime-резолв). Базовый compose фиксирует порт `8081`, поэтому для масштабирования — overlay `docker-compose.scale.yml`:

```bash
docker compose -f docker-compose.yml -f docker-compose.scale.yml up -d --scale backend=3
```

## Безопасность

**Образы/контейнеры:** не root (backend `nonroot` uid 65532, frontend `nginx` uid 101); минимальные образы; чистый финал (multi-stage); открыты только `80` и `8081`.

**Изоляция (compose):** `cap_drop: [ALL]`, `no-new-privileges`, `read_only` ФС (writable пути nginx в `tmpfs`), лимиты CPU/памяти (`deploy.resources`).

**Секреты.** Чувствительных данных в образах нет — приложение работает на in-memory store и не требует паролей/токенов в рантайме. Управление секретами:

- Креды DockerHub хранятся в **GitHub Secrets** (`DOCKER_USER`, `DOCKER_PASSWORD`) и подставляются в CI через `${{ secrets.* }}` — в код, образы и логи не попадают.
- Конфигурация передаётся через env-переменные и build-аргументы, а не хардкодится в образ.
- `.dockerignore` исключает `.git`, `.env` и прочие файлы из контекста сборки, чтобы секреты не утекли в слои образа.
- При появлении рантайм-секретов (напр. пароль БД) их следует подключать через **Docker Secrets** (`secrets:` в compose) — монтируются в `/run/secrets/<name>` как файлы, не как env, и не сохраняются в образе.

**Trivy:** сканирование `CRITICAL`/`HIGH` в CI. Локально: `trivy image docker-project-backend:latest`.

## CI/CD

`.github/workflows/deploy.yaml` при push в `main`: (1) сборка + push образов в DockerHub + Trivy; (2) проверка сборки через Docker Compose.

## Файлы

`backend/Dockerfile`, `frontend/Dockerfile`, `frontend/nginx.conf`, `docker-compose.yml`, `docker-compose.scale.yml`, `*/.dockerignore`.
