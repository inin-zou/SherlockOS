# SherlockOS Backend

Go backend service for the SherlockOS detective assistance platform.

## Prerequisites

- Go 1.22+
- PostgreSQL (via Supabase)
- Redis

## Project Structure

```
backend/
├── cmd/
│   └── server/          # Application entrypoint
├── internal/
│   ├── api/             # HTTP handlers and routing
│   ├── clients/         # External service clients (AI, storage)
│   ├── db/              # Database connection and queries
│   ├── models/          # Data structures and validation
│   ├── queue/           # Redis job queue
│   └── workers/         # Background job workers
├── migrations/          # SQL migration files
├── pkg/
│   └── config/          # Configuration management
├── go.mod
├── Makefile
└── README.md
```

## Quick Start

1. **Install dependencies**
   ```bash
   make deps
   ```

2. **Set up environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run migrations**
   ```bash
   make migrate-up
   ```

4. **Start the server**
   ```bash
   make run
   ```

## Development

### Running Tests
```bash
make test
```

### Running Tests with Coverage
```bash
make test-coverage
```

### Formatting Code
```bash
make fmt
```

### Linting
```bash
make lint
```

## API Endpoints

### Cases
- `POST /v1/cases` - Create a new case
- `GET /v1/cases/{caseId}` - Get case details
- `GET /v1/cases/{caseId}/snapshot` - Get current SceneGraph
- `GET /v1/cases/{caseId}/timeline` - List commits (timeline)

### Upload
- `POST /v1/cases/{caseId}/upload-intent` - Get presigned upload URLs

### Jobs
- `POST /v1/cases/{caseId}/jobs` - Create async job
- `GET /v1/jobs/{jobId}` - Get job status

### Witness Statements
- `POST /v1/cases/{caseId}/witness-statements` - Submit statements

### Branches
- `POST /v1/cases/{caseId}/branches` - Create hypothesis branch

### Actions
- `POST /v1/cases/{caseId}/reasoning` - Trigger reasoning job
- `POST /v1/cases/{caseId}/export` - Trigger export job

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | - |
| `SUPABASE_URL` | Supabase project URL | - |
| `SUPABASE_ANON_KEY` | Supabase anonymous key | - |
| `SUPABASE_SECRET_KEY` | Supabase service role key | - |
| `REDIS_URL` | Redis connection URL | `redis://localhost:6379` |
| `GEMINI_API_KEY` | Google Gemini API key | - |
| `ALLOWED_ORIGINS` | CORS allowed origins | `http://localhost:3000` |

## Database Migrations

Create a new migration:
```bash
make migrate-create name=add_new_table
```

Run migrations:
```bash
make migrate-up
```

Rollback last migration:
```bash
make migrate-down
```
