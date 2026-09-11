# Справка для Frontend: Изменения в Backend API (Спринт 7)

> **Дата:** 11 сентября 2026  
> **Backend:** Go 1.26.5, Connect RPC v1.20.0  
> **Статус:** 🟢 Все изменения применены, тесты зелёные

## TL;DR

В Спринте 7 закрыто 5 задач, из которых **3 влияют на фронтенд**:

| # | Задача | Влияние на фронт |
|---|---|---|
| #6 | Анти-дубликат + аудит модерации заявок | 🔴 **Сильное**: новый RPC, новый код ошибки |
| #7 | Email-уведомления заявителям | 🟢 Минимальное: только тост |
| #8 | Пагинация справочников + gzip | 🟡 Среднее: новые поля в контрактах |

---

## 1. ⚠️ Важные коррекции к вашему анализу

Прежде чем переходить к изменениям, разберу три заблуждения из вашего сообщения:

### ❌ "В `submission.proto` нет новых RPC для фич Спринта 7"

**На самом деле есть.** В задаче #6 добавлен новый RPC в `SubmissionAdminService`:

```protobuf
rpc ListSubmissionReviews(ListSubmissionReviewsRequest) 
  returns (ListSubmissionReviewsResponse);
```

Это позволяет получить **полную историю модерации** заявки: кто, когда и какое решение принял.

### ❌ "Нужен отдельный `CheckSubmissionUniqueness` RPC"

**Не нужен.** Бэкенд **уже проверяет уникальность** на этапе `CreateSubmission` через SHA-256 хеш нормализованного `payload_json`. При совпадении с активной заявкой возвращается:

- **gRPC код:** `CodeAlreadyExists` (HTTP 409)
- **Сообщение:** "Заявка с таким содержанием уже существует"
- **Domain код:** `ErrCodeSubmissionDuplicate`

Фронтенд просто должен **красиво обработать эту ошибку** — и пользователь увидит понятное сообщение.

### ❌ "Нет даты рассмотрения"

**Есть.** Поле `Submission.audit.updated_at` обновляется при каждом Review. Для status ≠ DRAFT это и есть дата рассмотрения.

---

## 2. 🔴 Критические изменения (задача #6)

### 2.1. Новый RPC: `ListSubmissionReviews`

**Назначение:** Получить историю модерации конкретной заявки.

```protobuf
// emh/v1/submission.proto

message SubmissionReview {
    string id = 1;                              // UUID записи аудита
    string submission_id = 2;                   // UUID заявки
    string reviewer_name = 3;                   // Имя модератора (из JWT)
    SubmissionReviewDecision decision = 4;      // APPROVE или REJECT
    string comment = 5;                         // Комментарий модератора
    google.protobuf.Timestamp created_at = 6;   // Дата решения
}

message ListSubmissionReviewsRequest {
    string submission_id = 1;  // UUID заявки
}

message ListSubmissionReviewsResponse {
    repeated SubmissionReview reviews = 1;  // История (последние сверху)
}

service SubmissionAdminService {
    // ... существующие RPC ...
    rpc ListSubmissionReviews(ListSubmissionReviewsRequest) 
      returns (ListSubmissionReviewsResponse);
}
```

**Когда использовать:**
- В диалоге рассмотрения **уже рассмотренной** заявки — показать историю
- В карточке заявки — блок "История модерации"
- При открытии архивной заявки — увидеть, кто и когда её рассматривал

### 2.2. Новый код ошибки: `CodeAlreadyExists`

**Когда возникает:**
Пользователь отправляет заявку с `payload_json`, **семантически идентичным** активной (DRAFT/PUBLISHED) заявке.

Примеры совпадения (детектируются как дубликат):
```json
// Заявка 1
{"first_name": "Пётр", "last_name": "Иванов", "rank": "рядовой"}

// Заявка 2 (другой порядок ключей, другое форматирование — НО ТОТ ЖЕ ХЕШ!)
{
  "rank": "рядовой",
  "last_name": "Иванов",
  "first_name": "Пётр"
}
```

**Что показывать пользователю:**
```
Такая заявка уже существует и находится на рассмотрении.
Пожалуйста, дождитесь ответа модератора или отправьте заявку
с дополнительными данными.
```

**Важно:** Отклонённые (REJECTED) заявки **не считаются дубликатами** — их можно переотправлять после исправлений.

