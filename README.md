# PhotoOrg -- AI-Powered Photo Organizer v2

Vision-driven photo organizer that uses LLM vision models (Ollama, OpenAI-compatible) to categorize images into folders. Browse a directory, configure your AI provider and categories, let the model classify every image, review and drag-and-drop corrections, then commit the file moves/copies.

**Status:** Active
**Stack:** Go 1.24 (Gin, sqlx, modernc/sqlite) + React 19 (TypeScript, Vite 6, TailwindCSS v4, Zustand 5, dnd-kit)
**Ports:** 3011 (frontend dev), 8012 (backend dev) | See [port-registry](../../references/port-registry.md)
**Deploy:** Local desktop tool (single binary via go:embed, not deployed to Coolify)
**Remote:** N/A (lives in Silvina monorepo)

## Features

- Browse filesystem and select folders (Windows drive detection, image count badges)
- Ollama and OpenAI-compatible vision LLM providers with model discovery
- Concurrent AI categorization with real-time SSE progress
- Two-phase workflow: categorize to DB first, then review before committing
- Review grid with drag-and-drop recategorization between category sections
- Full-size image lightbox with keyboard navigation and inline reassignment
- Commit moves or copies with automatic collision handling
- Full undo support (moves reversed, copies deleted)
- Job history with status tracking
- Industrial dark theme with custom palette

## Getting Started

```bash
# 1. Install frontend dependencies
cd frontend && npm install && cd ..

# 2. Configure environment (optional, defaults work for local Ollama)
cp .env.example .env

# 3. Build frontend
cd frontend && npm run build && cd ..

# 4. Run
go run .

# Open http://localhost:8012
```

### Development mode (hot reload)

```bash
# Terminal 1: Backend
go run .

# Terminal 2: Frontend (proxies API to :8012)
cd frontend && npm run dev

# Open http://localhost:3011
```

## Architecture

Single Go binary with embedded React SPA via `go:embed`. SQLite stores job history, file records, and app config. The categorization engine runs as a background goroutine with concurrent LLM calls (channel-based semaphore) and publishes real-time updates via SSE fan-out broker. Thumbnails are generated on-demand with SHA256-based disk cache.

```
main.go (embed + config + db + sse broker + graceful shutdown)
internal/
  api/        Gin router, handlers, middleware, SPA serving
  config/     Env-based config
  database/   SQLite init, migrations, queries (sqlx)
  engine/     Scanner, categorizer, committer, thumbnail cache
  llm/        Vision provider interface + Ollama/OpenAI implementations
  sse/        Fan-out SSE broker
frontend/
  src/
    api/        Typed Axios client
    components/ FolderBrowser, ProviderConfig, CategoryEditor,
                CategorizeProgress, ReviewGrid, ImageCard,
                ImageLightbox, Toasts
    hooks/      useSSE (EventSource -> Zustand)
    pages/      SetupPage, ReviewPage, HistoryPage
    stores/     useSettingsStore, useJobStore, useToastStore
```

## Workflow

1. **Setup**: Browse to a photo folder, pick LLM provider/model, set categories
2. **Categorize**: AI processes each image concurrently, progress streams via SSE
3. **Review**: Grid grouped by category, drag images between sections, click for lightbox
4. **Commit**: Move or copy files into category subfolders (with `.photo_org_managed` markers)
5. **Undo**: Reverse committed moves from History page

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| API_PORT | 8012 | Backend server port |
| API_HOST | 0.0.0.0 | Bind address |
| API_CORS_ORIGINS | http://localhost:3011 | CORS origins (comma-separated) |
| DATABASE_PATH | data/photoorg.db | SQLite database path |
| THUMBNAIL_DIR | data/thumbnails | Thumbnail cache directory |
| LOG_LEVEL | info | Log level (debug, info, warn, error) |
| ENV | development | Environment (development, production) |

## Gotchas

- Requires an Ollama instance with a vision model (e.g., `llava:7b`) for local use
- `go:embed` uses `all:frontend/dist` directive to include hashed asset filenames
- Managed folders (`.photo_org_managed` marker) are skipped during scanning to prevent re-processing
- Thumbnail cache grows unbounded; clear `data/thumbnails/` if disk space is a concern
- SQLite uses WAL mode with 5s busy timeout for concurrent read/write during categorization
- Image counts in folder browser are shallow (immediate directory only) for performance
