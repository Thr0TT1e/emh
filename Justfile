# Переменная для подключения к БД. Читается из .env или используется дефолтная.
db_url := env("DATABASE_URL", "postgres://hero_db:secret@localhost:5443/heroes?sslmode=disable")
export MIGRATIONS_DIR := "backend/migrations"

default:
    @just --list

# Создать новую SQL-миграцию
# Использование: just create add_heroes_table
create name:
    @goose -dir {{MIGRATIONS_DIR}} postgres "{{db_url}}" create {{name}} sql

# Применить все ожидающие миграции (Up)
up:
    @goose -dir {{MIGRATIONS_DIR}} postgres "{{db_url}}" up

# Откатить последнюю примененную миграцию (Down)
down:
    @goose -dir {{MIGRATIONS_DIR}} postgres "{{db_url}}" down

# Откатить ВСЕ миграции (Осторожно!)
reset:
    @goose -dir {{MIGRATIONS_DIR}} postgres "{{db_url}}" reset

# Показать статус миграций
status:
    @goose -dir {{MIGRATIONS_DIR}} postgres "{{db_url}}" status

# Подключиться к БД через psql (требует установленный psql на хосте)
psql:
    @psql "{{db_url}}"

# Сгенерировать bcrypt-хеш пароля администратора
# Использование: just hash-password 'надёжный_пароль'
hash-password password:
    @go run -C backend ./cmd/hashpassword {{password}}

# Сгененрировать AUTH_JWT_SECRET
jwt-secret:
    @openssl rand -hex 32

# Проверка корректности proto-файлов и генерация SDK
gen-sdk:
    @easyp lint -r proto && easyp generate

# Сгенерировать IP_PEPPER для хеширования адресов в contact_messages
ip-pepper:
    @openssl rand -hex 32

# --- Интеграционные тесты ---

# Переменная для тестовой БД (можно переопределить через env).
test_db_url := env(
    "TEST_DATABASE_URL",
    "postgres://emh_test:emh_test_secret@localhost:5444/emh_test_db?sslmode=disable"
)
export TEST_DATABASE_URL := test_db_url
migrations_dir := justfile_directory() + "/backend/migrations"
export TEST_MIGRATIONS_DIR := migrations_dir

# Поднять тестовую БД (контейнер + ожидание ready)
test-db-up:
    podman-compose -f infra/podman/test/docker-compose.test.yml up -d
    @echo "Ожидание готовности PostgreSQL..."
    @for i in $(seq 1 30); do \
        if podman exec emh-test-postgres pg_isready -U emh_test -d emh_test_db >/dev/null 2>&1; then \
            echo "PostgreSQL готов."; \
            break; \
        fi; \
        sleep 1; \
    done

# Остановить и удалить тестовую БД
test-db-down:
    podman-compose -f infra/podman/test/docker-compose.test.yml down -v

# Применить миграции к тестовой БД (без создания файла — миграции уже есть)
test-db-migrate:
    goose -dir {{TEST_MIGRATIONS_DIR}} postgres "{{test_db_url}}" up

# Сбросить тестовую БД (все таблицы + пересоздать миграции)
test-db-reset: test-db-down test-db-up test-db-migrate

# Unit-тесты (без БД)
test:
    go test -C backend -race -count=1 -cover ./internal/auth/... ./internal/config/... ./internal/delivery/... ./internal/domain/... ./internal/smtp/... ./internal/usecase/...

# Интеграционные тесты (с тестовой БД, требуют test-db-up)
# -p 1 - последовательный запуск тестов, чтобы избежать deadlock при cleanup
test-int:
    go test -C backend -race -count=1 -p 1 -tags=integration ./internal/repository/pg/... ./test/e2e/...

# Все тесты вместе
test-all: test test-int

# Coverage для unit + integration
cover:
    go test -C backend -race -count=1 -coverprofile=coverage.out \
        ./internal/usecase/... ./internal/smtp/... ./internal/delivery/... \
        ./internal/domain/... ./internal/repository/pg/... ./test/e2e/...
    go tool cover -func=backend/coverage.out | tail -20
    @echo "HTML-отчёт: backend/coverage.html"
    go tool cover -html=backend/coverage.out -o backend/coverage.html
