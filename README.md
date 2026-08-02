# NanoLink- A High-Performance URL Shortener

A production-ready URL shortener service built with Go and PostgreSQL, similar to Bitly/TinyURL. This project demonstrates clean architecture, efficient database design, and containerized deployment.

## Features

- **Redis Read-Through Caching:** Sub-millisecond latency for URL redirection by caching active URLs in RAM.
- **Rate Limiting:** Protects the API from abuse and DDoS attacks using Redis-backed IP tracking.
- **Short URL Generation:** Creates random base62 short codes with collision checks.
- **Custom Aliases:** Supports user-defined custom short URLs.
- **Click Tracking:** Asynchronous click counting for analytics decoupled from the main redirect flow.
- **Full CRUD:** Support for fetching, updating, and deleting shortened URLs.
- **Dockerized:** Fully containerized using Docker and Docker Compose.

## Tech Stack

- **Language:** Go (Golang)
- **Framework:** Standard Library (`net/http`)
- **Database:** PostgreSQL (with `database/sql` and `lib/pq`)
- **Cache & Rate Limiting:** Redis (with `go-redis/v9`)
- **Deployment:** Docker, Docker Compose

## Getting Started

### Prerequisites
- Docker and Docker Compose installed on your machine.

### Running the Project
```bash
docker-compose up -d --build
```
The service will be available at `http://localhost:8080`.

## API Documentation

### 1. Shorten URL (POST `/api/shorten`)
**Request Body:**
```json
{
  "original_url": "https://www.example.com",
  "custom_alias": "myalias"
}
```
**Response:**
```json
{
  "short_url": "http://localhost:8080/myalias"
}
```

### 2. Redirect URL (GET `/<short_code>`)
Redirects the user to the original URL via a 302 HTTP status code. Also asynchronously increments the click counter.

### 3. URL Statistics (GET `/api/stats?code=<short_code>`)
Returns metadata and click counts for a specific shortened URL.

### 4. Get All URLs (GET `/api/urls`)
Returns a JSON array of all shortened URLs stored in the system, sorted by creation date.

### 5. Update URL (PUT `/api/urls/update?code=<short_code>`)
Allows changing the destination (`original_url`) of an existing short code without changing the short link itself.
**Request Body:**
```json
{
  "original_url": "https://www.new-destination.com"
}
```

### 6. Delete URL (DELETE `/api/urls/delete?code=<short_code>`)
Permanently removes the URL from the system. Future requests to this short link will return a 404 Not Found.
