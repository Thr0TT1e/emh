# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [emh/v1/auth.proto](#emh-v1-auth-proto)
  - **Services**
    - [AuthService](#emh-v1-authservice)
  - **Messages**
    - [LoginRequest](#emh-v1-loginrequest)
    - [LoginResponse](#emh-v1-loginresponse)
    - [RefreshTokenRequest](#emh-v1-refreshtokenrequest)
    - [RefreshTokenResponse](#emh-v1-refreshtokenresponse)
    - [LogoutRequest](#emh-v1-logoutrequest)
    - [LogoutResponse](#emh-v1-logoutresponse)

<a name="emh-v1-auth-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## emh/v1/auth.proto

**Package:** `emh.v1`

<a name="emh-v1-authservice"></a>

## AuthService

AuthService публичный сервис аутентификации (точка входа, не защищён).

### Methods Overview

| Method | Type | HTTP | Description |
| ------ | ---- | ---- | ----------- |
| [Login](#emh-v1-authservice-login) | ➡️ Unary | — | Login выдаёт пару access/refresh токенов п... |
| [Refresh](#emh-v1-authservice-refresh) | ➡️ Unary | — | Refresh обменивает действующий refresh-... |
| [Logout](#emh-v1-authservice-logout) | ➡️ Unary | — | Logout отзывает refresh-токен (выход из ... |

<a name="emh-v1-authservice-login"></a>

### Login

```protobuf
rpc Login([LoginRequest](#emh-v1-loginrequest)) returns ([LoginResponse](#emh-v1-loginresponse))
```

Login выдаёт пару access/refresh токенов при успешной проверке учётных данных.

#### Request Example

```json
{
  "password": "string",
  "username": "string"
}
```

#### Response Example

```json
{
  "accessToken": "string",
  "expiresAt": "string",
  "expiresIn": 0,
  "refreshToken": "string",
  "tokenType": "string"
}
```

---

<a name="emh-v1-authservice-refresh"></a>

### Refresh

```protobuf
rpc Refresh([RefreshTokenRequest](#emh-v1-refreshtokenrequest)) returns ([RefreshTokenResponse](#emh-v1-refreshtokenresponse))
```

Refresh обменивает действующий refresh-токен на новую пару access/refresh.

#### Request Example

```json
{
  "refreshToken": "string"
}
```

#### Response Example

```json
{
  "accessToken": "string",
  "expiresAt": "string",
  "expiresIn": 0,
  "refreshToken": "string",
  "tokenType": "string"
}
```

---

<a name="emh-v1-authservice-logout"></a>

### Logout

```protobuf
rpc Logout([LogoutRequest](#emh-v1-logoutrequest)) returns ([LogoutResponse](#emh-v1-logoutresponse))
```

Logout отзывает refresh-токен (выход из системы).

#### Request Example

```json
{
  "refreshToken": "string"
}
```

#### Response Example

```json
{
  "success": true
}
```

---

<a name="emh-v1-loginrequest"></a>

### LoginRequest

LoginRequest запрос на аутентификацию администратора.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| username | string | optional | username - имя пользователя (обязательно). |
| password | string | optional | password - пароль (обязательно). |

<details>
<summary>JSON Example</summary>

```json
{
  "password": "string",
  "username": "string"
}
```

</details>

<a name="emh-v1-loginresponse"></a>

### LoginResponse

LoginResponse результат успешной аутентификации.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| access_token | string | optional | access_token - JWT для запросов к админским сервисам. |
| token_type | string | optional | token_type - тип токена (Bearer). |
| expires_at | string | optional | expires_at - время истечения access-токена (RFC3339). |
| refresh_token | string | optional | refresh_token - токен для получения новой пары access/refresh. |
| expires_in | int64 | optional | expires_in - время жизни access-токена в секундах. |

<details>
<summary>JSON Example</summary>

```json
{
  "accessToken": "string",
  "expiresAt": "string",
  "expiresIn": 0,
  "refreshToken": "string",
  "tokenType": "string"
}
```

</details>

<a name="emh-v1-refreshtokenrequest"></a>

### RefreshTokenRequest

RefreshTokenRequest запрос на обновление токенов.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| refresh_token | string | optional | refresh_token - действующий refresh-токен. |

<details>
<summary>JSON Example</summary>

```json
{
  "refreshToken": "string"
}
```

</details>

<a name="emh-v1-refreshtokenresponse"></a>

### RefreshTokenResponse

RefreshTokenResponse новая пара токенов после ротации.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| access_token | string | optional | access_token - новый JWT. |
| token_type | string | optional | token_type - тип токена (Bearer). |
| expires_at | string | optional | expires_at - время истечения access-токена (RFC3339). |
| refresh_token | string | optional | refresh_token - новый refresh-токен (старый отозван). |
| expires_in | int64 | optional | expires_in - время жизни access-токена в секундах. |

<details>
<summary>JSON Example</summary>

```json
{
  "accessToken": "string",
  "expiresAt": "string",
  "expiresIn": 0,
  "refreshToken": "string",
  "tokenType": "string"
}
```

</details>

<a name="emh-v1-logoutrequest"></a>

### LogoutRequest

LogoutRequest запрос на отзыв refresh-токена.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| refresh_token | string | optional | refresh_token - токен для отзыва. |

<details>
<summary>JSON Example</summary>

```json
{
  "refreshToken": "string"
}
```

</details>

<a name="emh-v1-logoutresponse"></a>

### LogoutResponse

LogoutResponse результат выхода из системы.

| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| success | bool | optional | success - true если токен успешно отозван. |

<details>
<summary>JSON Example</summary>

```json
{
  "success": true
}
```

</details>

