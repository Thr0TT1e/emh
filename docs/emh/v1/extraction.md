# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/extraction.proto](#emh-v1-extraction-proto)
  - **Services**
    - [ExtractionService](#emh-v1-extractionservice)
  - **Messages**
    - [ExtractHeroDataRequest](#emh-v1-extractherodatarequest)
    - [ExtractHeroDataResponse](#emh-v1-extractherodataresponse)
    - [ExtractFromSubmissionRequest](#emh-v1-extractfromsubmissionrequest)
    - [ExtractFromSubmissionResponse](#emh-v1-extractfromsubmissionresponse)
    - [ExtractedHero](#emh-v1-extractedhero)
    - [ExtractedConflict](#emh-v1-extractedconflict)
    - [ExtractedAward](#emh-v1-extractedaward)
    - [ExtractedLocation](#emh-v1-extractedlocation)

<a name="emh-v1-extraction-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/extraction.proto

**Package:** `emh.v1`

<a name="emh-v1-extractionservice"></a>

## ExtractionService

ExtractionService админский сервис извлечения структурированных данных через LLM.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [ExtractHeroData](#emh-v1-extractionservice-extractherodata) | ➡️ Unary | — | ExtractHeroData извлекает данные героя �... |
| [ExtractFromSubmission](#emh-v1-extractionservice-extractfromsubmission) | ➡️ Unary | — | ExtractFromSubmission извлекает данные гер... |

<a name="emh-v1-extractionservice-extractherodata"></a>

### ExtractHeroData

```protobuf
rpc ExtractHeroData([ExtractHeroDataRequest](#emh-v1-extractherodatarequest)) returns ([ExtractHeroDataResponse](#emh-v1-extractherodataresponse))
```

ExtractHeroData извлекает данные героя из текста или URL для проверки редактором.

#### Request Example

```json
{
  "rawText": "string",
  "sourceUrl": "string"
}
```

#### Response Example

```json
{
  "awards": [
    {
      "awardDate": "string",
      "decreeNumber": "string",
      "name": "string"
    }
  ],
  "conflicts": [
    {
      "name": "string",
      "rankAtConflict": "string",
      "specificLocation": "string",
      "suggestedConflictType": "string"
    }
  ],
  "duplicateHeroIds": [
    "string"
  ],
  "extractionLogId": "string",
  "hero": {
    "birthDate": "string",
    "causeOfDeath": "string",
    "deathDate": "string",
    "firstName": "string",
    "fullBio": "string",
    "lastName": "string",
    "memberships": [
      "string"
    ],
    "middleName": "string",
    "nickname": "string",
    "position": "string",
    "rank": "string",
    "serviceBranch": "string",
    "serviceStartDate": "string",
    "shortBio": "string",
    "unit": "string"
  },
  "locations": [
    {
      "heroLocationType": "string",
      "historicalName": "string",
      "locationType": "string",
      "name": "string"
    }
  ],
  "rawLlmJson": "string",
  "sourceUrls": [
    "string"
  ],
  "warnings": [
    "string"
  ]
}
```

---

<a name="emh-v1-extractionservice-extractfromsubmission"></a>

### ExtractFromSubmission

```protobuf
rpc ExtractFromSubmission([ExtractFromSubmissionRequest](#emh-v1-extractfromsubmissionrequest)) returns ([ExtractFromSubmissionResponse](#emh-v1-extractfromsubmissionresponse))
```

ExtractFromSubmission извлекает данные героя из payload_json существующей заявки.

#### Request Example

```json
{
  "submissionId": "string"
}
```

#### Response Example

```json
{
  "awards": [
    {
      "awardDate": "string",
      "decreeNumber": "string",
      "name": "string"
    }
  ],
  "conflicts": [
    {
      "name": "string",
      "rankAtConflict": "string",
      "specificLocation": "string",
      "suggestedConflictType": "string"
    }
  ],
  "extractionLogId": "string",
  "hero": {
    "birthDate": "string",
    "causeOfDeath": "string",
    "deathDate": "string",
    "firstName": "string",
    "fullBio": "string",
    "lastName": "string",
    "memberships": [
      "string"
    ],
    "middleName": "string",
    "nickname": "string",
    "position": "string",
    "rank": "string",
    "serviceBranch": "string",
    "serviceStartDate": "string",
    "shortBio": "string",
    "unit": "string"
  },
  "locations": [
    {
      "heroLocationType": "string",
      "historicalName": "string",
      "locationType": "string",
      "name": "string"
    }
  ],
  "rawLlmJson": "string",
  "warnings": [
    "string"
  ]
}
```

---

<a name="emh-v1-extractherodatarequest"></a>

### ExtractHeroDataRequest

ExtractHeroDataRequest запрос на извлечение данных героя из сырого текста или URL.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| raw_text | string | optional | raw_text - сырой текст статьи, поста или письма. |
| source_url | string | optional | source_url - URL источника (VK, сайт). Текст будет извлечён автоматически. |

<details>
<summary>JSON Example</summary>

```json
{
  "rawText": "string",
  "sourceUrl": "string"
}
```

</details>

<a name="emh-v1-extractherodataresponse"></a>

### ExtractHeroDataResponse

ExtractHeroDataResponse результат извлечения данных героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero | [ExtractedHero](#emh-v1-extractedhero) | optional | hero - извлечённые данные героя. |
| conflicts | [ExtractedConflict](#emh-v1-extractedconflict) | repeated | conflicts - извлечённые конфликты. |
| awards | [ExtractedAward](#emh-v1-extractedaward) | repeated | awards - извлечённые награды. |
| locations | [ExtractedLocation](#emh-v1-extractedlocation) | repeated | locations - извлечённые локации. |
| source_urls | string | repeated | source_urls - URL источников, использованных при извлечении. |
| warnings | string | repeated | warnings - предупреждения о низкой уверенности или неоднозначностях. |
| duplicate_hero_ids | string | repeated | duplicate_hero_ids - возможные дубликаты среди существующих героев. |
| raw_llm_json | string | optional | raw_llm_json - сырой JSON-ответ модели (для аудита и отладки). |
| extraction_log_id | string | optional | extraction_log_id - UUID записи в llm_extraction_logs (для трассировки). |

<details>
<summary>JSON Example</summary>

```json
{
  "awards": [
    {
      "awardDate": "string",
      "decreeNumber": "string",
      "name": "string"
    }
  ],
  "conflicts": [
    {
      "name": "string",
      "rankAtConflict": "string",
      "specificLocation": "string",
      "suggestedConflictType": "string"
    }
  ],
  "duplicateHeroIds": [
    "string"
  ],
  "extractionLogId": "string",
  "hero": {
    "birthDate": "string",
    "causeOfDeath": "string",
    "deathDate": "string",
    "firstName": "string",
    "fullBio": "string",
    "lastName": "string",
    "memberships": [
      "string"
    ],
    "middleName": "string",
    "nickname": "string",
    "position": "string",
    "rank": "string",
    "serviceBranch": "string",
    "serviceStartDate": "string",
    "shortBio": "string",
    "unit": "string"
  },
  "locations": [
    {
      "heroLocationType": "string",
      "historicalName": "string",
      "locationType": "string",
      "name": "string"
    }
  ],
  "rawLlmJson": "string",
  "sourceUrls": [
    "string"
  ],
  "warnings": [
    "string"
  ]
}
```

</details>

<a name="emh-v1-extractfromsubmissionrequest"></a>

### ExtractFromSubmissionRequest

ExtractFromSubmissionRequest запрос на извлечение данных из заявки.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| submission_id | string | optional | submission_id - UUID заявки, из которой извлекаются данные. |

<details>
<summary>JSON Example</summary>

```json
{
  "submissionId": "string"
}
```

</details>

<a name="emh-v1-extractfromsubmissionresponse"></a>

### ExtractFromSubmissionResponse

ExtractFromSubmissionResponse результат извлечения из заявки.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hero | [ExtractedHero](#emh-v1-extractedhero) | optional | hero - извлечённые данные героя. |
| conflicts | [ExtractedConflict](#emh-v1-extractedconflict) | repeated | conflicts - извлечённые конфликты. |
| awards | [ExtractedAward](#emh-v1-extractedaward) | repeated | awards - извлечённые награды. |
| locations | [ExtractedLocation](#emh-v1-extractedlocation) | repeated | locations - извлечённые локации. |
| warnings | string | repeated | warnings - предупреждения о низкой уверенности. |
| raw_llm_json | string | optional | raw_llm_json - сырой JSON-ответ модели. |
| extraction_log_id | string | optional | extraction_log_id - UUID записи в llm_extraction_logs. |

<details>
<summary>JSON Example</summary>

```json
{
  "awards": [
    {
      "awardDate": "string",
      "decreeNumber": "string",
      "name": "string"
    }
  ],
  "conflicts": [
    {
      "name": "string",
      "rankAtConflict": "string",
      "specificLocation": "string",
      "suggestedConflictType": "string"
    }
  ],
  "extractionLogId": "string",
  "hero": {
    "birthDate": "string",
    "causeOfDeath": "string",
    "deathDate": "string",
    "firstName": "string",
    "fullBio": "string",
    "lastName": "string",
    "memberships": [
      "string"
    ],
    "middleName": "string",
    "nickname": "string",
    "position": "string",
    "rank": "string",
    "serviceBranch": "string",
    "serviceStartDate": "string",
    "shortBio": "string",
    "unit": "string"
  },
  "locations": [
    {
      "heroLocationType": "string",
      "historicalName": "string",
      "locationType": "string",
      "name": "string"
    }
  ],
  "rawLlmJson": "string",
  "warnings": [
    "string"
  ]
}
```

</details>

<a name="emh-v1-extractedhero"></a>

### ExtractedHero

ExtractedHero извлечённые данные героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| last_name | string | optional | last_name - фамилия героя. |
| first_name | string | optional | first_name - имя героя. |
| middle_name | string | optional | middle_name - отчество героя. |
| nickname | string | optional | nickname - позывной или прозвище. |
| rank | string | optional | rank - воинское звание. |
| unit | string | optional | unit - подразделение или воинская часть. |
| position | string | optional | position - должность. |
| service_branch | string | optional | service_branch - ведомство или род войск. |
| birth_date | string | optional | birth_date - дата рождения (YYYY-MM-DD, YYYY-MM или YYYY). |
| death_date | string | optional | death_date - дата гибели (YYYY-MM-DD, YYYY-MM или YYYY). |
| cause_of_death | string | optional | cause_of_death - причина или обстоятельства гибели. |
| service_start_date | string | optional | service_start_date - дата начала службы. |
| short_bio | string | optional | short_bio - краткая биография (до 500 символов). |
| full_bio | string | optional | full_bio - полная биография или рассказ о подвиге. |
| memberships | string | repeated | memberships - членство в организациях и объединениях. |

<details>
<summary>JSON Example</summary>

```json
{
  "birthDate": "string",
  "causeOfDeath": "string",
  "deathDate": "string",
  "firstName": "string",
  "fullBio": "string",
  "lastName": "string",
  "memberships": [
    "string"
  ],
  "middleName": "string",
  "nickname": "string",
  "position": "string",
  "rank": "string",
  "serviceBranch": "string",
  "serviceStartDate": "string",
  "shortBio": "string",
  "unit": "string"
}
```

</details>

<a name="emh-v1-extractedconflict"></a>

### ExtractedConflict

ExtractedConflict извлечённый конфликт с привязкой к герою.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string | optional | name - название конфликта. |
| specific_location | string | optional | specific_location - место боя или гибели в рамках конфликта. |
| rank_at_conflict | string | optional | rank_at_conflict - звание героя на момент участия в конфликте. |
| suggested_conflict_type | string | optional | suggested_conflict_type - подсказка типа конфликта для связывания со справочником. |

<details>
<summary>JSON Example</summary>

```json
{
  "name": "string",
  "rankAtConflict": "string",
  "specificLocation": "string",
  "suggestedConflictType": "string"
}
```

</details>

<a name="emh-v1-extractedaward"></a>

### ExtractedAward

ExtractedAward извлечённая награда героя.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string | optional | name - название награды. |
| award_date | string | optional | award_date - дата награждения (YYYY-MM-DD или YYYY). |
| decree_number | string | optional | decree_number - номер приказа или указа о награждении. |

<details>
<summary>JSON Example</summary>

```json
{
  "awardDate": "string",
  "decreeNumber": "string",
  "name": "string"
}
```

</details>

<a name="emh-v1-extractedlocation"></a>

### ExtractedLocation

ExtractedLocation извлечённая локация с типом связи с героем.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string | optional | name - название населённого пункта или объекта. |
| historical_name | string | optional | historical_name - историческое название на период события. |
| location_type | string | optional | location_type - тип географического объекта (country/region/city/village/cemetery). |
| hero_location_type | string | optional | hero_location_type - тип связи с героем (birth/death/burial/residence). |

<details>
<summary>JSON Example</summary>

```json
{
  "heroLocationType": "string",
  "historicalName": "string",
  "locationType": "string",
  "name": "string"
}
```

</details>

