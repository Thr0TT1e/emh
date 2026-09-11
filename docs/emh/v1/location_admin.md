# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/location_admin.proto](#emh-v1-location-admin-proto)
  - **Services**
    - [LocationAdminService](#emh-v1-locationadminservice)
  - **Messages**
    - [CreateLocationRequest](#emh-v1-createlocationrequest)
    - [CreateLocationResponse](#emh-v1-createlocationresponse)
    - [UpdateLocationRequest](#emh-v1-updatelocationrequest)
    - [UpdateLocationResponse](#emh-v1-updatelocationresponse)
    - [DeleteLocationRequest](#emh-v1-deletelocationrequest)
    - [DeleteLocationResponse](#emh-v1-deletelocationresponse)
    - [ReparentLocationsRequest](#emh-v1-reparentlocationsrequest)
    - [ReparentLocationsResponse](#emh-v1-reparentlocationsresponse)

<a name="emh-v1-location-admin-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/location_admin.proto

**Package:** `emh.v1`

<a name="emh-v1-locationadminservice"></a>

## LocationAdminService

LocationAdminService защищенный сервис для управления справочником мест.

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [CreateLocation](#emh-v1-locationadminservice-createlocation) | ➡️ Unary | — | CreateLocation создает новую запись в с�... |
| [UpdateLocation](#emh-v1-locationadminservice-updatelocation) | ➡️ Unary | — | UpdateLocation частично обновляет суще�... |
| [DeleteLocation](#emh-v1-locationadminservice-deletelocation) | ➡️ Unary | — | DeleteLocation удаляет локацию (требует... |
| [ReparentLocations](#emh-v1-locationadminservice-reparentlocations) | ➡️ Unary | — | ReparentLocations массово переносит доче... |

<a name="emh-v1-locationadminservice-createlocation"></a>

### CreateLocation

```protobuf
rpc CreateLocation([CreateLocationRequest](#emh-v1-createlocationrequest)) returns ([CreateLocationResponse](#emh-v1-createlocationresponse))
```

CreateLocation создает новую запись в справочнике мест.

#### Request Example

```json
{
  "historicalName": "string",
  "latitude": 0,
  "longitude": 0,
  "name": "string",
  "parentId": "string",
  "type": "LocationType_VALUE"
}
```

#### Response Example

```json
{
  "id": "string"
}
```

---

<a name="emh-v1-locationadminservice-updatelocation"></a>

### UpdateLocation

```protobuf
rpc UpdateLocation([UpdateLocationRequest](#emh-v1-updatelocationrequest)) returns ([UpdateLocationResponse](#emh-v1-updatelocationresponse))
```

UpdateLocation частично обновляет существующую локацию.

#### Request Example

```json
{
  "fieldMask": [
    "string"
  ],
  "historicalName": "string",
  "id": "string",
  "latitude": 0,
  "longitude": 0,
  "name": "string",
  "parentId": "string",
  "type": "LocationType_VALUE"
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

<a name="emh-v1-locationadminservice-deletelocation"></a>

### DeleteLocation

```protobuf
rpc DeleteLocation([DeleteLocationRequest](#emh-v1-deletelocationrequest)) returns ([DeleteLocationResponse](#emh-v1-deletelocationresponse))
```

DeleteLocation удаляет локацию (требует предварительной репривязки детей).

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

<a name="emh-v1-locationadminservice-reparentlocations"></a>

### ReparentLocations

```protobuf
rpc ReparentLocations([ReparentLocationsRequest](#emh-v1-reparentlocationsrequest)) returns ([ReparentLocationsResponse](#emh-v1-reparentlocationsresponse))
```

ReparentLocations массово переносит дочерние локации к новому родителю.

#### Request Example

```json
{
  "newParentId": "string",
  "oldParentId": "string"
}
```

#### Response Example

```json
{
  "affectedCount": 0
}
```

---

<a name="emh-v1-createlocationrequest"></a>

### CreateLocationRequest

CreateLocationRequest данные для создания новой локации.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string | optional | name - современное название (обязательно). |
| historical_name | string | optional | historical_name - историческое название (опционально). |
| type | LocationType | optional | type - тип географического объекта. |
| parent_id | string | optional | parent_id - ID родительской локации. |
| latitude | double | optional | latitude - широта центра объекта (-90..90). |
| longitude | double | optional | longitude - долгота центра объекта (-180..180). |

<details>
<summary>JSON Example</summary>

```json
{
  "historicalName": "string",
  "latitude": 0,
  "longitude": 0,
  "name": "string",
  "parentId": "string",
  "type": "LocationType_VALUE"
}
```

</details>

<a name="emh-v1-createlocationresponse"></a>

### CreateLocationResponse

CreateLocationResponse результат создания локации.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID созданной локации. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-updatelocationrequest"></a>

### UpdateLocationRequest

UpdateLocationRequest данные для частичного обновления локации.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID обновляемой локации. |
| name | string | optional | name - новое современное название. |
| historical_name | string | optional | historical_name - новое историческое название. |
| type | LocationType | optional | type - новый тип локации. |
| parent_id | string | optional | parent_id - новый ID родительской локации. |
| latitude | double | optional | latitude - новая широта. |
| longitude | double | optional | longitude - новая долгота. |
| field_mask | string | repeated | field_mask - список полей для обновления. |

<details>
<summary>JSON Example</summary>

```json
{
  "fieldMask": [
    "string"
  ],
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

<a name="emh-v1-updatelocationresponse"></a>

### UpdateLocationResponse

UpdateLocationResponse обновленный объект локации.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| location | [Location](#emh-v1-location) | optional | location - актуальное состояние локации. |

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

<a name="emh-v1-deletelocationrequest"></a>

### DeleteLocationRequest

DeleteLocationRequest запрос на удаление локации.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | string | optional | id - UUID удаляемой локации. |

<details>
<summary>JSON Example</summary>

```json
{
  "id": "string"
}
```

</details>

<a name="emh-v1-deletelocationresponse"></a>

### DeleteLocationResponse

DeleteLocationResponse подтверждение удаления.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| success | bool | optional | success - true если удаление успешно. |

<details>
<summary>JSON Example</summary>

```json
{
  "success": true
}
```

</details>

<a name="emh-v1-reparentlocationsrequest"></a>

### ReparentLocationsRequest

ReparentLocationsRequest запрос на массовую перепривязку дочерних локаций.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| old_parent_id | string | optional | old_parent_id - ID старого родителя. |
| new_parent_id | string | optional | new_parent_id - ID нового родителя. |

<details>
<summary>JSON Example</summary>

```json
{
  "newParentId": "string",
  "oldParentId": "string"
}
```

</details>

<a name="emh-v1-reparentlocationsresponse"></a>

### ReparentLocationsResponse

ReparentLocationsResponse результат операции перепривязки.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| affected_count | int32 | optional | affected_count - количество успешно перенесенных локаций. |

<details>
<summary>JSON Example</summary>

```json
{
  "affectedCount": 0
}
```

</details>

