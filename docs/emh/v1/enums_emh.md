# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/enums_emh.proto](#emh-v1-enums-emh-proto)
  - **Enums**
    - [PublicationStatus](#emh-v1-publicationstatus)
    - [LocationType](#emh-v1-locationtype)
    - [HeroLocationType](#emh-v1-herolocationtype)
    - [ConflictType](#emh-v1-conflicttype)
    - [DatePrecision](#emh-v1-dateprecision)
    - [AwardType](#emh-v1-awardtype)
    - [AwardJurisdiction](#emh-v1-awardjurisdiction)

<a name="emh-v1-enums-emh-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/enums_emh.proto

**Package:** `emh.v1`

<a name="emh-v1-publicationstatus"></a>

### PublicationStatus

PublicationStatus определяет видимость записи на сайте и в админке.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `PUBLICATION_STATUS_UNSPECIFIED` | 0 | PUBLICATION_STATUS_UNSPECIFIED - статус не задан (ошибка валидации). |
| `PUBLICATION_STATUS_DRAFT` | 1 | PUBLICATION_STATUS_DRAFT - черновик или запись на модерации. |
| `PUBLICATION_STATUS_PUBLISHED` | 2 | PUBLICATION_STATUS_PUBLISHED - запись опубликована и доступна всем. |
| `PUBLICATION_STATUS_ARCHIVED` | 3 | PUBLICATION_STATUS_ARCHIVED - запись скрыта с сайта, но сохранена в БД. |

<a name="emh-v1-locationtype"></a>

### LocationType

LocationType классифицирует географические объекты.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `LOCATION_TYPE_UNSPECIFIED` | 0 | LOCATION_TYPE_UNSPECIFIED - тип не задан. |
| `LOCATION_TYPE_COUNTRY` | 1 | LOCATION_TYPE_COUNTRY - страна. |
| `LOCATION_TYPE_REGION` | 2 | LOCATION_TYPE_REGION - область, край, республика. |
| `LOCATION_TYPE_CITY` | 3 | LOCATION_TYPE_CITY - город. |
| `LOCATION_TYPE_VILLAGE` | 4 | LOCATION_TYPE_VILLAGE - село, деревня, поселок. |
| `LOCATION_TYPE_CEMETERY` | 5 | LOCATION_TYPE_CEMETERY - кладбище или мемориальный комплекс. |

<a name="emh-v1-herolocationtype"></a>

### HeroLocationType

HeroLocationType определяет роль локации в биографии героя.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `HERO_LOCATION_TYPE_UNSPECIFIED` | 0 | HERO_LOCATION_TYPE_UNSPECIFIED - тип связи не задан. |
| `HERO_LOCATION_TYPE_BIRTH` | 1 | HERO_LOCATION_TYPE_BIRTH - место рождения. |
| `HERO_LOCATION_TYPE_DEATH` | 2 | HERO_LOCATION_TYPE_DEATH - место гибели. |
| `HERO_LOCATION_TYPE_BURIAL` | 3 | HERO_LOCATION_TYPE_BURIAL - место захоронения. |
| `HERO_LOCATION_TYPE_RESIDENCE` | 4 | HERO_LOCATION_TYPE_RESIDENCE - место проживания до конфликта. |

<a name="emh-v1-conflicttype"></a>

### ConflictType

ConflictType классифицирует военные конфликты.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `CONFLICT_TYPE_UNSPECIFIED` | 0 | CONFLICT_TYPE_UNSPECIFIED - тип конфликта не задан. |
| `CONFLICT_TYPE_GLOBAL` | 1 | CONFLICT_TYPE_GLOBAL - мировые войны. |
| `CONFLICT_TYPE_LOCAL` | 2 | CONFLICT_TYPE_LOCAL - локальные военные конфликты. |
| `CONFLICT_TYPE_PEACEKEEPING` | 3 | CONFLICT_TYPE_PEACEKEEPING - миротворческие операции. |
| `CONFLICT_TYPE_COUNTER_TERRORISM` | 4 | CONFLICT_TYPE_COUNTER_TERRORISM - контртеррористические операции (КТО). |
| `CONFLICT_TYPE_SPECIAL_OPERATION` | 5 | CONFLICT_TYPE_SPECIAL_OPERATION - специальная военная операция (СВО) и подобные. |

<a name="emh-v1-dateprecision"></a>

### DatePrecision

DatePrecision определяет уровень точности гибкой даты.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `DATE_PRECISION_UNSPECIFIED` | 0 | DATE_PRECISION_UNSPECIFIED означает, что точность даты не указана. |
| `DATE_PRECISION_EXACT` | 1 | DATE_PRECISION_EXACT означает точную дату формата YYYY-MM-DD. |
| `DATE_PRECISION_MONTH` | 2 | DATE_PRECISION_MONTH означает, что известны только год и месяц. |
| `DATE_PRECISION_YEAR` | 3 | DATE_PRECISION_YEAR означает, что известен только год. |
| `DATE_PRECISION_SEASON` | 4 | DATE_PRECISION_SEASON означает сезон и год, например "Лето 1989". |
| `DATE_PRECISION_DAY_MONTH` | 5 | DATE_PRECISION_DAY_MONTH означает день и месяц без года, например "28 июля". |
| `DATE_PRECISION_RANGE` | 6 | DATE_PRECISION_RANGE зарезервировано для диапазона дат, например "1994-1995". |
| `DATE_PRECISION_UNKNOWN` | 7 | DATE_PRECISION_UNKNOWN означает, что дата неизвестна и есть только текстовое описание. |

<a name="emh-v1-awardtype"></a>

### AwardType

AwardType определяет тип награды. Используется для старшинства в орденской планке.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `AWARD_TYPE_UNSPECIFIED` | 0 | AWARD_TYPE_UNSPECIFIED - тип не указан. |
| `AWARD_TYPE_ORDER` | 1 | AWARD_TYPE_ORDER - орден. |
| `AWARD_TYPE_MEDAL` | 2 | AWARD_TYPE_MEDAL - медаль. |
| `AWARD_TYPE_BADGE` | 3 | AWARD_TYPE_BADGE - знак отличия или почётный знак. |

<a name="emh-v1-awardjurisdiction"></a>

### AwardJurisdiction

AwardJurisdiction определяет государственную принадлежность награды.
Задаёт первичную ось старшинства согласно приказу МО РФ №1500.

| Name | Number | Description |
| ---- | ------ | ----------- |
| `AWARD_JURISDICTION_UNSPECIFIED` | 0 | AWARD_JURISDICTION_UNSPECIFIED - принадлежность не указана. |
| `AWARD_JURISDICTION_RUSSIAN_FEDERATION` | 1 | AWARD_JURISDICTION_RUSSIAN_FEDERATION - награда Российской Федерации. |
| `AWARD_JURISDICTION_USSR` | 2 | AWARD_JURISDICTION_USSR - награда СССР. |
| `AWARD_JURISDICTION_DEPARTMENTAL` | 3 | AWARD_JURISDICTION_DEPARTMENTAL - ведомственная награда. |

