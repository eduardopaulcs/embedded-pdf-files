# AGENTS.md — Embedded PDF Files

## Project Overview

A Go web application that lets users upload PDFs, extract embedded files (PDF portfolio attachments), and download them individually or as a ZIP archive. No registration required, no database — everything is in-memory with automatic cleanup after 10 minutes.

Live site: https://filesinpdf.com

## Tech Stack

- **Go 1.26.2** — pure stdlib backend (`net/http`, `html/template`)
- **pdfcpu v0.12.0** — PDF embedded file extraction engine
- **go-cache** — in-memory caching for sessions and rate limiting
- **goldmark** — Markdown-to-HTML rendering (terms & privacy pages)
- **google/uuid** — session ID generation
- **HTML5 + CSS3 + Vanilla JavaScript** — no frontend frameworks
- **Docker + Docker Compose** — development & production environment
- **Air (air-verse/air)** — live-reload during development

## ⚠️ Critical Rule: Everything runs inside Docker

NEVER run server commands directly on the host machine.
Always use Docker Compose. The only commands allowed outside the container are:

```bash
docker compose up          # Start server with live-reload
docker compose up -d       # Start in background
docker compose stop        # Stop containers (preserves them)
docker compose down        # Stop and remove containers
docker compose logs -f     # View logs
docker compose exec app sh # Access the container shell
```

Any `go build`, `go run`, `go test`, or other Go commands must be executed
**inside** the container or via `docker compose exec app <command>`.

## Local Development

1. Copy `.env.example` to `.env` (optional)
2. `docker compose up`
3. Open http://localhost:8080
4. Server auto-reloads on changes to `.go`, `.html`, `.css`, `.js`, `.md` files

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `UPLOAD_LIMIT_MAX` | `3` | Max uploads per IP within the window |
| `UPLOAD_LIMIT_WINDOW` | `10m` | Rate-limit sliding window duration |
| `UMAMI_URL` | `https://cloud.umami.is` | Umami analytics host |
| `UMAMI_WEBSITE_ID` | `""` | Umami site ID (empty = disabled) |

## Project Structure

```
├── main.go                     # HTTP server, routes, handlers, rate limiting
├── internal/extractor/         # PDF extraction logic
│   └── extractor.go
├── resources/                  # Embedded static content (markdown, robots.txt, sitemap)
├── static/                     # Embedded frontend assets (CSS, JS, images)
├── templates/                  # Embedded HTML templates
├── Dockerfile                  # Multi-stage build (base → dev → builder → prod)
├── docker-compose.yml          # Dev environment with Air live-reload
├── .air.toml                   # Air configuration
├── go.mod / go.sum             # Go dependencies
└── .env.example                # Environment variable template
```

## HTTP Routes

| Route | Method | Description |
|---|---|---|
| `/` | GET | Main page (SPA) |
| `/upload` | POST | Upload PDF, returns JSON with session ID and file list |
| `/download?id=...&filename=...` | GET | Download individual file |
| `/download?id=...&all=true` | GET | Download all files as ZIP |
| `/terms` | GET | Terms of Service |
| `/privacy` | GET | Privacy Policy |
| `/robots.txt` | GET | SEO robots.txt |
| `/sitemap.xml` | GET | SEO sitemap XML |
| `/static/*` | GET | Static assets (CSS, JS, images) |

## Data Flow

1. User drags/clicks to select a PDF → frontend validates `.pdf` extension
2. POST to `/upload` as `multipart/form-data`
3. Server validates `%PDF` header, body limited to 10MB
4. Writes to temp file, uses pdfcpu to list & extract attachments
5. Extracted files read into memory, filenames sanitized (`[^a-zA-Z0-9_.]` → `_`)
6. If more than 1 file extracted, generates an in-memory ZIP archive
7. Stores result in cache with 10min TTL keyed by UUID
8. Returns JSON with session ID, file list, and ZIP availability
9. Frontend displays file list with individual download buttons and "Download All (ZIP)"
10. Downloads served via `/download` using session ID

## Rate Limiting

- Per-IP (priority: `X-Forwarded-For` → `X-Real-IP` → `RemoteAddr`)
- Sliding window in-memory via go-cache
- Default: 3 uploads per IP per 10 minutes

## Code Conventions

- No frontend frameworks — vanilla JavaScript only
- No frontend build step — CSS and JS served as-is
- All assets embedded into Go binary via `//go:embed`
- No database — everything in-memory
- Go packages: lowercase; files: snake_case
- Uses `var` in JS (not `let`/`const`) — consistent with existing code
- Line endings: LF enforced for Go, shell, Docker, config, markdown
- Branch strategy: `develop` ← `feature/*`, `master` as stable
