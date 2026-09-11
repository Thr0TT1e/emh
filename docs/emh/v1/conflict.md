# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/conflict.proto](#emh-v1-conflict-proto)
  - **Services**
    - [ConflictService](#emh-v1-conflictservice)
  - **Messages**
    - [Conflict](#emh-v1-conflict)
    - [GetConflictRequest](#emh-v1-getconflictrequest)
    - [GetConflictResponse](#emh-v1-getconflictresponse)
    - [ListConflictsRequest](#emh-v1-listconflictsrequest)
    - [ListConflictsResponse](#emh-v1-listconflictsresponse)

<a name="emh-v1-conflict-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/conflict.proto

**Package:** `emh.v1`

<a name="emh-v1-conflictservice"></a>

## ConflictService

ConflictService публичный сервис для чтения справочника конфликтов.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [GetConflict](#emh-v1-conflictservice-getconflict) | ➡️ Unary | — | GetConflict возвращает детальную инфо... |
| [ListConflicts](#emh-v1-conflictservice-listconflicts) | ➡️ Unary | — | ListConflicts возвращает отфильтрован�... |

<a name="emh-v1-conflictservice-getconflict"></a>

### GetConflict

```protobuf
rpc GetConflict([GetConflictRequest](#emh-v1-getconflictrequest)) returns ([GetConflictResponse](#emh-v1-getconflictresponse))
```

GetConflict возвращает детальную информацию о конкретном конфликте по ID.

#### Request Example

```json
{
  "id": "string"
}
```

#### Response Example

```json
{
  "conflict": {
    "description": "string",
    "endDate": {
      "nanos": 0,
      "seconds": 0
    },
    "endDateInfo": {
      "anchorDate": {
        "nanos": 0,
        "seconds": 0
      },
      "displayText": "string",
      "precision": "DatePrecision_VALUE"
    },
    "id": "string",
    "name": "string",
    "parentConflictId": "string",
    "startDate": {
      "nanos": 0,
      "seconds": 0
    },
    "startDateInfo": {
      "anchorDate": {
        "nanos": 0,
        "seconds": 0
      },
      "displayText": "string",
      "precision": "DatePrecision_VALUE"
    },
    "type": "ConflictType_VALUE"
  }
}
```

---

<a name="emh-v1-conflictservice-listconflicts"></a>

### ListConflicts

```protobuf
rpc ListConflicts([ListConflictsRequest](#emh-v1-listconflictsrequest)) returns ([ListConflictsResponse](#emh-v1-listconflictsresponse))
```

ListConflicts возвращает отфильтрованный список всех конфликтов.

#### Request Example

```json
{
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  },
  "parentConflictId": "string",
  "type": "ConflictType_VALUE"
}
```

#### Response Example

```json
{
  "conflicts": [
    {
      "description": "string",
      "endDate": {
        "nanos": 0,
        "seconds": 0
      },
      "endDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "id": "string",
      "name": "string",
      "parentConflictId": "string",
      "startDate": {
        "nanos": 0,
        "seconds": 0
      },
      "startDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "type": "ConflictType_VALUE"
    }
  ],
  "pagination": {
    "nextCursor": "string",
    "totalCount": 0
  }
}
```

---

<a name="emh-v1-conflict"></a>

### Conflict

Conflict представляет военный конфликт или отдельную операцию.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - уникальный идентификатор конфликта (UUID). |
| name | string | optional | name - официальное название конфликта (например, "Великая Отечественная война"). |
| description | string | optional | description - краткое историческое описание конфликта. |
| type | ConflictType | optional | type - категория конфликта для фильтрации. |
| start_date | [Timestamp](#google-protobuf-timestamp) | optional | start_date - дата начала конфликта. |
| end_date | [Timestamp](#google-protobuf-timestamp) | optional | end_date - дата окончания конфликта. Может быть пустой для текущих конфликтов. |
| parent_conflict_id | string | optional | parent_conflict_id - ID родительского конфликта для иерархии (например, битва внутри войны). |
| start_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | start_date_info - гибкая дата начала конфликта. Если заполнено, имеет приоритет над start_date. |
| end_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | end_date_info - гибкая дата окончания конфликта. Если заполнено, имеет приоритет над end_date. |

<details>
<summary>JSON Example</summary>

```json
{
  "description": "string",
  "endDate": {
    "nanos": 0,
    "seconds": 0
  },
  "endDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "id": "string",
  "name": "string",
  "parentConflictId": "string",
  "startDate": {
    "nanos": 0,
    "seconds": 0
  },
  "startDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "type": "ConflictType_VALUE"
}
```

</details>

<a name="emh-v1-getconflictrequest"></a>

### GetConflictRequest

GetConflictRequest запрос на получение детальной информации о конфликте.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID конфликта. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-getconflictresponse"></a>

### GetConflictResponse

GetConflictResponse содержит полную информацию о запрошенном конфликте.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| conflict | [Conflict](#emh-v1-conflict) | optional | conflict - объект конфликта. |

<details>
<summary>JSON Example</summary>

```json
{
  "conflict": {
    "description": "string",
    "endDate": {
      "nanos": 0,
      "seconds": 0
    },
    "endDateInfo": {
      "anchorDate": {
        "nanos": 0,
        "seconds": 0
      },
      "displayText": "string",
      "precision": "DatePrecision_VALUE"
    },
    "id": "string",
    "name": "string",
    "parentConflictId": "string",
    "startDate": {
      "nanos": 0,
      "seconds": 0
    },
    "startDateInfo": {
      "anchorDate": {
        "nanos": 0,
        "seconds": 0
      },
      "displayText": "string",
      "precision": "DatePrecision_VALUE"
    },
    "type": "ConflictType_VALUE"
  }
}
```

</details>

<a name="emh-v1-listconflictsrequest"></a>

### ListConflictsRequest

ListConflictsRequest параметры фильтрации и выборки списка конфликтов.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | ConflictType | optional | type - фильтр по типу конфликта. UNSPECIFIED возвращает все типы. |
| parent_conflict_id | string | optional | parent_conflict_id - фильтр по родительскому конфликту для получения подчиненных операций. |
| pagination | [PaginationRequest](#emh-v1-paginationrequest) | optional | pagination - параметры курсорной пагинации. |

<details>
<summary>JSON Example</summary>

```json
{
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  },
  "parentConflictId": "string",
  "type": "ConflictType_VALUE"
}
```

</details>

<a name="emh-v1-listconflictsresponse"></a>

### ListConflictsResponse

ListConflictsResponse список конфликтов без пагинации (справочник обычно небольшой).

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| conflicts | [Conflict](#emh-v1-conflict) | repeated | conflicts - массив объектов конфликтов. |
| pagination | [PaginationResponse](#emh-v1-paginationresponse) | optional | pagination - метаданные для перехода к следующей странице. |

<details>
<summary>JSON Example</summary>

```json
{
  "conflicts": [
    {
      "description": "string",
      "endDate": {
        "nanos": 0,
        "seconds": 0
      },
      "endDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "id": "string",
      "name": "string",
      "parentConflictId": "string",
      "startDate": {
        "nanos": 0,
        "seconds": 0
      },
      "startDateInfo": {
        "anchorDate": {
          "nanos": 0,
          "seconds": 0
        },
        "displayText": "string",
        "precision": "DatePrecision_VALUE"
      },
      "type": "ConflictType_VALUE"
    }
  ],
  "pagination": {
    "nextCursor": "string",
    "totalCount": 0
  }
}
```

</details>

