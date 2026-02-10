# SherlockOS Backend

Go backend service for the SherlockOS detective assistance platform — an AI-powered crime scene reconstruction and investigation system.

## Prerequisites

- Go 1.22+
- PostgreSQL (via Supabase)
- Redis (optional, falls back to in-memory queue)

## Project Structure

```
backend/
├── cmd/
│   ├── server/              # Application entrypoint
│   └── demo-seed/           # Demo data seeder
├── internal/
│   ├── api/                 # HTTP handlers and routing
│   ├── clients/             # External service clients
│   │   ├── gemini_client    # Gemini AI (reasoning, profiles, image gen)
│   │   ├── modal_client     # Modal (3D reconstruction, video replay)
│   │   └── storage_client   # Supabase Storage
│   ├── db/                  # Database connection and queries
│   ├── models/              # Data structures and validation
│   ├── queue/               # Redis/in-memory job queue
│   └── workers/             # Background job processors
├── pkg/
│   └── config/              # Configuration management
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

## AI Services

SherlockOS integrates multiple AI services for different capabilities:

| Service | Provider | Purpose |
|---------|----------|---------|
| **Reasoning** | Gemini 2.5 Flash | Trajectory hypothesis generation with thinking |
| **Profile Extraction** | Gemini 2.5 Flash | Extract suspect attributes from witness statements |
| **Image Generation** | Gemini Nano Banana | Suspect portraits, POV scenes, evidence boards |
| **Portrait Chat** | Gemini Nano Banana | Multi-turn iterative portrait refinement |
| **3D Reconstruction** | Modal (HunyuanWorld-Mirror) | Gaussian splatting from images/video |
| **Video Replay** | Modal (HY-World-1.5) | Camera trajectory video generation |
| **Scene Analysis** | Gemini 3 Pro Vision | Object detection from crime scene images |
| **3D Assets** | Replicate (Hunyuan3D-2) | Evidence 3D model generation |

## API Endpoints

### Cases
- `GET /v1/cases` - List all cases
- `POST /v1/cases` - Create a new case
- `GET /v1/cases/{caseId}` - Get case details
- `GET /v1/cases/{caseId}/snapshot` - Get current SceneGraph
- `GET /v1/cases/{caseId}/timeline` - List commits (timeline)

### Upload
- `POST /v1/cases/{caseId}/upload-intent` - Get presigned upload URLs

### Jobs
- `POST /v1/cases/{caseId}/jobs` - Create async job (reconstruction, imagegen, replay, asset3d, scene_analysis)
- `GET /v1/jobs/{jobId}` - Get job status and output

### Witness Statements
- `POST /v1/cases/{caseId}/witness-statements` - Submit statements (auto-triggers profile extraction)

### Portrait Generation
- `POST /v1/portrait/chat` - Multi-turn suspect portrait generation and refinement via Nano Banana

### Branches
- `POST /v1/cases/{caseId}/branches` - Create hypothesis branch

### Actions
- `POST /v1/cases/{caseId}/reasoning` - Trigger reasoning job
- `POST /v1/cases/{caseId}/export` - Trigger export job

## Job Types

| Type | Worker | Description |
|------|--------|-------------|
| `reconstruction` | ReconstructionWorker | 3D Gaussian splatting from images/video via Modal |
| `imagegen` | ImageGenWorker | Portrait, POV, evidence board generation via Nano Banana |
| `reasoning` | ReasoningWorker | Trajectory hypothesis generation via Gemini |
| `profile` | ProfileWorker | Suspect attribute extraction from witness statements |
| `replay` | ReplayWorker | Camera trajectory video via HY-World-1.5 |
| `asset3d` | Asset3DWorker | 3D evidence model via Hunyuan3D-2 |
| `scene_analysis` | SceneAnalysisWorker | Object detection via Gemini Vision |
| `export` | ExportWorker | HTML/PDF report generation |

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

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | - |
| `SUPABASE_URL` | Supabase project URL | - |
| `SUPABASE_ANON_KEY` | Supabase anonymous key | - |
| `SUPABASE_SECRET_KEY` | Supabase service role key | - |
| `REDIS_URL` | Redis connection URL | In-memory fallback |
| `GEMINI_API_KEY` | Google Gemini API key | - |
| `MODAL_MIRROR_URL` | Modal HunyuanWorld-Mirror base URL | - |
| `MODAL_WORLDPLAY_URL` | Modal HY-World-1.5 base URL | - |
| `REPLICATE_API_TOKEN` | Replicate API token (Hunyuan3D-2) | - |
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

## License

This project was built for the **Google Gemini API Developer Competition 2025**. It is provided as-is for demonstration and educational purposes under the [MIT License](../LICENSE).
