# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/api_key_admin.proto](#emh-v1-api-key-admin-proto)
  - **Services**
    - [ApiKeyAdminService](#emh-v1-apikeyadminservice)
  - **Messages**
    - [ApiKey](#emh-v1-apikey)
    - [CreateApiKeyRequest](#emh-v1-createapikeyrequest)
    - [CreateApiKeyResponse](#emh-v1-createapikeyresponse)
    - [ListApiKeysRequest](#emh-v1-listapikeysrequest)
    - [ListApiKeysResponse](#emh-v1-listapikeysresponse)
    - [RevokeApiKeyRequest](#emh-v1-revokeapikeyrequest)
    - [RevokeApiKeyResponse](#emh-v1-revokeapikeyresponse)

<a name="emh-v1-api-key-admin-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/api_key_admin.proto

**Package:** `emh.v1`

<a name="emh-v1-apikeyadminservice"></a>

## ApiKeyAdminService

ApiKeyAdminService защищённый сервис управления API-ключами.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [CreateApiKey](#emh-v1-apikeyadminservice-createapikey) | ➡️ Unary | — | CreateApiKey создаёт ключ и возвращает... |
| [ListApiKeys](#emh-v1-apikeyadminservice-listapikeys) | ➡️ Unary | — | ListApiKeys возвращает список ключей �... |
| [RevokeApiKey](#emh-v1-apikeyadminservice-revokeapikey) | ➡️ Unary | — | RevokeApiKey отзывает ключ (без возмож�... |

<a name="emh-v1-apikeyadminservice-createapikey"></a>

### CreateApiKey

```protobuf
rpc CreateApiKey([CreateApiKeyRequest](#emh-v1-createapikeyrequest)) returns ([CreateApiKeyResponse](#emh-v1-createapikeyresponse))
```

CreateApiKey создаёт ключ и возвращает его полное значение (один раз).

#### Request Example

```json
{
  "description": "string",
  "expiresAt": {
    "nanos": 0,
    "seconds": 0
  },
  "name": "string",
  "role": "string"
}
```

#### Response Example

```json
{
  "apiKey": {
    "createdAt": {
      "nanos": 0,
      "seconds": 0
    },
    "createdBy": "string",
    "description": "string",
    "expiresAt": {
      "nanos": 0,
      "seconds": 0
    },
    "id": "string",
    "keyId": "string",
    "lastUsedAt": {
      "nanos": 0,
      "seconds": 0
    },
    "name": "string",
    "revokedAt": {
      "nanos": 0,
      "seconds": 0
    },
    "role": "string"
  },
  "fullKey": "string"
}
```

---

<a name="emh-v1-apikeyadminservice-listapikeys"></a>

### ListApiKeys

```protobuf
rpc ListApiKeys([ListApiKeysRequest](#emh-v1-listapikeysrequest)) returns ([ListApiKeysResponse](#emh-v1-listapikeysresponse))
```

ListApiKeys возвращает список ключей с их статусами.

#### Request Example

```json
{
  "includeRevoked": true
}
```

#### Response Example

```json
{
  "apiKeys": [
    {
      "createdAt": {
        "nanos": 0,
        "seconds": 0
      },
      "createdBy": "string",
      "description": "string",
      "expiresAt": {
        "nanos": 0,
        "seconds": 0
      },
      "id": "string",
      "keyId": "string",
      "lastUsedAt": {
        "nanos": 0,
        "seconds": 0
      },
      "name": "string",
      "revokedAt": {
        "nanos": 0,
        "seconds": 0
      },
      "role": "string"
    }
  ]
}
```

---

<a name="emh-v1-apikeyadminservice-revokeapikey"></a>

### RevokeApiKey

```protobuf
rpc RevokeApiKey([RevokeApiKeyRequest](#emh-v1-revokeapikeyrequest)) returns ([RevokeApiKeyResponse](#emh-v1-revokeapikeyresponse))
```

RevokeApiKey отзывает ключ (без возможности восстановления).

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

<a name="emh-v1-apikey"></a>

### ApiKey

ApiKey метаданные API-ключа (без секретной части).

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID записи ключа. |
| key_id | string | optional | key_id - публичный идентификатор ключа (не секрет). |
| name | string | optional | name - человекочитаемое имя ключа. |
| description | string | optional | description - описание назначения ключа. |
| role | string | optional | role - права ключа. |
| created_by | string | optional | created_by - кто создал ключ. |
| revoked_at | [Timestamp](#google-protobuf-timestamp) | optional | revoked_at - время отзыва (пусто если активен). |
| expires_at | [Timestamp](#google-protobuf-timestamp) | optional | expires_at - срок действия (пусто если бессрочный). |
| last_used_at | [Timestamp](#google-protobuf-timestamp) | optional | last_used_at - время последнего использования. |
| created_at | [Timestamp](#google-protobuf-timestamp) | optional | created_at - время создания. |

<details>
<summary>JSON Example</summary>

```json
{
  "createdAt": {
    "nanos": 0,
    "seconds": 0
  },
  "createdBy": "string",
  "description": "string",
  "expiresAt": {
    "nanos": 0,
    "seconds": 0
  },
  "id": "string",
  "keyId": "string",
  "lastUsedAt": {
    "nanos": 0,
    "seconds": 0
  },
  "name": "string",
  "revokedAt": {
    "nanos": 0,
    "seconds": 0
  },
  "role": "string"
}
```

</details>

<a name="emh-v1-createapikeyrequest"></a>

### CreateApiKeyRequest

CreateApiKeyRequest запрос на создание API-ключа.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string | optional | name - имя ключа (обязательно). |
| description | string | optional | description - описание назначения. |
| role | string | optional | role - права ключа (по умолчанию admin). |
| expires_at | [Timestamp](#google-protobuf-timestamp) | optional | expires_at - срок действия (опционально). |

<details>
<summary>JSON Example</summary>

```json
{
  "description": "string",
  "expiresAt": {
    "nanos": 0,
    "seconds": 0
  },
  "name": "string",
  "role": "string"
}
```

</details>

<a name="emh-v1-createapikeyresponse"></a>

### CreateApiKeyResponse

CreateApiKeyResponse результат создания ключа.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| api_key | [ApiKey](#emh-v1-apikey) | optional | api_key - метаданные созданного ключа. |
| full_key | string | optional | full_key - полный ключ. Показывается ТОЛЬКО ОДИН РАЗ — сохраните его. |

<details>
<summary>JSON Example</summary>

```json
{
  "apiKey": {
    "createdAt": {
      "nanos": 0,
      "seconds": 0
    },
    "createdBy": "string",
    "description": "string",
    "expiresAt": {
      "nanos": 0,
      "seconds": 0
    },
    "id": "string",
    "keyId": "string",
    "lastUsedAt": {
      "nanos": 0,
      "seconds": 0
    },
    "name": "string",
    "revokedAt": {
      "nanos": 0,
      "seconds": 0
    },
    "role": "string"
  },
  "fullKey": "string"
}
```

</details>

<a name="emh-v1-listapikeysrequest"></a>

### ListApiKeysRequest

ListApiKeysRequest запрос списка ключей.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| include_revoked | bool | optional | include_revoked - включать ли отозванные ключи. |

<details>
<summary>JSON Example</summary>

```json
{
  "includeRevoked": true
}
```

</details>

<a name="emh-v1-listapikeysresponse"></a>

### ListApiKeysResponse

ListApiKeysResponse список ключей.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| api_keys | [ApiKey](#emh-v1-apikey) | repeated | api_keys - массив ключей (без секретов). |

<details>
<summary>JSON Example</summary>

```json
{
  "apiKeys": [
    {
      "createdAt": {
        "nanos": 0,
        "seconds": 0
      },
      "createdBy": "string",
      "description": "string",
      "expiresAt": {
        "nanos": 0,
        "seconds": 0
      },
      "id": "string",
      "keyId": "string",
      "lastUsedAt": {
        "nanos": 0,
        "seconds": 0
      },
      "name": "string",
      "revokedAt": {
        "nanos": 0,
        "seconds": 0
      },
      "role": "string"
    }
  ]
}
```

</details>

<a name="emh-v1-revokeapikeyrequest"></a>

### RevokeApiKeyRequest

RevokeApiKeyRequest запрос на отзыв ключа.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID отзывааемого ключа. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-revokeapikeyresponse"></a>

### RevokeApiKeyResponse

RevokeApiKeyResponse подтверждение отзыва.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| success | bool | optional | success - true если ключ отозван. |

<details>
<summary>JSON Example</summary>

```json
{
  "success": true
}
```

</details>

