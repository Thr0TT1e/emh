#!/usr/bin/env fish
# Проверка P1-3: Rate-limiting
# ⚠️ Запускать ПОСЛЕ test_auth.fish или подождать 1 минуту

set BASE_URL "http://localhost:3480"

# ─── 1. AuthService: 6 запросов, 6-й должен вернуть 429 ─────
echo "═══ 1. Rate-limit AuthService (лимит 5/min) ═══"
echo "Отправляю 6 запросов Login подряд..."
echo ""

set LOGIN_JSON (jq -n '{"username":"admin","password":"wrong"}' | string collect)

for i in (seq 1 6)
    set CODE (curl -s -o /dev/null -w "%{http_code}" \
        -X POST "$BASE_URL/emh.v1.AuthService/Login" \
        -H "Content-Type: application/json" \
        -d "$LOGIN_JSON")

    if test "$CODE" = "429"
        echo "  Request $i: HTTP $CODE ← rate limit сработал ✅"
    else
        echo "  Request $i: HTTP $CODE"
    end
end

# ─── 2. Healthz не лимитируется ─────────────────────────────
echo ""
echo "═══ 2. Healthz (не лимитируется) ═══"
set FAIL_COUNT 0
for i in (seq 1 20)
    set CODE (curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/healthz")
    if test "$CODE" != "200"
        set FAIL_COUNT (math $FAIL_COUNT + 1)
    end
end

if test "$FAIL_COUNT" -eq 0
    echo "✅ PASS: 20/20 запросов healthz вернули 200"
else
    echo "❌ FAIL: $FAIL_COUNT из 20 запросов healthz не вернули 200"
end

# ─── 3. Публичные read-сервисы (лимит 120/min) ─────────────
echo ""
echo "═══ 3. Read-сервисы (лимит 120/min) ═══"
echo "Отправляю 5 запросов ListHeroes подряд..."

set LIST_JSON (jq -n '{"pagination":{"page_size":1}}' | string collect)

for i in (seq 1 5)
    set CODE (curl -s -o /dev/null -w "%{http_code}" \
        -X POST "$BASE_URL/emh.v1.HeroService/ListHeroes" \
        -H "Content-Type: application/json" \
        -d "$LIST_JSON")
    echo "  Request $i: HTTP $CODE"
end

# ─── 4. Админские сервисы не лимитируются ───────────────────
echo ""
echo "═══ 4. Admin-сервисы (JWT, без rate-limit) ═══"
echo "Сначала логинимся..."

set LOGIN_REQ (jq -n '{"username":"admin","password":"<pass>"}' | string collect)
set LOGIN_RESP (curl -s -X POST "$BASE_URL/emh.v1.AuthService/Login" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_REQ")

# Если Login вернул 429 из-за предыдущего теста — ждём и повторяем
set LOGIN_CODE (echo "$LOGIN_RESP" | jq -r '.code // empty')
if test "$LOGIN_CODE" = "resource_exhausted"
    echo "⏳ Лимит Login исчерпан, жду 60 сек..."
    sleep 60
    set LOGIN_RESP (curl -s -X POST "$BASE_URL/emh.v1.AuthService/Login" \
        -H "Content-Type: application/json" \
        -d "$LOGIN_REQ")
end

set TOKEN (echo "$LOGIN_RESP" | jq -r '.accessToken')

echo "Отправляю 20 запросов к HeroAdminService подряд..."
set ADMIN_FAIL 0
set ADMIN_JSON (jq -n '{"id":"00000000-0000-0000-0000-000000000000"}' | string collect)

for i in (seq 1 20)
    set CODE (curl -s -o /dev/null -w "%{http_code}" \
        -X POST "$BASE_URL/emh.v1.HeroAdminService/DeleteHero" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d "$ADMIN_JSON")
    if test "$CODE" = "429"
        set ADMIN_FAIL (math $ADMIN_FAIL + 1)
    end
end

if test "$ADMIN_FAIL" -eq 0
    echo "✅ PASS: 20/20 admin-запросов прошли без 429"
else
    echo "❌ FAIL: $ADMIN_FAIL из 20 admin-запросов получили 429"
end

echo ""
echo "═══ Тест завершён ═══"
