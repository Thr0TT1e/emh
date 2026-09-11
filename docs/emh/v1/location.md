# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/location.proto](#emh-v1-location-proto)
  - **Services**
    - [LocationService](#emh-v1-locationservice)
  - **Messages**
    - [Location](#emh-v1-location)
    - [GetLocationRequest](#emh-v1-getlocationrequest)
    - [GetLocationResponse](#emh-v1-getlocationresponse)
    - [ListLocationsRequest](#emh-v1-listlocationsrequest)
    - [ListLocationsResponse](#emh-v1-listlocationsresponse)

<a name="emh-v1-location-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/location.proto

**Package:** `emh.v1`

<a name="emh-v1-locationservice"></a>

## LocationService

LocationService публичный сервис для работы со справочником мест.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [GetLocation](#emh-v1-locationservice-getlocation) | ➡️ Unary | — | GetLocation возвращает локацию по ID. |
| [ListLocations](#emh-v1-locationservice-listlocations) | ➡️ Unary | — | ListLocations возвращает отфильтрован�... |

<a name="emh-v1-locationservice-getlocation"></a>

### GetLocation

```protobuf
rpc GetLocation([GetLocationRequest](#emh-v1-getlocationrequest)) returns ([GetLocationResponse](#emh-v1-getlocationresponse))
```

GetLocation возвращает локацию по ID.

#### Request Example

```json
{
  "id": "string"
}
```

#### Response Example

```json
{
  "location": {
    "historicalName": "string",
    "id": "string",
    "latitude": 0,
    "longitude": 0,
    "name": "string",
    "parentId": "string",
    "type": "LocationType_VALUE"
  }
}
```

---

<a name="emh-v1-locationservice-listlocations"></a>

### ListLocations

```protobuf
rpc ListLocations([ListLocationsRequest](#emh-v1-listlocationsrequest)) returns ([ListLocationsResponse](#emh-v1-listlocationsresponse))
```

ListLocations возвращает отфильтрованный список локаций.

#### Request Example

```json
{
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  },
  "parentId": "string",
  "searchQuery": "string",
  "type": "LocationType_VALUE"
}
```

#### Response Example

```json
{
  "locations": [
    {
      "historicalName": "string",
      "id": "string",
      "latitude": 0,
      "longitude": 0,
      "name": "string",
      "parentId": "string",
      "type": "LocationType_VALUE"
    }
  ],
  "pagination": {
    "nextCursor": "string",
    "totalCount": 0
  }
}
```

---

<a name="emh-v1-location"></a>

### Location

Location представляет географический объект в справочнике мест.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - уникальный идентификатор локации (UUID). |
| name | string | optional | name - современное название населенного пункта или объекта. |
| historical_name | string | optional | historical_name - историческое название на период события (опционально). |
| type | LocationType | optional | type - тип географического объекта. |
| parent_id | string | optional | parent_id - ID родительской локации (региона или страны). |
| latitude | double | optional | latitude - широта для отображения на карте. |
| longitude | double | optional | longitude - долгота для отображения на карте. |

<details>
<summary>JSON Example</summary>

```json
{
  "historicalName": "string",
  "id": "string",
  "latitude": 0,
  "longitude": 0,
  "name": "string",
  "parentId": "string",
  "type": "LocationType_VALUE"
}
```

</details>

<a name="emh-v1-getlocationrequest"></a>

### GetLocationRequest

GetLocationRequest запрос на получение локации по ID.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID локации. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-getlocationresponse"></a>

### GetLocationResponse

GetLocationResponse ответ с данными локации.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| location | [Location](#emh-v1-location) | optional | location - объект локации. |

<details>
<summary>JSON Example</summary>

```json
{
  "location": {
    "historicalName": "string",
    "id": "string",
    "latitude": 0,
    "longitude": 0,
    "name": "string",
    "parentId": "string",
    "type": "LocationType_VALUE"
  }
}
```

</details>

<a name="emh-v1-listlocationsrequest"></a>

### ListLocationsRequest

ListLocationsRequest параметры выборки локаций.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| parent_id | string | optional | parent_id - фильтр по родительской локации (например, города области). |
| search_query | string | optional | search_query - текстовый поиск по названию. |
| type | LocationType | optional | type - фильтр по типу локации. |
| pagination | [PaginationRequest](#emh-v1-paginationrequest) | optional | pagination - параметры курсорной пагинации. |

<details>
<summary>JSON Example</summary>

```json
{
  "pagination": {
    "cursor": "string",
    "pageSize": 0
  },
  "parentId": "string",
  "searchQuery": "string",
  "type": "LocationType_VALUE"
}
```

</details>

<a name="emh-v1-listlocationsresponse"></a>

### ListLocationsResponse

ListLocationsResponse список найденных локаций.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| locations | [Location](#emh-v1-location) | repeated | locations - массив объектов локаций. |
| pagination | [PaginationResponse](#emh-v1-paginationresponse) | optional | pagination - метаданные для перехода к следующей странице. |

<details>
<summary>JSON Example</summary>

```json
{
  "locations": [
    {
      "historicalName": "string",
      "id": "string",
      "latitude": 0,
      "longitude": 0,
      "name": "string",
      "parentId": "string",
      "type": "LocationType_VALUE"
    }
  ],
  "pagination": {
    "nextCursor": "string",
    "totalCount": 0
  }
}
```

</details>

