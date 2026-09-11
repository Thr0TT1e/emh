# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/conflict_admin.proto](#emh-v1-conflict-admin-proto)
  - **Services**
    - [ConflictAdminService](#emh-v1-conflictadminservice)
  - **Messages**
    - [CreateConflictRequest](#emh-v1-createconflictrequest)
    - [CreateConflictResponse](#emh-v1-createconflictresponse)
    - [UpdateConflictRequest](#emh-v1-updateconflictrequest)
    - [UpdateConflictResponse](#emh-v1-updateconflictresponse)
    - [DeleteConflictRequest](#emh-v1-deleteconflictrequest)
    - [DeleteConflictResponse](#emh-v1-deleteconflictresponse)

<a name="emh-v1-conflict-admin-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/conflict_admin.proto

**Package:** `emh.v1`

<a name="emh-v1-conflictadminservice"></a>

## ConflictAdminService

ConflictAdminService защищенный сервис для управления справочником конфликтов.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [CreateConflict](#emh-v1-conflictadminservice-createconflict) | ➡️ Unary | — | CreateConflict создает новую запись о к�... |
| [UpdateConflict](#emh-v1-conflictadminservice-updateconflict) | ➡️ Unary | — | UpdateConflict частично обновляет суще�... |
| [DeleteConflict](#emh-v1-conflictadminservice-deleteconflict) | ➡️ Unary | — | DeleteConflict удаляет конфликт. При на�... |

<a name="emh-v1-conflictadminservice-createconflict"></a>

### CreateConflict

```protobuf
rpc CreateConflict([CreateConflictRequest](#emh-v1-createconflictrequest)) returns ([CreateConflictResponse](#emh-v1-createconflictresponse))
```

CreateConflict создает новую запись о конфликте.

#### Request Example

```json
{
  "description": "string",
  "endDate": "string",
  "endDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "name": "string",
  "parentConflictId": "string",
  "startDate": "string",
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

#### Response Example

```json
{
  "id": "string"
}
```

---

<a name="emh-v1-conflictadminservice-updateconflict"></a>

### UpdateConflict

```protobuf
rpc UpdateConflict([UpdateConflictRequest](#emh-v1-updateconflictrequest)) returns ([UpdateConflictResponse](#emh-v1-updateconflictresponse))
```

UpdateConflict частично обновляет существующий конфликт по field_mask.

#### Request Example

```json
{
  "description": "string",
  "endDate": "string",
  "endDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "fieldMask": [
    "string"
  ],
  "id": "string",
  "name": "string",
  "parentConflictId": "string",
  "startDate": "string",
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

<a name="emh-v1-conflictadminservice-deleteconflict"></a>

### DeleteConflict

```protobuf
rpc DeleteConflict([DeleteConflictRequest](#emh-v1-deleteconflictrequest)) returns ([DeleteConflictResponse](#emh-v1-deleteconflictresponse))
```

DeleteConflict удаляет конфликт. При наличии связанных героев вернет ошибку.

#### Request Example

```json
{
  "id": "string"
}
```

#### Response Example

```json
{
  "success": true
}
```

---

<a name="emh-v1-createconflictrequest"></a>

### CreateConflictRequest

CreateConflictRequest данные для создания нового конфликта. Даты в формате RFC3339.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string | optional | name - обязательное название конфликта. |
| description | string | optional | description - опциональное описание. |
| type | ConflictType | optional | type - обязательный тип конфликта. |
| start_date | string | optional | start_date - дата начала в формате RFC3339. |
| end_date | string | optional | end_date - дата окончания в формате RFC3339. Может быть пустой. |
| parent_conflict_id | string | optional | parent_conflict_id - ID родительского конфликта. |
| start_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | start_date_info - гибкая дата начала конфликта. Если заполнено, имеет приоритет над start_date. |
| end_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | end_date_info - гибкая дата окончания конфликта. Если заполнено, имеет приоритет над end_date. |

<details>
<summary>JSON Example</summary>

```json
{
  "description": "string",
  "endDate": "string",
  "endDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "name": "string",
  "parentConflictId": "string",
  "startDate": "string",
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

<a name="emh-v1-createconflictresponse"></a>

### CreateConflictResponse

CreateConflictResponse результат успешного создания конфликта.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID созданного конфликта. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-updateconflictrequest"></a>

### UpdateConflictRequest

UpdateConflictRequest данные для частичного обновления конфликта.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID обновляемого конфликта (обязательно). |
| name | string | optional | name - новое название. |
| description | string | optional | description - новое описание. |
| type | ConflictType | optional | type - новый тип. |
| start_date | string | optional | start_date - новая дата начала. |
| end_date | string | optional | end_date - новая дата окончания. |
| parent_conflict_id | string | optional | parent_conflict_id - новый родительский конфликт. |
| field_mask | string | repeated | field_mask - список полей для обновления. |
| start_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | start_date_info - новая гибкая дата начала. |
| end_date_info | [FlexibleDate](#emh-v1-flexibledate) | optional | end_date_info - новая гибкая дата окончания. |

<details>
<summary>JSON Example</summary>

```json
{
  "description": "string",
  "endDate": "string",
  "endDateInfo": {
    "anchorDate": {
      "nanos": 0,
      "seconds": 0
    },
    "displayText": "string",
    "precision": "DatePrecision_VALUE"
  },
  "fieldMask": [
    "string"
  ],
  "id": "string",
  "name": "string",
  "parentConflictId": "string",
  "startDate": "string",
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

<a name="emh-v1-updateconflictresponse"></a>

### UpdateConflictResponse

UpdateConflictResponse обновленный объект конфликта.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| conflict | [Conflict](#emh-v1-conflict) | optional | conflict - актуальное состояние конфликта после обновления. |

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

<a name="emh-v1-deleteconflictrequest"></a>

### DeleteConflictRequest

DeleteConflictRequest запрос на удаление конфликта.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID удаляемого конфликта. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-deleteconflictresponse"></a>

### DeleteConflictResponse

DeleteConflictResponse подтверждение удаления.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| success | bool | optional | success - true если удаление прошло успешно. |

<details>
<summary>JSON Example</summary>

```json
{
  "success": true
}
```

</details>

