# SherlockOS - Product Requirements Document (PRD)

> **Version:** 1.0
> **Last Updated:** 2026-02-01
> **Status:** Draft for Hackathon MVP

---

## 1. Executive Summary

SherlockOS is a detective assistance web application built on "World Model + Timeline + Explainable Reasoning" architecture. It helps investigators reconstruct crime scenes, profile suspects from witness testimony, and generate evidence-based movement trajectory hypotheses.

**Core Principle:** SceneGraph (structured world state) is the Single Source of Truth. All reasoning derives from structured data, not raw images.

---

## 2. Tech Stack Summary

| Layer | Technology |
|-------|------------|
| Backend | Go 1.22+ (chi router, pgx) |
| Frontend | Next.js, TypeScript, three.js, Zustand, Tailwind |
| Database | Supabase (Postgres + Storage + Realtime) |
| Queue | Redis Streams (hackathon) |
| Scene Reconstruction | HunyuanWorld-Mirror / HunyuanWorld-1.0 |
| Image Generation | Nano Banana (gemini-2.5-flash-image) |
| Reasoning | Gemini 2.5 Flash with Thinking |

---

## 3. Development Phases

### Phase 1: Core Data Layer
**Goal:** Establish foundational data structures and database schema.

#### Objective 1.1: Database Schema Setup
**Description:** Create all required database tables, enums, and indexes.

**Deliverables:**
- [ ] SQL migration files for all enum types (commit_type, job_type, job_status, asset_kind)
- [ ] `cases` table with constraints
- [ ] `commits` table (append-only timeline)
- [ ] `branches` table for hypothesis branching
- [ ] `scene_snapshots` table (current state)
- [ ] `suspect_profiles` table
- [ ] `jobs` table with idempotency support
- [ ] `assets` table

**Unit Tests:**
- Test enum validation for all types
- Test foreign key constraints
- Test unique constraints (idempotency_key, branch name per case)
- Test index performance on common queries

---

#### Objective 1.2: SceneGraph Data Structures (Go)
**Description:** Implement Go structs for SceneGraph and related types.

**Deliverables:**
- [ ] `SceneGraph` struct with JSON marshaling
- [ ] `SceneObject` struct with ObjectType enum
- [ ] `Pose`, `BoundingBox` structs
- [ ] `EvidenceCard`, `EvidenceSource` structs
- [ ] `Constraint`, `UncertaintyRegion` structs
- [ ] Validation functions for all structs

**Unit Tests:**
- Test JSON serialization/deserialization round-trip
- Test validation (required fields, value ranges)
- Test ObjectType enum parsing
- Test confidence value bounds (0-1)
- Test BoundingBox validity (min < max)

---

#### Objective 1.3: Commit/Timeline Logic
**Description:** Implement commit creation and timeline traversal.

**Deliverables:**
- [ ] `CreateCommit` function with parent linking
- [ ] `GetCommitsByCase` with pagination (cursor-based)
- [ ] `GetCommitDiff` to compute changes between commits
- [ ] `ReplayToCommit` to reconstruct state at any point

**Unit Tests:**
- Test commit chain integrity (parent references)
- Test pagination cursor encoding/decoding
- Test diff computation (additions, updates, deletions)
- Test replay produces correct SceneGraph at each commit

---

### Phase 2: API Layer
**Goal:** Implement REST API endpoints for case and job management.

#### Objective 2.1: Case Management APIs
**Description:** CRUD operations for cases.

**Deliverables:**
- [ ] `POST /v1/cases` - Create case
- [ ] `GET /v1/cases/{caseId}` - Get case details
- [ ] `GET /v1/cases/{caseId}/snapshot` - Get current SceneGraph
- [ ] `GET /v1/cases/{caseId}/timeline` - List commits with pagination

**Unit Tests:**
- Test case creation with valid/invalid input
- Test case title length constraint (<=200 chars)
- Test snapshot returns latest commit's SceneGraph
- Test timeline pagination (cursor, limit)
- Test 404 for non-existent case

---

#### Objective 2.2: Upload & Asset Management
**Description:** Handle file uploads via presigned URLs.

**Deliverables:**
- [ ] `POST /v1/cases/{caseId}/upload-intent` - Generate presigned URLs
- [ ] Presigned URL generation with expiry
- [ ] Asset record creation after successful upload
- [ ] Storage key path convention: `cases/{caseId}/{kind}/{batchId}/{filename}`

**Unit Tests:**
- Test presigned URL generation
- Test batch upload intent (multiple files)
- Test file size/type validation
- Test storage key format correctness
- Test URL expiry configuration

