# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/award_admin.proto](#emh-v1-award-admin-proto)
  - **Services**
    - [AwardAdminService](#emh-v1-awardadminservice)
  - **Messages**
    - [CreateAwardRequest](#emh-v1-createawardrequest)
    - [CreateAwardResponse](#emh-v1-createawardresponse)
    - [UpdateAwardRequest](#emh-v1-updateawardrequest)
    - [UpdateAwardResponse](#emh-v1-updateawardresponse)
    - [DeleteAwardRequest](#emh-v1-deleteawardrequest)
    - [DeleteAwardResponse](#emh-v1-deleteawardresponse)

<a name="emh-v1-award-admin-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/award_admin.proto

**Package:** `emh.v1`

<a name="emh-v1-awardadminservice"></a>

## AwardAdminService

AwardAdminService защищенный сервис для управления справочником наград.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [CreateAward](#emh-v1-awardadminservice-createaward) | ➡️ Unary | — | CreateAward создает новую запись в спр... |
| [UpdateAward](#emh-v1-awardadminservice-updateaward) | ➡️ Unary | — | UpdateAward частично обновляет сущест... |
| [DeleteAward](#emh-v1-awardadminservice-deleteaward) | ➡️ Unary | — | DeleteAward удаляет награду. При налич... |

<a name="emh-v1-awardadminservice-createaward"></a>

### CreateAward

```protobuf
rpc CreateAward([CreateAwardRequest](#emh-v1-createawardrequest)) returns ([CreateAwardResponse](#emh-v1-createawardresponse))
```

CreateAward создает новую запись в справочнике наград.

#### Request Example

```json
{
  "description": "string",
  "imageUrl": "string",
  "name": "string",
  "sortOrder": 0
}
```

#### Response Example

```json
{
  "id": "string"
}
```

---

<a name="emh-v1-awardadminservice-updateaward"></a>

### UpdateAward

```protobuf
rpc UpdateAward([UpdateAwardRequest](#emh-v1-updateawardrequest)) returns ([UpdateAwardResponse](#emh-v1-updateawardresponse))
```

UpdateAward частично обновляет существующую награду по field_mask.

#### Request Example

```json
{
  "description": "string",
  "fieldMask": [
    "string"
  ],
  "id": "string",
  "imageUrl": "string",
  "name": "string",
  "sortOrder": 0
}
```

#### Response Example

```json
{
  "award": {
    "description": "string",
    "id": "string",
    "imageUrl": "string",
    "name": "string",
    "sortOrder": 0
  }
}
```

---

<a name="emh-v1-awardadminservice-deleteaward"></a>

### DeleteAward

```protobuf
rpc DeleteAward([DeleteAwardRequest](#emh-v1-deleteawardrequest)) returns ([DeleteAwardResponse](#emh-v1-deleteawardresponse))
```

DeleteAward удаляет награду. При наличии связей с героями вернет ошибку.

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

<a name="emh-v1-createawardrequest"></a>

### CreateAwardRequest

CreateAwardRequest данные для создания новой награды в справочнике.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string | optional | name - официальное название награды (обязательно). |
| description | string | optional | description - описание или статут награды. |
| image_url | string | optional | image_url - URL загруженного изображения ленты или знака награды. |
| sort_order | int32 | optional | sort_order - приоритет сортировки при отображении (старшие награды выше). |

<details>
<summary>JSON Example</summary>

```json
{
  "description": "string",
  "imageUrl": "string",
  "name": "string",
  "sortOrder": 0
}
```

</details>

<a name="emh-v1-createawardresponse"></a>

### CreateAwardResponse

CreateAwardResponse результат успешного создания награды.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID созданной награды. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-updateawardrequest"></a>

### UpdateAwardRequest

UpdateAwardRequest данные для частичного обновления существующей награды.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID обновляемой награды (обязательно). |
| name | string | optional | name - новое официальное название. |
| description | string | optional | description - новое описание или статут. |
| image_url | string | optional | image_url - новый URL изображения награды. |
| sort_order | int32 | optional | sort_order - новый приоритет сортировки. |
| field_mask | string | repeated | field_mask - список полей для обновления. Позволяет отличить очистку поля от отсутствия изменений. |

<details>
<summary>JSON Example</summary>

```json
{
  "description": "string",
  "fieldMask": [
    "string"
  ],
  "id": "string",
  "imageUrl": "string",
  "name": "string",
  "sortOrder": 0
}
```

</details>

<a name="emh-v1-updateawardresponse"></a>

### UpdateAwardResponse

UpdateAwardResponse обновленный объект награды.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| award | [Award](#emh-v1-award) | optional | award - актуальное состояние награды после обновления. |

<details>
<summary>JSON Example</summary>

```json
{
  "award": {
    "description": "string",
    "id": "string",
    "imageUrl": "string",
    "name": "string",
    "sortOrder": 0
  }
}
```

</details>

<a name="emh-v1-deleteawardrequest"></a>

### DeleteAwardRequest

DeleteAwardRequest запрос на удаление награды из справочника.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID удаляемой награды. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-deleteawardresponse"></a>

### DeleteAwardResponse

DeleteAwardResponse подтверждение операции удаления.

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

