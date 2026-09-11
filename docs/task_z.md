Посмотрю структуру backend и ключевые места (аутентификация, конфиг, обработка входных данных), затем дам рекомендации.

Провёл аудит безопасности backend. Есть находки, приоритетные к исправлению:

## Средний приоритет

4. **HTTP-сервер без таймаутов** — `cmd/api/main.go:386-389`: нет `ReadHeaderTimeout`/`ReadTimeout`/`IdleTimeout` → slowloris. Также стоит задать `connect.WithReadMaxBytes`.
5. **Username enumeration по времени** — `usecase/auth_usecase.go:57-66`: для несуществующего пользователя bcrypt не вызывается. Фикс — фиктивный bcrypt-хеш. Плюс нет per-account lockout.
6. **Refresh-токены без детекции переиспользования** — `auth_usecase.go:72-92`: при попытке повторного использования отозванного токена семейство не инвалидируется (так делают для детекта кражи токена). Заодно: `DeleteExpired` из репозитория, похоже, нигде не вызывается.
7. **Секреты в репозитории** — `config/config.yaml:16,46`: пароль БД и bcrypt-хеш реального админа закоммичены; дефолтный DB URL с паролем в `internal/config/config.go:263`.
8. **Статические API-ключи = вечный admin без аудита identity** — `interceptor/auth.go:166-170`: в контекст кладутся пустые claims; стоит хотя бы логировать, какой ключ использован.

## Низкий приоритет

9. LIKE-инъекция (`%`/`_` не экранируются в `hero_repository.go:452`) — CPU-нагрузка патологическими паттернами.
10. Thumbnail worker без `io.LimitReader` и проверки габаритов декодируемого изображения (`thumbnail_worker.go:118-131`) — decompression bomb, усиливается находкой 2.
11. Аудит аутентификации хранит сырые IP, тогда как контакты хешируются через HMAC-pepper — непоследовательная политика PII.
12. CORS: проверить точное (не паттерн-подобное) сопоставление origin при `AllowCredentials: true`.

**Рекомендуемый порядок:** находки 1–3 первыми — они связаны (подделка IP + анонимная загрузка взаимно усиливают друг друга), потом таймауты сервера и username enumeration. Могу сразу заняться исправлением любого пункта — скажите, с какого начать.


❮ git status -sb
## master...codeberg/master
 M backend/cmd/api/main.go
 M backend/internal/delivery/interceptor/auth.go
 M backend/internal/delivery/interceptor/auth_interceptor_test.go
 M backend/internal/delivery/middleware/ratelimit.go
 M backend/internal/delivery/middleware/ratelimit_test.go
 M backend/internal/delivery/v1/contact_server.go
 M backend/internal/domain/media.go
 M backend/internal/storage/s3/media_storage.go
 M backend/internal/usecase/media_usecase.go
 M backend/internal/usecase/photo_usecase_test.go
 M backend/test/e2e/smoke_test.go
 M infra/podman/Caddyfile
?? .mimosa/
?? backend/.mimosa/
?? backend/internal/netutil/