---

#### Objective 2.3: Job Management APIs
**Description:** Job creation, status tracking, and idempotency.

**Deliverables:**
- [ ] `POST /v1/cases/{caseId}/jobs` - Create job with idempotency key
- [ ] `GET /v1/jobs/{jobId}` - Get job status
- [ ] Job type validation (reconstruction, imagegen, reasoning, profile, export)
- [ ] Idempotency: return existing job for duplicate key

**Unit Tests:**
- Test job creation for each job type
- Test idempotency key conflict (409 response)
- Test job status transitions
- Test progress bounds (0-100)
- Test job input schema validation per type

---

#### Objective 2.4: Witness Statement API
**Description:** Submit witness testimony and trigger profile updates.

**Deliverables:**
- [ ] `POST /v1/cases/{caseId}/witness-statements` - Submit statements
- [ ] Create `witness_statement` commit
- [ ] Auto-trigger `profile` job
- [ ] Credibility scoring (0-1)

**Unit Tests:**
- Test statement submission creates commit
- Test profile job auto-creation
- Test credibility validation (0-1 range)
- Test multiple statements in single request

---

### Phase 3: Worker Infrastructure
**Goal:** Build robust async job processing system.

#### Objective 3.1: Job Queue System
**Description:** Implement Redis-based job queue with reliable delivery.

**Deliverables:**
- [ ] Job enqueue function
- [ ] Job dequeue with visibility timeout
- [ ] Job acknowledgment (success/failure)
- [ ] Dead letter queue for failed jobs

**Unit Tests:**
- Test enqueue/dequeue ordering (FIFO)
- Test visibility timeout re-queues unacknowledged jobs
- Test acknowledgment removes job from queue
- Test dead letter after max retries

---

#### Objective 3.2: Worker Base Infrastructure
**Description:** Generic worker framework with lifecycle management.

**Deliverables:**
- [ ] Worker interface definition
- [ ] Worker registration and startup
- [ ] Graceful shutdown handling
- [ ] Progress reporting callback
- [ ] Error handling with categorization

**Unit Tests:**
- Test worker starts and processes jobs
- Test graceful shutdown completes in-flight work
- Test progress updates persist to database
- Test error categorization (retryable vs fatal)

---

#### Objective 3.3: Retry & Heartbeat System
**Description:** Implement exponential backoff retry and heartbeat monitoring.

**Deliverables:**
- [ ] Exponential backoff retry (2s → 4s → 8s, max 30s)
- [ ] Max retry count (default 3)
- [ ] Heartbeat update every 30s
- [ ] Zombie job detection (no heartbeat for 2min)
- [ ] Zombie job recovery (re-queue or fail)

**Unit Tests:**
- Test retry intervals follow exponential backoff
- Test max retry exhaustion marks job failed
- Test heartbeat updates `updated_at` timestamp
- Test zombie detection query correctness
- Test zombie recovery respects retry count

---

### Phase 4: AI Worker Integration
**Goal:** Implement AI-powered workers with proper mocking for tests.

#### Objective 4.1: Reconstruction Worker
**Description:** Process scan images to generate/update SceneGraph.

**Deliverables:**
- [ ] `ReconstructionInput` validation
- [ ] HunyuanWorld-Mirror API client (with interface for mocking)
- [ ] `ReconstructionOutput` to SceneGraph conversion
- [ ] Commit creation with `reconstruction_update` type
- [ ] Fallback to mock SceneGraph on failure

**Unit Tests:**
- Test input validation (required fields, asset keys exist)
- Test output parsing from mock API response
- Test SceneGraph merge logic (create/update/remove objects)
- Test commit payload contains correct diff
- Test fallback produces valid mock SceneGraph

---

#### Objective 4.2: Profile Worker
**Description:** Extract structured attributes from witness statements.

**Deliverables:**
- [ ] Statement parsing with NLP/LLM
- [ ] Attribute extraction (age, height, build, etc.)
- [ ] Confidence aggregation from multiple sources
- [ ] Conflict detection between statements
- [ ] Trigger ImageGen job for portrait

**Unit Tests:**
- Test attribute extraction from sample statements
- Test confidence weighted by source credibility
- Test conflict detection (contradicting descriptions)
- Test ImageGen job triggered after profile update
- Test incremental attribute update (merge with existing)

---

#### Objective 4.3: ImageGen Worker (Nano Banana)
**Description:** Generate suspect portraits and evidence visuals.

