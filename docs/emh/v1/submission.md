# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/submission.proto](#emh-v1-submission-proto)
  - **Services**
    - [SubmissionService](#emh-v1-submissionservice)
    - [SubmissionAdminService](#emh-v1-submissionadminservice)
  - **Messages**
    - [Submission](#emh-v1-submission)
    - [CreateSubmissionRequest](#emh-v1-createsubmissionrequest)
    - [CreateSubmissionResponse](#emh-v1-createsubmissionresponse)
    - [ListSubmissionsRequest](#emh-v1-listsubmissionsrequest)
    - [ListSubmissionsResponse](#emh-v1-listsubmissionsresponse)
    - [GetSubmissionRequest](#emh-v1-getsubmissionrequest)
    - [GetSubmissionResponse](#emh-v1-getsubmissionresponse)
    - [ReviewSubmissionRequest](#emh-v1-reviewsubmissionrequest)
    - [ReviewSubmissionResponse](#emh-v1-reviewsubmissionresponse)
    - [SubmissionReview](#emh-v1-submissionreview)
    - [ListSubmissionReviewsRequest](#emh-v1-listsubmissionreviewsrequest)
    - [ListSubmissionReviewsResponse](#emh-v1-listsubmissionreviewsresponse)
  - **Enums**
    - [SubmissionReviewDecision](#emh-v1-submissionreviewdecision)

<a name="emh-v1-submission-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/submission.proto

**Package:** `emh.v1`

<a name="emh-v1-submissionservice"></a>

## SubmissionService

SubmissionService публичный сервис для создания заявок посетителями сайта.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [CreateSubmission](#emh-v1-submissionservice-createsubmission) | ➡️ Unary | — | CreateSubmission создаёт новую заявку от... |

<a name="emh-v1-submissionservice-createsubmission"></a>

### CreateSubmission

```protobuf
rpc CreateSubmission([CreateSubmissionRequest](#emh-v1-createsubmissionrequest)) returns ([CreateSubmissionResponse](#emh-v1-createsubmissionresponse))
```

CreateSubmission создаёт новую заявку от посетителя.

#### Request Example

```json
{
  "attachmentUrls": [
    "string"
  ],
  "payloadJson": "string",
  "submitterEmail": "string",
  "submitterName": "string",
  "targetHeroId": "string"
}
```

#### Response Example

```json
{
  "submissionId": "string"
}
```

---

<a name="emh-v1-submissionadminservice"></a>

## SubmissionAdminService

SubmissionAdminService защищённый сервис для модерации заявок.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [ListSubmissions](#emh-v1-submissionadminservice-listsubmissions) | ➡️ Unary | — | ListSubmissions возвращает список заяво... |
| [GetSubmission](#emh-v1-submissionadminservice-getsubmission) | ➡️ Unary | — | GetSubmission возвращает заявку по ID. |
| [ReviewSubmission](#emh-v1-submissionadminservice-reviewsubmission) | ➡️ Unary | — | ReviewSubmission одобряет или отклоняет ... |
| [ListSubmissionReviews](#emh-v1-submissionadminservice-listsubmissionreviews) | ➡️ Unary | — | ListSubmissionReviews возвращает историю м... |

<a name="emh-v1-submissionadminservice-listsubmissions"></a>

### ListSubmissions

```protobuf
rpc ListSubmissions([ListSubmissionsRequest](#emh-v1-listsubmissionsrequest)) returns ([ListSubmissionsResponse](#emh-v1-listsubmissionsresponse))
```

ListSubmissions возвращает список заявок с фильтрацией и пагинацией.

#### Request Example

```json
{
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  },
  "status": "PublicationStatus_VALUE"
}
```

#### Response Example

```json
{
  "pagination": {
    "nextCursor": "string",
    "totalCount": 0
  },
  "submissions": [
    {
      "attachmentUrls": [
        "string"
      ],
      "audit": {
        "createdAt": {
          "nanos": 0,
          "seconds": 0
        },
        "createdBy": "string",
        "updatedAt": {
          "nanos": 0,
          "seconds": 0
        }
      },
      "contentHash": "string",
      "id": "string",
      "moderatorComment": "string",
      "payloadJson": "string",
      "status": "PublicationStatus_VALUE",
      "submitterEmail": "string",
      "submitterName": "string",
      "targetHeroId": "string"
    }
  ]
}
```

---

<a name="emh-v1-submissionadminservice-getsubmission"></a>

### GetSubmission

```protobuf
rpc GetSubmission([GetSubmissionRequest](#emh-v1-getsubmissionrequest)) returns ([GetSubmissionResponse](#emh-v1-getsubmissionresponse))
```

GetSubmission возвращает заявку по ID.

#### Request Example

```json
{
  "id": "string"
}
```

#### Response Example

```json
{
  "submission": {
    "attachmentUrls": [
      "string"
    ],
    "audit": {
      "createdAt": {
        "nanos": 0,
        "seconds": 0
      },
      "createdBy": "string",
      "updatedAt": {
        "nanos": 0,
        "seconds": 0
      }
    },
    "contentHash": "string",
    "id": "string",
    "moderatorComment": "string",
    "payloadJson": "string",
    "status": "PublicationStatus_VALUE",
    "submitterEmail": "string",
    "submitterName": "string",
    "targetHeroId": "string"
  }
}
```

---

<a name="emh-v1-submissionadminservice-reviewsubmission"></a>

### ReviewSubmission

```protobuf
rpc ReviewSubmission([ReviewSubmissionRequest](#emh-v1-reviewsubmissionrequest)) returns ([ReviewSubmissionResponse](#emh-v1-reviewsubmissionresponse))
```

ReviewSubmission одобряет или отклоняет заявку с комментарием модератора.
Создаёт запись аудита в submission_reviews.

#### Request Example

```json
{
  "decision": "SubmissionReviewDecision_VALUE",
  "id": "string",
  "moderatorComment": "string"
}
```

#### Response Example

```json
{
  "submission": {
    "attachmentUrls": [
      "string"
    ],
    "audit": {
      "createdAt": {
        "nanos": 0,
        "seconds": 0
      },
      "createdBy": "string",
      "updatedAt": {
        "nanos": 0,
        "seconds": 0
      }
    },
    "contentHash": "string",
    "id": "string",
    "moderatorComment": "string",
    "payloadJson": "string",
    "status": "PublicationStatus_VALUE",
    "submitterEmail": "string",
    "submitterName": "string",
    "targetHeroId": "string"
  }
}
```

---

<a name="emh-v1-submissionadminservice-listsubmissionreviews"></a>

### ListSubmissionReviews

```protobuf
rpc ListSubmissionReviews([ListSubmissionReviewsRequest](#emh-v1-listsubmissionreviewsrequest)) returns ([ListSubmissionReviewsResponse](#emh-v1-listsubmissionreviewsresponse))
```

ListSubmissionReviews возвращает историю модерации заявки (аудит-трейл).

#### Request Example

```json
{
  "submissionId": "string"
}
```

#### Response Example

```json
{
  "reviews": [
    {
      "comment": "string",
      "createdAt": {
        "nanos": 0,
        "seconds": 0
      },
      "decision": "SubmissionReviewDecision_VALUE",
      "id": "string",
      "reviewerName": "string",
      "submissionId": "string"
    }
  ]
}
```

---

<a name="emh-v1-submission"></a>

### Submission

Submission заявка на добавление или исправление данных о герое.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - уникальный идентификатор заявки. |
| submitter_name | string | optional | submitter_name - имя отправителя. |
| submitter_email | string | optional | submitter_email - email отправителя для обратной связи. |
| target_hero_id | string | optional | target_hero_id - ID героя, к которому относится заявка (пусто для нового героя). |
| payload_json | string | optional | payload_json - JSON с данными заявки (гибкая структура). |
| attachment_urls | string | repeated | attachment_urls - URL прикреплённых документов или фото. |
| status | PublicationStatus | optional | status - статус обработки (DRAFT=на модерации, PUBLISHED=одобрена, ARCHIVED=отклонена). |
| moderator_comment | string | optional | moderator_comment - комментарий модератора по результату рассмотрения. |
| audit | [AuditInfo](#emh-v1-auditinfo) | optional | audit - метаданные создания и обновления. |
| content_hash | string | optional | content_hash - SHA-256 хеш нормализованного payload_json (read-only). |

<details>
<summary>JSON Example</summary>

```json
{
  "attachmentUrls": [
    "string"
  ],
  "audit": {
    "createdAt": {
      "nanos": 0,
      "seconds": 0
    },
    "createdBy": "string",
    "updatedAt": {
      "nanos": 0,
      "seconds": 0
    }
  },
  "contentHash": "string",
  "id": "string",
  "moderatorComment": "string",
  "payloadJson": "string",
  "status": "PublicationStatus_VALUE",
  "submitterEmail": "string",
  "submitterName": "string",
  "targetHeroId": "string"
}
```

</details>

<a name="emh-v1-createsubmissionrequest"></a>

### CreateSubmissionRequest

CreateSubmissionRequest данные для создания заявки посетителем.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| submitter_name | string | optional | submitter_name - имя отправителя (обязательно). |
| submitter_email | string | optional | submitter_email - контактный email отправителя. |
| target_hero_id | string | optional | target_hero_id - ID существующего героя (пусто для заявки на нового героя). |
| payload_json | string | optional | payload_json - данные заявки в формате JSON. |
| attachment_urls | string | repeated | attachment_urls - ссылки на прикреплённые файлы. |

<details>
<summary>JSON Example</summary>

```json
{
  "attachmentUrls": [
    "string"
  ],
  "payloadJson": "string",
  "submitterEmail": "string",
  "submitterName": "string",
  "targetHeroId": "string"
}
```

</details>

<a name="emh-v1-createsubmissionresponse"></a>

### CreateSubmissionResponse

CreateSubmissionResponse результат создания заявки.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| submission_id | string | optional | submission_id - UUID созданной заявки. |

<details>
<summary>JSON Example</summary>

```json
{
  "submissionId": "string"
}
```

</details>

<a name="emh-v1-listsubmissionsrequest"></a>

### ListSubmissionsRequest

ListSubmissionsRequest параметры выборки заявок для модерации.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pagination | [PaginationRequest](#emh-v1-paginationrequest) | optional | pagination - параметры курсорной пагинации. |
| status | PublicationStatus | optional | status - фильтр по статусу заявки. |

<details>
<summary>JSON Example</summary>

```json
{
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  },
  "status": "PublicationStatus_VALUE"
}
```

</details>

<a name="emh-v1-listsubmissionsresponse"></a>

### ListSubmissionsResponse

ListSubmissionsResponse список заявок для модератора.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| submissions | [Submission](#emh-v1-submission) | repeated | submissions - массив заявок. |
| pagination | [PaginationResponse](#emh-v1-paginationresponse) | optional | pagination - метаданные пагинации. |

<details>
<summary>JSON Example</summary>

```json
{
  "pagination": {
    "nextCursor": "string",
    "totalCount": 0
  },
  "submissions": [
    {
      "attachmentUrls": [
        "string"
      ],
      "audit": {
        "createdAt": {
          "nanos": 0,
          "seconds": 0
        },
        "createdBy": "string",
        "updatedAt": {
          "nanos": 0,
          "seconds": 0
        }
      },
      "contentHash": "string",
      "id": "string",
      "moderatorComment": "string",
      "payloadJson": "string",
      "status": "PublicationStatus_VALUE",
      "submitterEmail": "string",
      "submitterName": "string",
      "targetHeroId": "string"
    }
  ]
}
```

</details>

<a name="emh-v1-getsubmissionrequest"></a>

### GetSubmissionRequest

GetSubmissionRequest запрос на получение заявки по ID.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID заявки. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-getsubmissionresponse"></a>

### GetSubmissionResponse

GetSubmissionResponse ответ с заявкой.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| submission | [Submission](#emh-v1-submission) | optional | submission - объект заявки. |

<details>
<summary>JSON Example</summary>

```json
{
  "submission": {
    "attachmentUrls": [
      "string"
    ],
    "audit": {
      "createdAt": {
        "nanos": 0,
        "seconds": 0
      },
      "createdBy": "string",
      "updatedAt": {
        "nanos": 0,
        "seconds": 0
      }
    },
    "contentHash": "string",
    "id": "string",
    "moderatorComment": "string",
    "payloadJson": "string",
    "status": "PublicationStatus_VALUE",
    "submitterEmail": "string",
    "submitterName": "string",
    "targetHeroId": "string"
  }
}
```

</details>

<a name="emh-v1-reviewsubmissionrequest"></a>

### ReviewSubmissionRequest

ReviewSubmissionRequest запрос на модерацию заявки.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID заявки. |
| decision | [SubmissionReviewDecision](#emh-v1-submissionreviewdecision) | optional | decision - решение модератора (одобрить/отклонить). |
| moderator_comment | string | optional | moderator_comment - комментарий модератора. |

<details>
<summary>JSON Example</summary>

```json
{
  "decision": "SubmissionReviewDecision_VALUE",
  "id": "string",
  "moderatorComment": "string"
}
```

</details>

<a name="emh-v1-reviewsubmissionresponse"></a>

### ReviewSubmissionResponse

ReviewSubmissionResponse заявка после модерации.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| submission | [Submission](#emh-v1-submission) | optional | submission - обновлённый объект заявки. |

<details>
<summary>JSON Example</summary>

```json
{
  "submission": {
    "attachmentUrls": [
      "string"
    ],
    "audit": {
      "createdAt": {
        "nanos": 0,
        "seconds": 0
      },
      "createdBy": "string",
      "updatedAt": {
        "nanos": 0,
        "seconds": 0
      }
    },
    "contentHash": "string",
    "id": "string",
    "moderatorComment": "string",
    "payloadJson": "string",
    "status": "PublicationStatus_VALUE",
    "submitterEmail": "string",
    "submitterName": "string",
    "targetHeroId": "string"
  }
}
```

</details>

<a name="emh-v1-submissionreview"></a>

### SubmissionReview

SubmissionReview запись аудита модерации заявки.
Фиксирует кто, когда и какое решение принял по конкретной заявке.
История неизменяема: записи только добавляются, не редактируются.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - уникальный идентификатор записи аудита (UUID). |
| submission_id | string | optional | submission_id - UUID заявки, к которой относится решение. |
| reviewer_name | string | optional | reviewer_name - имя модератора, принявшего решение (извлекается из JWT). |
| decision | [SubmissionReviewDecision](#emh-v1-submissionreviewdecision) | optional | decision - решение модератора (одобрить/отклонить). |
| comment | string | optional | comment - комментарий модератора к решению (причина отклонения и т.д.). |
| created_at | [Timestamp](#google-protobuf-timestamp) | optional | created_at - время принятия решения. |

<details>
<summary>JSON Example</summary>

```json
{
  "comment": "string",
  "createdAt": {
    "nanos": 0,
    "seconds": 0
  },
  "decision": "SubmissionReviewDecision_VALUE",
  "id": "string",
  "reviewerName": "string",
  "submissionId": "string"
}
```

</details>

<a name="emh-v1-listsubmissionreviewsrequest"></a>

### ListSubmissionReviewsRequest

ListSubmissionReviewsRequest запрос на получение истории модерации заявки.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| submission_id | string | optional | submission_id - UUID заявки, для которой запрашивается история. |

<details>
<summary>JSON Example</summary>

```json
{
  "submissionId": "string"
}
```

</details>

<a name="emh-v1-listsubmissionreviewsresponse"></a>

### ListSubmissionReviewsResponse

ListSubmissionReviewsResponse история модерации заявки в обратном
хронологическом порядке (новые записи первыми).

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| reviews | [SubmissionReview](#emh-v1-submissionreview) | repeated | reviews - записи аудита модерации. |

<details>
<summary>JSON Example</summary>

```json
{
  "reviews": [
    {
      "comment": "string",
      "createdAt": {
        "nanos": 0,
        "seconds": 0
      },
      "decision": "SubmissionReviewDecision_VALUE",
      "id": "string",
      "reviewerName": "string",
      "submissionId": "string"
    }
  ]
}
```

</details>

<a name="emh-v1-submissionreviewdecision"></a>

### SubmissionReviewDecision

SubmissionReviewDecision решение модератора по заявке.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `SUBMISSION_REVIEW_DECISION_UNSPECIFIED` | 0 | SUBMISSION_REVIEW_DECISION_UNSPECIFIED - решение не задано. |
| `SUBMISSION_REVIEW_DECISION_APPROVE` | 1 | SUBMISSION_REVIEW_DECISION_APPROVE - одобрить заявку. |
| `SUBMISSION_REVIEW_DECISION_REJECT` | 2 | SUBMISSION_REVIEW_DECISION_REJECT - отклонить заявку. |

