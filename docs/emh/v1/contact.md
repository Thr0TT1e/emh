# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/contact.proto](#emh-v1-contact-proto)
  - **Services**
    - [ContactService](#emh-v1-contactservice)
  - **Messages**
    - [CreateContactMessageRequest](#emh-v1-createcontactmessagerequest)
    - [CreateContactMessageResponse](#emh-v1-createcontactmessageresponse)

<a name="emh-v1-contact-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/contact.proto

**Package:** `emh.v1`

<a name="emh-v1-contactservice"></a>

## ContactService

ContactService — публичный сервис обратной связи.
Endpoint: POST /emh.v1.ContactService/CreateContactMessage

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [CreateContactMessage](#emh-v1-contactservice-createcontactmessage) | ➡️ Unary | — | CreateContactMessage принимает сообщение �... |

<a name="emh-v1-contactservice-createcontactmessage"></a>

### CreateContactMessage

```protobuf
rpc CreateContactMessage([CreateContactMessageRequest](#emh-v1-createcontactmessagerequest)) returns ([CreateContactMessageResponse](#emh-v1-createcontactmessageresponse))
```

CreateContactMessage принимает сообщение формы обратной связи,
сохраняет его в БД и (если SMTP включён и это не honeypot)
инициирует отправку письма владельцу проекта.

#### Request Example

```json
{
  "consent": true,
  "email": "string",
  "honeypot": "string",
  "message": "string",
  "name": "string",
  "pageUrl": "string",
  "subject": "string"
}
```

#### Response Example

```json
{
  "id": "string"
}
```

---

<a name="emh-v1-createcontactmessagerequest"></a>

### CreateContactMessageRequest

CreateContactMessageRequest — запрос на создание сообщения обратной связи.
Форма НЕ предназначена для отправки данных о героях (для этого есть SubmissionService).

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string | optional | name — имя отправителя (обязательно, 2–200 символов). |
| email | string | optional | email — email отправителя для ответа. |
| subject | string | optional | subject — тема обращения (опционально, до 200 символов). |
| message | string | optional | message — текст сообщения (обязательно, 10–5000 символов). |
| page_url | string | optional | page_url — страница, с которой была отправлена форма (для контекста). |
| honeypot | string | optional | honeypot — скрытое поле для простых ботов. Должно быть пустым. Если заполнено — запрос принимается, но email не отправляется. |
| consent | bool | optional | consent — согласие на обработку персональных данных. Обязательно true. |

<details>
<summary>JSON Example</summary>

```json
{
  "consent": true,
  "email": "string",
  "honeypot": "string",
  "message": "string",
  "name": "string",
  "pageUrl": "string",
  "subject": "string"
}
```

</details>

<a name="emh-v1-createcontactmessageresponse"></a>

### CreateContactMessageResponse

CreateContactMessageResponse — результат создания сообщения.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id — UUID созданного сообщения в таблице contact_messages. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