**Deliverables:**
- [ ] `ImageGenInput` validation
- [ ] Nano Banana API client (with interface for mocking)
- [ ] Model selection logic (Nano Banana vs Pro based on resolution)
- [ ] Image upload to Supabase Storage
- [ ] Thumbnail generation

**Unit Tests:**
- Test input validation per gen_type
- Test model selection (1k→Nano Banana, 2k/4k→Pro)
- Test portrait prompt construction from attributes
- Test asset storage key generation
- Test cost calculation correctness

---

#### Objective 4.4: Reasoning Worker (Gemini 2.5 Flash)
**Description:** Generate trajectory hypotheses with explanations.

**Deliverables:**
- [ ] `ReasoningInput` validation
- [ ] Gemini 2.5 Flash API client with Thinking config
- [ ] Prompt template rendering
- [ ] `ReasoningOutput` parsing (trajectories, segments, evidence refs)
- [ ] WebSocket streaming for thinking chunks
- [ ] Commit creation with `reasoning_result` type

**Unit Tests:**
- Test input validation (SceneGraph required)
- Test thinking_budget configuration (0-24576)
- Test prompt template rendering with SceneGraph JSON
- Test output parsing from mock API response
- Test trajectory segment validation (positions, evidence refs)
- Test streaming chunk delivery

---

### Phase 5: Frontend Core
**Goal:** Build core UI components with testable state management.

#### Objective 5.1: State Management (Zustand)
**Description:** Implement global state stores for app data.

**Deliverables:**
- [ ] `useCaseStore` - Current case, commits, snapshot
- [ ] `useJobStore` - Active jobs, progress tracking
- [ ] `useViewModeStore` - Evidence/Reasoning/Explain mode toggle
- [ ] `useSelectionStore` - Selected objects, commits, trajectory segments

**Unit Tests:**
- Test store initialization with default values
- Test state updates (add commit, update job progress)
- Test derived state (filtered commits, active jobs count)
- Test store reset on case change

---

#### Objective 5.2: Timeline Component
**Description:** Display commit history with diff visualization.

**Deliverables:**
- [ ] Commit list with type icons
- [ ] Commit summary and timestamp display
- [ ] Diff view (objects added/updated/removed)
- [ ] Click-to-select for state replay
- [ ] Auto-scroll on new commit

**Unit Tests:**
- Test commit list renders correct count
- Test type icon mapping
- Test diff calculation display
- Test selection state updates on click
- Test real-time subscription updates list

---

#### Objective 5.3: Scene View (three.js)
**Description:** 3D visualization of SceneGraph.

**Deliverables:**
- [ ] Scene setup (camera, lights, controls)
- [ ] Object rendering from SceneGraph
- [ ] Object type-based materials/colors
- [ ] Hover highlight effect
- [ ] Click-to-select object
- [ ] Evidence card popup on selection

**Unit Tests:**
- Test scene initializes without errors
- Test objects render at correct positions
- Test object count matches SceneGraph
- Test selection state sync with store
- Test camera bounds fit scene bounds

---

#### Objective 5.4: Reasoning Panel
**Description:** Display trajectory hypotheses and explanations.

**Deliverables:**
- [ ] Trajectory list (ranked by confidence)
- [ ] Trajectory path visualization (3D overlay)
- [ ] Segment click → highlight + evidence cards
- [ ] Evidence reference links
- [ ] Confidence score display

**Unit Tests:**
- Test trajectory list renders all hypotheses
- Test ranking order (highest confidence first)
- Test segment selection highlights correct path
- Test evidence card display on segment click
- Test confidence score formatting

---

### Phase 6: Integration & E2E Features
**Goal:** Complete end-to-end flows and export functionality.

#### Objective 6.1: Real-time Updates (Supabase Realtime)
**Description:** Subscribe to database changes for live updates.

**Deliverables:**
- [ ] Commits table subscription (INSERT)
- [ ] Jobs table subscription (UPDATE)
- [ ] Connection state management
- [ ] Reconnection handling

**Unit Tests:**
- Test subscription establishes connection
- Test INSERT event triggers commit list update
- Test UPDATE event triggers job progress update
- Test reconnection after disconnect

---

#### Objective 6.2: Hypothesis Branching
**Description:** Create and compare alternative investigation branches.

**Deliverables:**
- [ ] `POST /v1/cases/{caseId}/branches` - Create branch
- [ ] Branch selector UI
- [ ] Constraint override in branch
- [ ] A/B comparison view

**Unit Tests:**
- Test branch creation from commit
- Test branch-specific commits isolation
- Test constraint override applies to reasoning
- Test comparison renders both trajectories

---

