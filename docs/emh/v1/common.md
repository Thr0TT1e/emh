# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/common.proto](#emh-v1-common-proto)
  - **Messages**
    - [PaginationRequest](#emh-v1-paginationrequest)
    - [PaginationResponse](#emh-v1-paginationresponse)
    - [AuditInfo](#emh-v1-auditinfo)
    - [FlexibleDate](#emh-v1-flexibledate)

<a name="emh-v1-common-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/common.proto

**Package:** `emh.v1`

<a name="emh-v1-paginationrequest"></a>

### PaginationRequest

PaginationRequest параметры cursor-based пагинации для списков.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| cursor | string | optional | cursor - непрозрачный курсор из предыдущего ответа. Пустая строка для первой страницы. |
| page_size | int32 | optional | page_size - количество элементов на странице. Максимум 100, по умолчанию 20. |

<details>
<summary>JSON Example</summary>

```json
{
  "cursor": "string",
  "pageSize": 0
}
```

</details>

<a name="emh-v1-paginationresponse"></a>

### PaginationResponse

PaginationResponse метаданные пагинации для навигации по спискам.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| next_cursor | string | optional | next_cursor - курсор следующей страницы. Пустая строка означает конец списка. |
| total_count | int64 | optional | total_count - общее количество найденных записей (опционально, может быть 0 если подсчет отключен). |

<details>
<summary>JSON Example</summary>

```json
{
  "nextCursor": "string",
  "totalCount": 0
}
```

</details>

<a name="emh-v1-auditinfo"></a>

### AuditInfo

AuditInfo системные метаданные записи для аудита и отслеживания изменений.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| created_at | [Timestamp](#google-protobuf-timestamp) | optional | created_at - дата и время создания записи в UTC. |
| updated_at | [Timestamp](#google-protobuf-timestamp) | optional | updated_at - дата и время последнего обновления записи в UTC. |
| created_by | string | optional | created_by - идентификатор пользователя или системы, создавшей запись. |

<details>
<summary>JSON Example</summary>

```json
{
  "createdAt": {
    "nanos": 0,
    "seconds": 0
  },
  "createdBy": "string",
  "updatedAt": {
    "nanos": 0,
    "seconds": 0
  }
}
```

</details>

<a name="emh-v1-flexibledate"></a>

### FlexibleDate

FlexibleDate представляет гибкую дату с уровнем точности и текстовым представлением.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| anchor_date | [Timestamp](#google-protobuf-timestamp) | optional | anchor_date - опорная дата для сортировки и фильтрации. Для точной даты это YYYY-MM-DD. Для месяца - первое число месяца. Для года - 1 января. Для сезона - условная дата начала сезона. Для DAY_MONTH и UNKNOWN должно быть пустым. |
| precision | DatePrecision | optional | precision - уровень точности даты. |
| display_text | string | optional | display_text - человекочитаемое представление даты из источника. Например: "28 июля", "Февраль 1994", "Лето 1989". |

<details>
<summary>JSON Example</summary>

```json
{
  "anchorDate": {
    "nanos": 0,
    "seconds": 0
  },
  "displayText": "string",
  "precision": "DatePrecision_VALUE"
}
```

</details>