### 2.3. TypeScript пример обработки

```typescript
// composables/useSubmission.ts
import { Code } from '@connectrpc/connect'

export async function createSubmission(params: CreateSubmissionParams) {
  try {
    const resp = await submissionClient.createSubmission(params)
    return { success: true, submissionId: resp.submissionId }
  } catch (err) {
    if (isConnectError(err)) {
      switch (err.code) {
        case Code.AlreadyExists:
          return {
            success: false,
            error: 'duplicate',
            message: 'Такая заявка уже существует и находится на рассмотрении'
          }
        case Code.InvalidArgument:
          return {
            success: false,
            error: 'validation',
            message: err.message
          }
      }
    }
    throw err
  }
}
```

### 2.4. Модальное окно "История модерации"

```vue
<!-- components/admin/SubmissionReviewHistory.vue -->
<template>
  <Dialog v-model:visible="visible" header="История модерации">
    <div v-if="loading">Загрузка...</div>
    <div v-else-if="reviews.length === 0">
      Заявка ещё не рассматривалась
    </div>
    <Timeline v-else :value="reviews">
      <template #content="slotProps">
        <div class="review-item">
          <Tag 
            :severity="slotProps.data.decision === 1 ? 'success' : 'danger'"
            :value="slotProps.data.decision === 1 ? 'Одобрено' : 'Отклонено'"
          />
          <div class="reviewer">{{ slotProps.data.reviewerName }}</div>
          <div class="date">{{ formatDate(slotProps.data.createdAt) }}</div>
          <div v-if="slotProps.data.comment" class="comment">
            {{ slotProps.data.comment }}
          </div>
        </div>
      </template>
    </Timeline>
  </Dialog>
</template>

<script setup lang="ts">
const reviews = ref<SubmissionReview[]>([])

async function loadHistory(submissionId: string) {
  const resp = await submissionAdminClient.listSubmissionReviews({
    submissionId
  })
  reviews.value = resp.reviews
}
</script>
```

---

## 3. 🟡 Изменения в справочниках (задача #8)

### 3.1. Пагинация для `ListConflicts`, `ListLocations`, `ListAwards`

Все три метода теперь поддерживают **курсорную пагинацию** (единообразно с `ListHeroes`).

**Было:**
```protobuf
message ListConflictsRequest {
    ConflictType type = 1;
    string parent_conflict_id = 2;
}

message ListConflictsResponse {
    repeated Conflict conflicts = 1;
}
```

**Стало:**
```protobuf
message ListConflictsRequest {
    ConflictType type = 1;
    string parent_conflict_id = 2;
    PaginationRequest pagination = 3;  // 🆕
}

message ListConflictsResponse {
    repeated Conflict conflicts = 1;
    PaginationResponse pagination = 2;  // 🆕
}
```

Аналогично для `ListLocationsRequest/Response` и `ListAwardsRequest/Response`.

### 3.2. Обратная совместимость

**Старые клиенты продолжают работать.** Если `pagination` не передан:
- `page_size` = 20 (default)
- `cursor` = "" (первая страница)
- Ответ содержит `pagination.next_cursor` и `pagination.total_count`

### 3.3. Обратная совместимость полей

В `Conflict`/`Location`/`Award` **ничего не удалялось**, только добавились поля пагинации в Request/Response.

### 3.4. TypeScript пример

```typescript
// Было (работает, но неэффективно для больших списков)
const resp = await conflictClient.listConflicts({})
const conflicts = resp.conflicts

// Стало (рекомендуется для больших справочников)
const loadConflicts = async (cursor?: string) => {
  const resp = await conflictClient.listConflicts({
    pagination: {
      pageSize: BigInt(50),
      cursor: cursor ?? ''
    }
  })
  return {
    conflicts: resp.conflicts,
    nextCursor: resp.pagination?.nextCursor,
    totalCount: Number(resp.pagination?.totalCount ?? 0)
  }
}
```

### 3.5. Когда внедрять пагинацию на фронте

**Рекомендация:** Отложить. Сейчас справочники небольшие:
- Конфликты: ~50 записей
- Локации: ~200 записей
- Награды: ~100 записей

Пагинация понадобится, когда справочники вырастут до 500+ записей. Сейчас можно оставить `pageSize=100` и загружать всё за один запрос.

