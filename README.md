# chirpy

A RESTful API server built with guidance of a course on [Boot.dev](https://boot.dev)

### Goal

The goal with `chirpy` is to learn how HTTP servers are traditionally done in Golang:

* Using an HTTP request multiplexer
* Handling CRUD routes (GET, POST, PUT, DELETE)
* Reading to and writing from a PostgreSQL database
* Storing sensitive data in `.env` file and exposing it to the server with [godotenv](https://pkg.go.dev/github.com/joho/Godotenv)
* Managing database migrations with [goose](https://pkg.go.dev/github.com/pressly/goose/v3) and generating type-safe code with [SQLC](https://github.com/sqlc-dev/sqlc)
* Hashing the passwords before storing them in the database using Bcrypt
* Using [JWTs](https://pkg.go.dev/github.com/golang-jwt/jwt/v5) and refresh tokens for authentication and authorization
* Handling webhooks

## ⚙️ Installation

### Prerequisites

- [Go toolchain](https://webinstall.dev/golang/)
- [PostgreSQL](https://www.postgresql.org/download/)
- goose
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### Configuration

1. Clone the repository to a desired location:
```bash
git clone https://github.com/Burush0/chirpy.git
```
2. Create a new database in PostgreSQL (I named mine `chirpy`): (example instructions on Linux)
```bash
sudo -u postgres psql
```
```sql
CREATE DATABASE chirpy;
ALTER USER postgres PASSWORD '<your_password>';
```
I used `postgres` as a "default" password for the database.

3. Obtain the database connection string. The format is as follows:
```
protocol://username:password@host:port/database
```
For example, my string ended up looking like this:
```
postgres://postgres:postgres@localhost:5432/chirpy
```
4. Enter the `sql/schema` directory in the project and run the following command:
```bash
goose postgres <your_database_string> up
```
This will create all the necessary tables in the database.

5. Configure the `.env` file:
    1. Create a file named `.env` in the root folder of the project.
    2. The following four fields are required for the server to work:
    ```
    DB_URL="<your_database_string>?sslmode=disable"
    PLATFORM="dev"
    JWT_SECRET="somesecret"
    POLKA_KEY="someapikey"
    ```
    3. To generate a strong JWT secret, you can use this command
    ```bash
    openssl rand -base64 64
    ```
    4. The `POLKA_KEY` field was a part of the guided course, you can put anything there, just beware that the endpoint relying on it (`POST /api/polka/webhooks`) will rely on the `Authorization` header of this shape: `Authorization: ApiKey someapikey`

6. You should now be able to run the server with either of these commands (run from the root of the project):
```bash
go run .
```
```bash
go build -o out && ./out
```

## Endpoints

### TL;DR

1. **`POST /api/users`**: Create a new user.
2. **`POST /api/login`**: Log in a user and return a JWT and refresh token.
3. **`PUT /api/users`**: Update a user's email and/or password.
4. **`POST /api/refresh`**: Refresh a user's access token using a valid refresh token.
5. **`POST /api/revoke`**: Revoke a refresh token.
6. **`POST /api/chirps`**: Create a new chirp.
7. **`GET /api/chirps`**: Retrieve a list of chirps (filterable by author and sortable by creation time).
8. **`GET /api/chirps/{chirpID}`**: Retrieve a specific chirp by its ID.
9. **`DELETE /api/chirps/{chirpID}`**: Delete a specific chirp by its ID.
10. **`POST /api/polka/webhooks`**: Handle Polka webhook events (e.g., upgrade user to Chirpy Red).
11. **`POST /admin/reset`**: Reset server metrics and database (dev environment only).
12. **`GET /admin/metrics`**: Display file server visit metrics in an HTML page.
13. **`GET /api/healthz`**: Health check endpoint.

### Users

#### **1. `POST /api/users`**
- **Description**: Creates a new user.
- **Request Body**:
  ```json
  {
    "email": "string",
    "password": "string"
  }
  ```
- **Response Body** (on success):
  ```json
  {
    "id": "uuid",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "email": "string"
  }
  ```
- **Response Codes**:
  - `201 Created`: User successfully created.
  - `500 Internal Server Error`: Failed to decode parameters, hash password, or create user.

---

#### **2. `POST /api/login`**
- **Description**: Logs in a user and returns a JWT token and a refresh token.
- **Request Body**:
  ```json
  {
    "email": "string",
    "password": "string"
  }
  ```
- **Response Body** (on success):
  ```json
  {
    "id": "uuid",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "email": "string",
    "token": "string (JWT)",
    "refresh_token": "string",
    "is_chirpy_red": "boolean"
  }
  ```
- **Response Codes**:
  - `200 OK`: Login successful.
  - `401 Unauthorized`: Incorrect email or password.
  - `500 Internal Server Error`: Failed to decode parameters, generate JWT, or create refresh token.

---

#### **3. `PUT /api/users`**
- **Description**: Updates a user's email and/or password. Requires authentication via JWT.
- **Authentication**: Bearer token in the `Authorization` header.
- **Request Body**:
  ```json
  {
    "email": "string",
    "password": "string"
  }
  ```
- **Response Body** (on success):
  ```json
  {
    "email": "string"
  }
  ```
- **Response Codes**:
  - `200 OK`: User successfully updated.
  - `401 Unauthorized`: Invalid or missing JWT.
  - `500 Internal Server Error`: Failed to decode parameters, hash password, or update user.

---

- **Error Response**:
  ```json
  {
    "error": "string",
    "message": "string",
    "details": "string (optional)"
  }
  ```
- **Authentication**:
  - For endpoints requiring authentication, a valid JWT must be provided in the `Authorization` header as a Bearer token.
  - Example: `Authorization: Bearer <JWT>`


---

### Refresh tokens

#### **4. `POST /api/refresh`**
- **Description**: Refreshes a user's access token using a valid refresh token.
- **Authentication**: Bearer token in the `Authorization` header (refresh token).
- **Request Body**: None.
- **Response Body** (on success):
  ```json
  {
    "token": "string (new JWT)"
  }
  ```
- **Response Codes**:
  - `200 OK`: New access token generated successfully.
  - `401 Unauthorized`: Invalid, expired, or revoked refresh token.
  - `500 Internal Server Error`: Failed to retrieve user or generate JWT.

---

#### **5. `POST /api/revoke`**
- **Description**: Revokes a refresh token, making it invalid for future use.
- **Authentication**: Bearer token in the `Authorization` header (refresh token).
- **Request Body**: None.
- **Response Body**: None.
- **Response Codes**:
  - `204 No Content`: Refresh token successfully revoked.
  - `401 Unauthorized`: Invalid or missing refresh token.
  - `500 Internal Server Error`: Failed to revoke refresh token.

---

- **Authentication**:
  - Both endpoints require a valid refresh token in the `Authorization` header as a Bearer token.
  - Example: `Authorization: Bearer <refresh_token>`

---

### Chirps

#### **6. `POST /api/chirps`**
- **Description**: Creates a new chirp. Requires authentication via JWT.
- **Authentication**: Bearer token in the `Authorization` header.
- **Request Body**:
  ```json
  {
    "body": "string"
  }
  ```
- **Response Body** (on success):
  ```json
  {
    "id": "uuid",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "body": "string",
    "user_id": "uuid"
  }
  ```
- **Response Codes**:
  - `201 Created`: Chirp successfully created.
  - `400 Bad Request`: Chirp body is too long (max 140 characters) or contains invalid content.
  - `401 Unauthorized`: Invalid or missing JWT.
  - `500 Internal Server Error`: Failed to decode parameters or create chirp.

---

#### **7. `GET /api/chirps`**
- **Description**: Retrieves a list of chirps. Can be filtered by author and sorted by creation time.
- **Query Parameters**:
  - `author_id` (optional): Filters chirps by the specified user ID.
  - `sort` (optional): Sorts chirps by creation time. Valid values: `asc` (default) or `desc`.
- **Response Body** (on success):
  ```jsonc
  [
    {
      "id": "uuid",
      "created_at": "timestamp",
      "updated_at": "timestamp",
      "body": "string",
      "user_id": "uuid"
    },
    // ...
  ]
  ```
- **Response Codes**:
  - `200 OK`: Chirps retrieved successfully.
  - `400 Bad Request`: Invalid `author_id` format.
  - `500 Internal Server Error`: Failed to retrieve chirps.

---

#### **8. `GET /api/chirps/{chirpID}`**
- **Description**: Retrieves a specific chirp by its ID.
- **Path Parameters**:
  - `chirpID`: The UUID of the chirp to retrieve.
- **Response Body** (on success):
  ```json
  {
    "id": "uuid",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "body": "string",
    "user_id": "uuid"
  }
  ```
- **Response Codes**:
  - `200 OK`: Chirp retrieved successfully.
  - `400 Bad Request`: Invalid `chirpID` format.
  - `404 Not Found`: Chirp with the specified ID does not exist.
  - `500 Internal Server Error`: Failed to retrieve chirp.

---

#### **9. `DELETE /api/chirps/{chirpID}`**
- **Description**: Deletes a specific chirp by its ID. Requires authentication via JWT, and the user must be the owner of the chirp.
- **Authentication**: Bearer token in the `Authorization` header.
- **Path Parameters**:
  - `chirpID`: The UUID of the chirp to delete.
- **Response Body**: None.
- **Response Codes**:
  - `204 No Content`: Chirp successfully deleted.
  - `400 Bad Request`: Invalid `chirpID` format.
  - `401 Unauthorized`: Invalid or missing JWT.
  - `403 Forbidden`: User is not the owner of the chirp.
  - `404 Not Found`: Chirp with the specified ID does not exist.
  - `500 Internal Server Error`: Failed to delete chirp.

---

- **Authentication**:
  - Endpoints requiring authentication expect a valid JWT in the `Authorization` header as a Bearer token.
  - Example: `Authorization: Bearer <JWT>`


---

### Webhooks

#### **10. `POST /api/polka/webhooks`**
- **Description**: Handles webhook events from Polka. Specifically, it processes the `user.upgraded` event to upgrade a user to Chirpy Red status.
- **Authentication**: API key in the `Authorization` header.
- **Request Body**:
  ```json
  {
    "event": "string",
    "data": {
      "user_id": "uuid"
    }
  }
  ```
- **Response Body**: None.
- **Response Codes**:
  - `204 No Content`: Webhook processed successfully (valid event or no action required).
  - `401 Unauthorized`: Invalid or missing API key.
  - `404 Not Found`: User with the specified ID does not exist.
  - `500 Internal Server Error`: Failed to decode parameters or update user.

---

- **Authentication**:
  - The endpoint requires a valid API key in the `Authorization` header.
  - Example: `Authorization: ApiKey <polka_key>`
- **Event Handling**:
  - Only the `user.upgraded` event triggers an action (upgrading the user to Chirpy Red).
  - Other events are ignored, and the endpoint responds with `204 No Content`.

---

#### **11. `POST /admin/reset`**
- **Description**: Resets the server's metrics (file server hits) and database to their initial state. Only available in the `dev` environment.
- **Environment Restriction**: Only works in the `dev` environment.
- **Request Body**: None.
- **Response Body** (on success):
  ```
  Hits reset to 0 and database reset to initial state.
  ```
- **Response Codes**:
  - `200 OK`: Reset successful.
  - `403 Forbidden`: Reset is not allowed in the current environment (e.g., production).

---

#### **12. `GET /admin/metrics`**
- **Description**: Returns an HTML page displaying the number of times the file server has been visited.
- **Response Body** (on success):
  ```html
  <html>
    <body>
      <h1>Welcome, Chirpy Admin</h1>
      <p>Chirpy has been visited {count} times!</p>
    </body>
  </html>
  ```
- **Response Codes**:
  - `200 OK`: Metrics page successfully returned.

---

### **13. `GET /api/healthz`**
- **Description**: A health check endpoint to verify that the server is running.
- **Request Body**: None.
- **Response Body**:
  ```
  OK
  ```
- **Response Codes**:
  - `200 OK`: Server is healthy and running.

---

### **Middleware: `middlewareMetricsInc`**
- **Description**: Middleware that increments the file server hit counter for every request to the file server.
- **Usage**: Applied to the `/app/` endpoint to track visits to the file server.
