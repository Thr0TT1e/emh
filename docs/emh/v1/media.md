# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/media.proto](#emh-v1-media-proto)
  - **Services**
    - [MediaService](#emh-v1-mediaservice)
  - **Messages**
    - [GetUploadUrlRequest](#emh-v1-getuploadurlrequest)
    - [GetUploadUrlResponse](#emh-v1-getuploadurlresponse)
    - [BatchGetUploadUrlsRequest](#emh-v1-batchgetuploadurlsrequest)
    - [BatchGetUploadUrlsResponse](#emh-v1-batchgetuploadurlsresponse)
    - [FileToUpload](#emh-v1-filetoupload)
    - [UploadUrlInfo](#emh-v1-uploadurlinfo)
  - **Enums**
    - [UploadType](#emh-v1-uploadtype)

<a name="emh-v1-media-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/media.proto

**Package:** `emh.v1`

<a name="emh-v1-mediaservice"></a>

## MediaService

MediaService сервис для работы с медиа-контентом.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [GetUploadUrl](#emh-v1-mediaservice-getuploadurl) | ➡️ Unary | — | GetUploadUrl генерирует presigned URL для бе�... |
| [BatchGetUploadUrls](#emh-v1-mediaservice-batchgetuploadurls) | ➡️ Unary | — | BatchGetUploadUrls генерирует несколько p... |

<a name="emh-v1-mediaservice-getuploadurl"></a>

### GetUploadUrl

```protobuf
rpc GetUploadUrl([GetUploadUrlRequest](#emh-v1-getuploadurlrequest)) returns ([GetUploadUrlResponse](#emh-v1-getuploadurlresponse))
```

GetUploadUrl генерирует presigned URL для безопасной прямой загрузки в хранилище.

#### Request Example

```json
{
  "contentType": "string",
  "filename": "string",
  "type": "UploadType_VALUE"
}
```

#### Response Example

```json
{
  "expiresAt": "string",
  "publicUrl": "string",
  "uploadUrl": "string"
}
```

---

<a name="emh-v1-mediaservice-batchgetuploadurls"></a>

### BatchGetUploadUrls

```protobuf
rpc BatchGetUploadUrls([BatchGetUploadUrlsRequest](#emh-v1-batchgetuploadurlsrequest)) returns ([BatchGetUploadUrlsResponse](#emh-v1-batchgetuploadurlsresponse))
```

BatchGetUploadUrls генерирует несколько presigned URL за один запрос.

#### Request Example

```json
{
  "files": [
    {
      "contentType": "string",
      "filename": "string"
    }
  ],
  "type": "UploadType_VALUE"
}
```

#### Response Example

```json
{
  "count": 0,
  "urls": [
    {
      "expiresAt": "string",
      "filename": "string",
      "publicUrl": "string",
      "uploadUrl": "string"
    }
  ]
}
```

---

<a name="emh-v1-getuploadurlrequest"></a>

### GetUploadUrlRequest

GetUploadUrlRequest запрос на получение presigned URL для загрузки.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| type | [UploadType](#emh-v1-uploadtype) | optional | type - тип загружаемого контента. |
| filename | string | optional | filename - оригинальное имя файла. |
| content_type | string | optional | content_type - MIME-тип файла. |

<details>
<summary>JSON Example</summary>

```json
{
  "contentType": "string",
  "filename": "string",
  "type": "UploadType_VALUE"
}
```

</details>

<a name="emh-v1-getuploadurlresponse"></a>

### GetUploadUrlResponse

GetUploadUrlResponse содержит URL для прямой загрузки и будущий публичный путь.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| upload_url | string | optional | upload_url - presigned PUT URL для загрузки файла в S3/MinIO. |
| public_url | string | optional | public_url - финальный публичный URL файла после успешной загрузки. |
| expires_at | string | optional | expires_at - время истечения действия presigned URL (RFC3339). |

<details>
<summary>JSON Example</summary>

```json
{
  "expiresAt": "string",
  "publicUrl": "string",
  "uploadUrl": "string"
}
```

</details>

<a name="emh-v1-batchgetuploadurlsrequest"></a>

### BatchGetUploadUrlsRequest

BatchGetUploadUrlsRequest Batch-запрос для пакетной генерации presigned URL

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| files | [FileToUpload](#emh-v1-filetoupload) | repeated | files - список файлов для загрузки (не более 50 за один запрос). |
| type | [UploadType](#emh-v1-uploadtype) | optional | type - тип загружаемого контента (одинаковый для всех файлов в пакете). |

<details>
<summary>JSON Example</summary>

```json
{
  "files": [
    {
      "contentType": "string",
      "filename": "string"
    }
  ],
  "type": "UploadType_VALUE"
}
```

</details>

<a name="emh-v1-batchgetuploadurlsresponse"></a>

### BatchGetUploadUrlsResponse

BatchGetUploadUrlsResponse Batch-ответ с массивом presigned URL

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| urls | [UploadUrlInfo](#emh-v1-uploadurlinfo) | repeated | urls - массив presigned URL для каждого файла. |
| count | int32 | optional | count - количество сгенерированных URL. |

<details>
<summary>JSON Example</summary>

```json
{
  "count": 0,
  "urls": [
    {
      "expiresAt": "string",
      "filename": "string",
      "publicUrl": "string",
      "uploadUrl": "string"
    }
  ]
}
```

</details>

<a name="emh-v1-filetoupload"></a>

### FileToUpload

FileToUpload Описание одного файла для batch-загрузки

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filename | string | optional | filename - оригинальное имя файла. |
| content_type | string | optional | content_type - MIME-тип файла. |

<details>
<summary>JSON Example</summary>

```json
{
  "contentType": "string",
  "filename": "string"
}
```

</details>

<a name="emh-v1-uploadurlinfo"></a>

### UploadUrlInfo

UploadUrlInfo Элемент batch-ответа

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| filename | string | optional | filename - имя файла (для маппинга на фронтенде). |
| upload_url | string | optional | upload_url - presigned PUT URL для загрузки файла в S3/MinIO. |
| public_url | string | optional | public_url - финальный публичный URL файла после успешной загрузки. |
| expires_at | string | optional | expires_at - время истечения действия presigned URL (RFC3339). |

<details>
<summary>JSON Example</summary>

```json
{
  "expiresAt": "string",
  "filename": "string",
  "publicUrl": "string",
  "uploadUrl": "string"
}
```

</details>

<a name="emh-v1-uploadtype"></a>

### UploadType

UploadType определяет назначение загружаемого файла.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `UPLOAD_TYPE_UNSPECIFIED` | 0 | UPLOAD_TYPE_UNSPECIFIED - тип не задан. |
| `UPLOAD_TYPE_HERO_PHOTO` | 1 | UPLOAD_TYPE_HERO_PHOTO - фотография героя. |
| `UPLOAD_TYPE_AWARD_IMAGE` | 2 | UPLOAD_TYPE_AWARD_IMAGE - изображение награды. |
| `UPLOAD_TYPE_SUBMISSION_ATTACHMENT` | 3 | UPLOAD_TYPE_SUBMISSION_ATTACHMENT - документ к заявке. |