---

## 4. 🟢 Прозрачные изменения

### 4.1. Gzip compression

**Backend автоматически сжимает JSON-ответы**, если клиент отправляет `Accept-Encoding: gzip`.

- Connect-Web библиотека **сама обрабатывает** gzip — никаких изменений в коде
- Экономия трафика: 60-80% для больших JSON
- Ускорение загрузки на медленных сетях

**Проверка:**
```bash
curl -H "Accept-Encoding: gzip" \
  -H "Content-Type: application/json" \
  -X POST http://localhost:3480/emh.v1.ConflictService/ListConflicts \
  -d '{}' -v 2>&1 | grep -i "content-encoding"
# Должно показать: < content-encoding: gzip
```

### 4.2. `Submission.content_hash` (read-only)

В proto-сообщение `Submission` добавлено поле:
```protobuf
string content_hash = 10;  // SHA-256 хеш payload_json
```

**Назначение:** Отладка и аудит. На фронте можно **игнорировать** или показывать в devtools.

---

## 5. 📧 Email-уведомления заявителям (задача #7)

### Что происходит на бэкенде

После успешного `ReviewSubmission`:
1. Бэкенд ставит уведомление в очередь `SubmissionEmailWorker`
2. Воркер асинхронно отправляет письмо заявителю
3. Шаблон зависит от решения (Approve/Reject)

### Что делать на фронте

**Опционально** — добавить тост-подтверждение:

```typescript
// После успешного Review
async function handleReview(submissionId: string, decision: Decision) {
  const resp = await submissionAdminClient.reviewSubmission({
    id: submissionId,
    decision,
    moderatorComment: comment.value
  })
  
  toast.add({
    severity: 'success',
    summary: 'Заявка рассмотрена',
    detail: decision === Decision.APPROVE 
      ? 'Заявитель получит уведомление на email' 
      : 'Заявитель получит уведомление с причиной отклонения',
    life: 5000
  })
}
```

**Важно:** Уведомление отправляется **best-effort** — если SMTP недоступен, письмо не придёт, но модерация всё равно завершится успешно.

---

## 6. 🎯 Конкретные ответы на ваши предложения

### ✅ Приоритет 1 — Аудит модерации (можно сделать СЕЙЧАС)

Ваши предложения — **правильные**, но с учётом нового RPC можно сделать ещё лучше:

| Ваше предложение | Статус | Улучшение |
|---|---|---|
| Колонка «Рассмотрена» с `audit.updatedAt` | ✅ Реализуемо | + Показывать `reviewer_name` через `ListSubmissionReviews` |
| `moderatorComment` в списке (тултип) | ✅ Реализуемо | + Кнопка "История" для полной картины |
| История в диалоге рассмотрения | ✅ Реализуемо | ✅ **Новый RPC `ListSubmissionReviews`** |
| Фильтр «Показать только рассмотренные» | ✅ Клиентский | Можно сделать серверный через `status` фильтр (уже есть) |

### ❌ Приоритет 2 — Уникальные заявки (нужен бэкенд)

**Не нужен новый RPC.** Бэкенд уже возвращает `CodeAlreadyExists`. Вместо запроса нового API:

1. ✅ **Обработать ошибку** на фронте (см. раздел 2.3)
2. ✅ **Клиентская мягкая проверка** через `localStorage` для UX:

```typescript
// composables/useDuplicateCheck.ts
const RECENT_SUBMISSIONS_KEY = 'emh_recent_submissions'

export function useDuplicateCheck() {
  const checkLocal = (targetHeroId: string, payloadJson: string) => {
    const recent = JSON.parse(
      localStorage.getItem(RECENT_SUBMISSIONS_KEY) || '[]'
    )
    const hash = simpleHash(payloadJson) // быстрая хеш-функция
    return recent.find(
      (s: any) => s.targetHeroId === targetHeroId && s.hash === hash
    )
  }
  
  const recordSubmission = (targetHeroId: string, payloadJson: string) => {
    const recent = JSON.parse(
      localStorage.getItem(RECENT_SUBMISSIONS_KEY) || '[]'
    )
    recent.unshift({
      targetHeroId,
      hash: simpleHash(payloadJson),
      timestamp: Date.now()
    })
    // Храним только последние 10
    localStorage.setItem(
      RECENT_SUBMISSIONS_KEY, 
      JSON.stringify(recent.slice(0, 10))
    )
  }
  
  return { checkLocal, recordSubmission }
}
```

