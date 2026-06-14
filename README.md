# Go User Management REST API

A simple REST API built with Go and PostgreSQL for user management and authentication.

## Features

* Create User
* Get All Users
* Get User by ID
* Update User
* Delete User
* Basic Login Endpoint
* Password Validation
* PostgreSQL Integration
* Environment Variable Configuration

## Tech Stack

* Go
* PostgreSQL
* net/http
* godotenv

## API Endpoints

| Method | Endpoint    | Description    |
| ------ | ----------- | -------------- |
| POST   | /users      | Create User    |
| GET    | /users      | Get All Users  |
| GET    | /users/{id} | Get User By ID |
| PUT    | /users/{id} | Update User    |
| DELETE | /users/{id} | Delete User    |
| POST   | /login      | User Login     |

## Validation

* Password must contain:

  * Minimum 8 characters
  * One uppercase letter
  * One lowercase letter
  * One number

## Run Locally

```bash
go mod tidy
go run main.go
```

Server runs on:

```text
http://localhost:8080
```
