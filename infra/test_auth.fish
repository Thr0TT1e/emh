#!/usr/bin/env fish
# Проверка P1-4: Refresh-токены (Login → Refresh → ротация → Logout)

set BASE_URL "http://localhost:3480"
set USERNAME "admin"
set PASSWORD "password"

# ─── 1. Login ───────────────────────────────────────────────
echo "═══ 1. Login ═══"

set LOGIN_JSON (jq -n --arg u "$USERNAME" --arg p "$PASSWORD" \
    '{"username": $u, "password": $p}' | string collect)

set LOGIN_RESPONSE (curl -s -X POST "$BASE_URL/emh.v1.AuthService/Login" \
    -H "Content-Type: application/json" \
    -d "$LOGIN_JSON")

echo "$LOGIN_RESPONSE" | jq .

# ИСПРАВЛЕНО: camelCase (protobuf-json mapping)
set ACCESS_TOKEN  (echo "$LOGIN_RESPONSE" | jq -r '.accessToken')
set REFRESH_TOKEN (echo "$LOGIN_RESPONSE" | jq -r '.refreshToken')
set EXPIRES_IN    (echo "$LOGIN_RESPONSE" | jq -r '.expiresIn')

echo ""
echo "Access token:  "(string sub -l 30 "$ACCESS_TOKEN")"..."
echo "Refresh token: "(string sub -l 20 "$REFRESH_TOKEN")"..."
echo "Expires in:    $EXPIRES_IN сек"

# ─── 2. Refresh — обмен на новую пару ───────────────────────
echo ""
echo "═══ 2. Refresh (ротация) ═══"

set REFRESH_JSON (jq -n --arg t "$REFRESH_TOKEN" '{"refresh_token": $t}' | string collect)
set REFRESH_RESPONSE (curl -s -X POST "$BASE_URL/emh.v1.AuthService/Refresh" \
    -H "Content-Type: application/json" \
    -d "$REFRESH_JSON")

echo "$REFRESH_RESPONSE" | jq .

# ИСПРАВЛЕНО: camelCase
set NEW_ACCESS_TOKEN  (echo "$REFRESH_RESPONSE" | jq -r '.accessToken')
set NEW_REFRESH_TOKEN (echo "$REFRESH_RESPONSE" | jq -r '.refreshToken')

# ─── 3. Повторный Refresh со СТАРЫМ токеном (должен 401) ────
echo ""
echo "═══ 3. Refresh со старым токеном (ожидаем Unauthenticated) ═══"

set STALE_JSON (jq -n --arg t "$REFRESH_TOKEN" '{"refresh_token": $t}' | string collect)
set STALE_RESPONSE (curl -s -X POST "$BASE_URL/emh.v1.AuthService/Refresh" \
    -H "Content-Type: application/json" \
    -d "$STALE_JSON")

set STALE_CODE (echo "$STALE_RESPONSE" | jq -r '.code // empty')
if test "$STALE_CODE" = "unauthenticated"
    echo "✅ PASS: старый refresh отозван (code=$STALE_CODE)"
else
    echo "❌ FAIL: ожидался unauthenticated, получено:"
    echo "$STALE_RESPONSE" | jq .
end

# ─── 4. Logout — отзыв нового refresh ───────────────────────
echo ""
echo "═══ 4. Logout ═══"

set LOGOUT_JSON (jq -n --arg t "$NEW_REFRESH_TOKEN" '{"refresh_token": $t}' | string collect)
set LOGOUT_RESPONSE (curl -s -X POST "$BASE_URL/emh.v1.AuthService/Logout" \
    -H "Content-Type: application/json" \
    -d "$LOGOUT_JSON")
echo "$LOGOUT_RESPONSE" | jq .

# ─── 5. Refresh после Logout (должен 401) ───────────────────
echo ""
echo "═══ 5. Refresh после Logout (ожидаем Unauthenticated) ═══"

set POST_LOGOUT_JSON (jq -n --arg t "$NEW_REFRESH_TOKEN" '{"refresh_token": $t}' | string collect)
set POST_LOGOUT (curl -s -X POST "$BASE_URL/emh.v1.AuthService/Refresh" \
    -H "Content-Type: application/json" \
    -d "$POST_LOGOUT_JSON")

set POST_LOGOUT_CODE (echo "$POST_LOGOUT" | jq -r '.code // empty')
if test "$POST_LOGOUT_CODE" = "unauthenticated"
    echo "✅ PASS: refresh после logout отклонён (code=$POST_LOGOUT_CODE)"
else
    echo "❌ FAIL: ожидался unauthenticated, получено:"
    echo "$POST_LOGOUT" | jq .
end

# ─── 6. Access-токен работает для админских запросов ────────
echo ""
echo "═══ 6. Access-токен валиден для HeroAdminService ═══"

set ADMIN_JSON (jq -n '{"id":"00000000-0000-0000-0000-000000000000"}' | string collect)
set ADMIN_CODE (curl -s -o /dev/null -w "%{http_code}" \
    -X POST "$BASE_URL/emh.v1.HeroAdminService/DeleteHero" \
    -H "Authorization: Bearer $NEW_ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -d "$ADMIN_JSON")

if test "$ADMIN_CODE" != "401"; and test "$ADMIN_CODE" != "403"
    echo "✅ PASS: JWT принят (HTTP $ADMIN_CODE — NotFound ожидаем для фейкового ID)"
else
    echo "❌ FAIL: JWT отклонён (HTTP $ADMIN_CODE)"
end

echo ""
echo "═══ Тест завершён ═══"