**Логика на `/submit`:**
1. Перед отправкой — `checkLocal()`
2. Если найдено — показать **предупреждение** (не блокировать!)
3. После успешной отправки — `recordSubmission()`

### ✅ Приоритет 3 — Уведомления

Ваши выводы правильные. Добавляю только одно: уведомление отправляется **асинхронно** — между `ReviewSubmission` и фактической отправкой письма может пройти 100-500мс. Тост "Уведомление отправлено" показываем **сразу** после успешного Review.

---

## 7. 📋 Чек-лист изменений для фронтенда

### Обязательные изменения (P0)

- [ ] **Обработать `CodeAlreadyExists`** в `createSubmission`
  - Показать понятное сообщение "Такая заявка уже существует"
  - Предложить дождаться ответа или отправить с дополнительными данными

- [ ] **Показывать `moderator_comment`** в списке заявок
  - В тултипе или как отдельную колонку
  - Для status ≠ DRAFT

- [ ] **Показывать `audit.updated_at`** для рассмотренных заявок
  - Колонка "Рассмотрена" с датой

### Желательные изменения (P1)

- [ ] **Добавить модальное окно "История модерации"**
  - Использовать новый RPC `ListSubmissionReviews`
  - Timeline с решениями, комментариями и именами модераторов

- [ ] **Тост после Review**
  - "Заявитель получит уведомление на email"

- [ ] **Клиентская мягкая проверка дубликатов**
  - Через `localStorage` на `/submit`
  - Предупреждение перед отправкой

### Отложить (P2)

- [ ] Пагинация справочников (когда вырастут до 500+ записей)
- [ ] Показ `content_hash` в devtools

---

## 8. 🔗 Связанные файлы

### Backend
- `backend/proto/emh/v1/submission.proto` — новый RPC + поля
- `backend/proto/emh/v1/conflict.proto` — пагинация
- `backend/proto/emh/v1/location.proto` — пагинация
- `backend/proto/emh/v1/award.proto` — пагинация
- `backend/internal/domain/errors.go` — `ErrSubmissionDuplicate`
- `backend/internal/delivery/v1/errors.go` — маппинг в `CodeAlreadyExists`

### Frontend (предполагаемые изменения)
- `app/pages/admin/submissions/index.vue` — колонки, фильтр
- `app/pages/submit.vue` — обработка дубликатов
- `app/lib/errors.ts` — обработка `CodeAlreadyExists`
- `app/composables/useSubmission.ts` — новый метод `listSubmissionReviews`

---

## 9. 🧪 Как протестировать изменения

### Тест 1: Анти-дубликат
1. Отправить заявку на `/submit` с определённым `payload_json`
2. Дождаться статуса DRAFT в админке
3. Отправить **ту же самую** заявку снова
4. **Ожидаемо:** Ошибка `CodeAlreadyExists` с понятным сообщением

### Тест 2: История модерации
1. Отправить заявку → рассмотреть (Approve) с комментарием
2. Открыть карточку заявки в админке
3. **Ожидаемо:** Видна история: модератор, дата, комментарий

### Тест 3: Email-уведомление
1. Отправить заявку с валидным email
2. Рассмотреть (Approve или Reject)
3. **Ожидаемо:** Письмо приходит на email заявителя (если SMTP настроен)

### Тест 4: Пагинация справочников
1. Вызвать `ListConflicts` с `pagination.page_size = 5`
2. **Ожидаемо:** Возвращается 5 конфликтов + `next_cursor` + `total_count`
3. Вызвать снова с `cursor = next_cursor`
4. **Ожидаемо:** Следующие 5 конфликтов

---

## 10. 📞 Вопросы к Backend

Если что-то непонятно или нужны дополнительные фичи:

1. **Нужен ли серверный фильтр по дате рассмотрения** в `ListSubmissions`?  
   Сейчас можно фильтровать только по статусу.

2. **Нужен ли `reviewer_name` в `Submission`** (денормализованное поле)?  
   Сейчас нужно делать отдельный запрос `ListSubmissionReviews`.

3. **Нужен ли bulk API** для получения истории нескольких заявок сразу?  
   Сейчас только по одной `submission_id`.
