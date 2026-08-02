# NanoLink - High-Performance URL Shortener

NanoLink is a fast, containerized URL shortener written in Go. It uses PostgreSQL for persistent storage and Redis for read-through caching and rate limiting.

I built this project to dive deeper into Go concurrency, standard library HTTP routing, and Docker.

## Architecture

- **Go (`net/http`)**: Core API server.
- **PostgreSQL**: Stores URL mappings and analytics (clicks).
- **Redis**: Caches short codes for faster redirects and tracks IP requests to limit abuse.

## Running the project

You need Docker and Docker Compose installed.

```bash
git clone https://github.com/parthsarthi-dutt/NanoLink.git
cd NanoLink
docker-compose up -d --build
```
The server will start on `http://localhost:8080`.

## API Endpoints

- `POST /api/shorten` - Shorten a URL (accepts `original_url` and optional `custom_alias`)
- `GET /<short_code>` - Redirect to the original URL
- `GET /api/stats?code=<short_code>` - View click count
- `GET /api/urls` - Fetch all URLs
- `PUT /api/urls/update?code=<short_code>` - Update a destination URL
- `DELETE /api/urls/delete?code=<short_code>` - Delete a short code

## To-Do / Future Improvements

- Pre-generate Base62 codes via an offline Key Generation Service (KGS) to completely eliminate DB collisions on write.
- Add OAuth2 for user accounts.
- Add Prometheus/Grafana metrics.