#### Objective 6.3: Export Report
**Description:** Generate investigation summary report.

**Deliverables:**
- [ ] `POST /v1/cases/{caseId}/export` - Trigger export job
- [ ] HTML report template
- [ ] Scene screenshots capture
- [ ] Evidence list compilation
- [ ] Trajectory explanation export
- [ ] Download link generation

**Unit Tests:**
- Test export job creation
- Test HTML template renders all sections
- Test evidence list includes all cards
- Test trajectory explanations format correctly
- Test download URL is valid and accessible

---

#### Objective 6.4: Full Integration Flow
**Description:** End-to-end smoke tests for critical paths.

**E2E Tests:**
- [ ] Upload scan → Reconstruction → SceneGraph display
- [ ] Submit witness statement → Profile update → Portrait generation
- [ ] Trigger reasoning → Trajectory display → Explain mode interaction
- [ ] Create branch → Modify constraint → Compare results
- [ ] Export report → Download → Verify content

---

## 4. API Response Formats

### Success Response
```json
{
  "success": true,
  "data": { ... },
  "meta": { "cursor": "...", "total": 100 }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Description of error",
    "details": { ... }
  }
}
```

### Error Codes
| Code | HTTP Status | Description |
|------|-------------|-------------|
| INVALID_REQUEST | 400 | Malformed request |
| UNAUTHORIZED | 401 | Not authenticated |
| FORBIDDEN | 403 | No permission |
| NOT_FOUND | 404 | Resource doesn't exist |
| CONFLICT | 409 | Duplicate resource |
| RATE_LIMITED | 429 | Too many requests |
| JOB_FAILED | 500 | Job execution failed |
| MODEL_UNAVAILABLE | 503 | AI model service down |
| INTERNAL_ERROR | 500 | Unexpected server error |

---

## 5. Testing Strategy

### Unit Test Coverage Targets
- **Data Layer:** 90%+ coverage on validation, serialization
- **API Layer:** 85%+ coverage on handlers, middleware
- **Worker Layer:** 80%+ coverage with mocked AI clients
- **Frontend Stores:** 90%+ coverage on state logic

### Integration Test Focus
- API → Database round-trips
- Worker → Job lifecycle
- Real-time subscription delivery

### E2E Test Focus
- Critical user flows (upload → reasoning)
- Error recovery paths
- Export functionality

### Mocking Strategy
All AI service clients implement interfaces:
```go
type ReconstructionClient interface {
    Reconstruct(ctx context.Context, input ReconstructionInput) (ReconstructionOutput, error)
}

type ImageGenClient interface {
    Generate(ctx context.Context, input ImageGenInput) (ImageGenOutput, error)
}

type ReasoningClient interface {
    Reason(ctx context.Context, input ReasoningInput) (ReasoningOutput, error)
}
```

Mock implementations return deterministic outputs for testing.

---

## 6. MVP Checklist

### P0 (Must Have for Demo)
- [ ] Phase 1: All objectives (Data Layer)
- [ ] Phase 2: Objectives 2.1, 2.2, 2.3 (Core APIs)
- [ ] Phase 3: Objectives 3.1, 3.2 (Worker Infrastructure)
- [ ] Phase 4: Objectives 4.1, 4.4 (Reconstruction + Reasoning)
- [ ] Phase 5: Objectives 5.1, 5.2, 5.3, 5.4 (Frontend Core)
- [ ] Phase 6: Objective 6.1 (Real-time)

### P1 (Strong Bonus)
- [ ] Phase 2: Objective 2.4 (Witness Statements)
- [ ] Phase 4: Objectives 4.2, 4.3 (Profile + ImageGen)
- [ ] Phase 6: Objectives 6.2, 6.3 (Branching + Export)

---

## 7. File Structure (Proposed)

```
SherlockOS/
├── backend/
│   ├── cmd/
│   │   └── server/main.go
│   ├── internal/
│   │   ├── api/           # HTTP handlers
│   │   ├── db/            # Database access
│   │   ├── models/        # Data structures
│   │   ├── queue/         # Job queue
│   │   ├── workers/       # AI workers
│   │   └── clients/       # External API clients
│   ├── migrations/        # SQL migrations
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── app/           # Next.js pages
│   │   ├── components/    # React components
│   │   ├── stores/        # Zustand stores
│   │   ├── hooks/         # Custom hooks
│   │   ├── lib/           # Utilities
│   │   └── types/         # TypeScript types
│   └── package.json
├── PRD.md
├── SherlockOS.md
└── README.md
```

---

*Document Version: 1.0 | Created: 2026-02-01*
