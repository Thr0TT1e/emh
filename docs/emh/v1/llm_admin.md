# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/llm_admin.proto](#emh-v1-llm-admin-proto)
  - **Services**
    - [LlmAdminService](#emh-v1-llmadminservice)
  - **Messages**
    - [LlmProviderInfo](#emh-v1-llmproviderinfo)
    - [ListLlmProvidersRequest](#emh-v1-listllmprovidersrequest)
    - [ListLlmProvidersResponse](#emh-v1-listllmprovidersresponse)
    - [SetActiveLlmProviderRequest](#emh-v1-setactivellmproviderrequest)
    - [SetActiveLlmProviderResponse](#emh-v1-setactivellmproviderresponse)
    - [TestLlmProviderRequest](#emh-v1-testllmproviderrequest)
    - [TestLlmProviderResponse](#emh-v1-testllmproviderresponse)
    - [UpdateLlmProviderRequest](#emh-v1-updatellmproviderrequest)
    - [UpdateLlmProviderResponse](#emh-v1-updatellmproviderresponse)

<a name="emh-v1-llm-admin-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/llm_admin.proto

**Package:** `emh.v1`

<a name="emh-v1-llmadminservice"></a>

## LlmAdminService

LlmAdminService админский сервис управления LLM-провайдерами.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [ListLlmProviders](#emh-v1-llmadminservice-listllmproviders) | ➡️ Unary | — | ListLlmProviders возвращает список всех ... |
| [SetActiveLlmProvider](#emh-v1-llmadminservice-setactivellmprovider) | ➡️ Unary | — | SetActiveLlmProvider активирует указанный... |
| [TestLlmProvider](#emh-v1-llmadminservice-testllmprovider) | ➡️ Unary | — | TestLlmProvider отправляет тестовый зап... |
| [UpdateLlmProvider](#emh-v1-llmadminservice-updatellmprovider) | ➡️ Unary | — | UpdateLlmProvider обновляет приоритет и �... |

<a name="emh-v1-llmadminservice-listllmproviders"></a>

### ListLlmProviders

```protobuf
rpc ListLlmProviders([ListLlmProvidersRequest](#emh-v1-listllmprovidersrequest)) returns ([ListLlmProvidersResponse](#emh-v1-listllmprovidersresponse))
```

ListLlmProviders возвращает список всех настроенных провайдеров.

#### Response Example

```json
{
  "count": 0,
  "providers": [
    {
      "createdAt": {
        "nanos": 0,
        "seconds": 0
      },
      "id": "string",
      "isActive": true,
      "model": "string",
      "name": "string",
      "notes": "string",
      "priority": 0,
      "type": "string",
      "updatedAt": {
        "nanos": 0,
        "seconds": 0
      }
    }
  ]
}
```

---

<a name="emh-v1-llmadminservice-setactivellmprovider"></a>

### SetActiveLlmProvider

```protobuf
rpc SetActiveLlmProvider([SetActiveLlmProviderRequest](#emh-v1-setactivellmproviderrequest)) returns ([SetActiveLlmProviderResponse](#emh-v1-setactivellmproviderresponse))
```

SetActiveLlmProvider активирует указанный провайдер (остальные деактивируются триггером БД).

#### Request Example

```json
{
  "id": "string"
}
```

#### Response Example

```json
{
  "provider": {
    "createdAt": {
      "nanos": 0,
      "seconds": 0
    },
    "id": "string",
    "isActive": true,
    "model": "string",
    "name": "string",
    "notes": "string",
    "priority": 0,
    "type": "string",
    "updatedAt": {
      "nanos": 0,
      "seconds": 0
    }
  },
  "success": true
}
```

---

<a name="emh-v1-llmadminservice-testllmprovider"></a>

### TestLlmProvider

```protobuf
rpc TestLlmProvider([TestLlmProviderRequest](#emh-v1-testllmproviderrequest)) returns ([TestLlmProviderResponse](#emh-v1-testllmproviderresponse))
```

TestLlmProvider отправляет тестовый запрос к провайдеру для проверки доступности.

#### Request Example

```json
{
  "id": "string",
  "testPrompt": "string"
}
```

#### Response Example

```json
{
  "errorMessage": "string",
  "processingTimeMs": 0,
  "rawResponse": "string",
  "success": true
}
```

---

<a name="emh-v1-llmadminservice-updatellmprovider"></a>

### UpdateLlmProvider

```protobuf
rpc UpdateLlmProvider([UpdateLlmProviderRequest](#emh-v1-updatellmproviderrequest)) returns ([UpdateLlmProviderResponse](#emh-v1-updatellmproviderresponse))
```

UpdateLlmProvider обновляет приоритет и заметки провайдера.

#### Request Example

```json
{
  "fieldMask": [
    "string"
  ],
  "id": "string",
  "notes": "string",
  "priority": 0
}
```

#### Response Example

```json
{
  "provider": {
    "createdAt": {
      "nanos": 0,
      "seconds": 0
    },
    "id": "string",
    "isActive": true,
    "model": "string",
    "name": "string",
    "notes": "string",
    "priority": 0,
    "type": "string",
    "updatedAt": {
      "nanos": 0,
      "seconds": 0
    }
  }
}
```

---

<a name="emh-v1-llmproviderinfo"></a>

### LlmProviderInfo

LlmProviderInfo информация о провайдере для админки.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - уникальный идентификатор записи провайдера в БД (UUID). |
| name | string | optional | name - уникальное имя провайдера (совпадает с конфигом). |
| type | string | optional | type - тип провайдера (ollama, openai_compatible, anthropic). |
| model | string | optional | model - название используемой модели. |
| is_active | bool | optional | is_active - является ли провайдер текущим активным. |
| priority | int32 | optional | priority - приоритет провайдера (для fallback-логики). |
| notes | string | optional | notes - заметки администратора. |
| created_at | [Timestamp](#google-protobuf-timestamp) | optional | created_at - дата и время создания записи. |
| updated_at | [Timestamp](#google-protobuf-timestamp) | optional | updated_at - дата и время последнего обновления. |

<details>
<summary>JSON Example</summary>

```json
{
  "createdAt": {
    "nanos": 0,
    "seconds": 0
  },
  "id": "string",
  "isActive": true,
  "model": "string",
  "name": "string",
  "notes": "string",
  "priority": 0,
  "type": "string",
  "updatedAt": {
    "nanos": 0,
    "seconds": 0
  }
}
```

</details>

<a name="emh-v1-listllmprovidersrequest"></a>

### ListLlmProvidersRequest

ListLlmProvidersRequest запрос списка провайдеров.

<a name="emh-v1-listllmprovidersresponse"></a>

### ListLlmProvidersResponse

ListLlmProvidersResponse список провайдеров.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| providers | [LlmProviderInfo](#emh-v1-llmproviderinfo) | repeated | providers - массив настроенных провайдеров. |
| count | int32 | optional | count - общее количество провайдеров в системе. |

<details>
<summary>JSON Example</summary>

```json
{
  "count": 0,
  "providers": [
    {
      "createdAt": {
        "nanos": 0,
        "seconds": 0
      },
      "id": "string",
      "isActive": true,
      "model": "string",
      "name": "string",
      "notes": "string",
      "priority": 0,
      "type": "string",
      "updatedAt": {
        "nanos": 0,
        "seconds": 0
      }
    }
  ]
}
```

</details>

<a name="emh-v1-setactivellmproviderrequest"></a>

### SetActiveLlmProviderRequest

SetActiveLlmProviderRequest запрос на активацию провайдера.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID записи провайдера в БД. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-setactivellmproviderresponse"></a>

### SetActiveLlmProviderResponse

SetActiveLlmProviderResponse результат активации.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| success | bool | optional | success - флаг успешного выполнения операции. |
| provider | [LlmProviderInfo](#emh-v1-llmproviderinfo) | optional | provider - обновлённая информация о провайдере. |

<details>
<summary>JSON Example</summary>

```json
{
  "provider": {
    "createdAt": {
      "nanos": 0,
      "seconds": 0
    },
    "id": "string",
    "isActive": true,
    "model": "string",
    "name": "string",
    "notes": "string",
    "priority": 0,
    "type": "string",
    "updatedAt": {
      "nanos": 0,
      "seconds": 0
    }
  },
  "success": true
}
```

</details>

<a name="emh-v1-testllmproviderrequest"></a>

### TestLlmProviderRequest

TestLlmProviderRequest запрос на тестирование провайдера.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID записи провайдера в БД. |
| test_prompt | string | optional | test_prompt - текст промпта для тестовой генерации. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string",
  "testPrompt": "string"
}
```

</details>

<a name="emh-v1-testllmproviderresponse"></a>

### TestLlmProviderResponse

TestLlmProviderResponse результат тестирования.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| success | bool | optional | success - флаг успешного получения ответа от LLM. |
| raw_response | string | optional | raw_response - сырой текст ответа от модели. |
| processing_time_ms | int64 | optional | processing_time_ms - время выполнения запроса в миллисекундах. |
| error_message | string | optional | error_message - текст ошибки, если success = false. |

<details>
<summary>JSON Example</summary>

```json
{
  "errorMessage": "string",
  "processingTimeMs": 0,
  "rawResponse": "string",
  "success": true
}
```

</details>

<a name="emh-v1-updatellmproviderrequest"></a>

### UpdateLlmProviderRequest

UpdateLlmProviderRequest запрос на обновление метаданных провайдера.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID записи провайдера в БД. |
| priority | int32 | optional | priority - новый приоритет провайдера. |
| notes | string | optional | notes - новые заметки администратора. |
| field_mask | string | repeated | field_mask - список обновляемых полей (priority, notes). |

<details>
<summary>JSON Example</summary>

```json
{
  "fieldMask": [
    "string"
  ],
  "id": "string",
  "notes": "string",
  "priority": 0
}
```

</details>

<a name="emh-v1-updatellmproviderresponse"></a>

### UpdateLlmProviderResponse

UpdateLlmProviderResponse обновлённый провайдер.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| provider | [LlmProviderInfo](#emh-v1-llmproviderinfo) | optional | provider - обновлённая информация о провайдере. |

<details>
<summary>JSON Example</summary>

```json
{
  "provider": {
    "createdAt": {
      "nanos": 0,
      "seconds": 0
    },
    "id": "string",
    "isActive": true,
    "model": "string",
    "name": "string",
    "notes": "string",
    "priority": 0,
    "type": "string",
    "updatedAt": {
      "nanos": 0,
      "seconds": 0
    }
  }
}
```

</details>

