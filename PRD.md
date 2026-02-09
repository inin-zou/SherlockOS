# SherlockOS - Product Requirements Document (PRD)

> **Version:** 2.0
> **Last Updated:** 2026-02-09
> **Status:** Draft for Hackathon MVP

---

## 1. Executive Summary

SherlockOS is a detective assistance system built on **"World Model + Evidence Reliability Hierarchy"** architecture. It maps evidence from multiple sources (Tier 0–3) onto a unified spatial-temporal canvas, leverages AI to discover **Spatio-Temporal Paradoxes**, and reconstructs the logical truth of a case.

**Core Principles:**

1. **Reliability Hierarchy** — Evidence is auto-weighted by objectivity, from immutable physical boundaries (Tier 0) to subjective testimony (Tier 3).
2. **Spatial Anchoring** — Every piece of evidence owns coordinates or an observation angle in 3D/2.5D space.
3. **Spatio-Temporal Deduction** — Hard Anchors (Tier 0–1) correct fuzzy Soft Events (Tier 2–3); contradictions surface as **Paradox Alerts**.

---

## 2. Evidence Reliability Hierarchy

| Tier | Name | Definition | Typical Input | Weight |
|------|------|------------|---------------|--------|
| **Tier 0** | Environment | Physical infrastructure / impassable boundaries | Floor plans, static scans, 3D reconstruction | 100% / Physical base |
| **Tier 1** | Ground Truth | Raw visual/audio recordings | CCTV, dash-cam, key-frame video | High / Hard anchors |
| **Tier 2** | Electronic Logs | Digitally-triggered records | Smart-lock logs, Wi-Fi connections, sensors | Medium / Real but ambiguous |
| **Tier 3** | Testimonials | Subjective descriptions | Witness statements, suspect explanations | Low / Subjective probability |

---

## 3. Tech Stack Summary

| Layer | Technology |
|-------|------------|
| Backend | Go 1.22+ (chi router, pgx) |
| Frontend | Next.js, TypeScript, three.js, Zustand, Tailwind |
| Database | Supabase (Postgres + Storage + Realtime) |
| Queue | Redis Streams (hackathon) |
| Core AI Engine | Gemini 3 Pro (1M context, Thought Signatures) |
| Scene Reconstruction | Scene Reconstruction Engine (Proxy Geometry) |
| Image Generation | Nano Banana Pro (Localized Paint-to-Edit) |
| Future 4D Tracking | D4RT (Dynamic 4D Reconstruction & Tracking) |

---

## 4. The Sherlock Pipeline

```
Ingestion → Model → Deduction → Simulation
```

### Stage 1: Ingestion (Evidence Intake & Spatial Conversion)
- Ingest multi-modal evidence and auto-assign Tier (0–3) weights
- **Spatial Extraction**: Extract spatial trajectories from Tier 1 (D4RT/vision) and Tier 3 (Gemini infers motion paths from testimony)

### Stage 2: Model (World Building)
- Reconstruct static physical environment via Scene Reconstruction Engine
- Generate **Proxy Geometry** (Box/Cylinder simplifications) for spatial logic
- Establish spatial coordinate system and "hard impassable" boundaries

### Stage 3: Deduction (Logical Reasoning & Paradox Detection)
- **Perspective Validation**: Simulate witness POV in 3D, verify line-of-sight claims against Proxy Geometry occlusion
- **Discrepancy Identification**: Compare Tier 3 against Tier 0/1/2, highlight illogical **Spatio-Temporal Paradox** nodes
  - **Sightline Paradox**: Witness claims to see object X, but Tier 0 wall blocks line of sight
  - **Temporal Paradox**: Suspect claims location A, but Tier 2 logs show device at location B

### Stage 4: Simulation (Truth Reconstruction)
- Combine verified hard facts with calibrated soft evidence
- Run 4D dynamic panoramic simulation
- Generate "best-fit" case reconstruction

---

## 5. View Modes

| Mode | Definition | Core Function |
|------|------------|---------------|
| **Evidence Mode** | Static mapping & labeling | Display Evidence Archive. View reconstructed physical scene (Tier 0) with synchronized physical evidence assets. |
| **Reasoning Mode** | Dynamic deduction & comparison | **Multi-track Timeline** organized by Distinct Scene Volumes and Stakeholders. Click time-block to trigger cross-scene **Motion Path Ghosts**. |
| **Simulation Mode** | 4D panoramic reconstruction | **Truth Replay**. Drag timeline for 4D simulation with auto-detected **Paradox Alerts** highlighted. |

---

## 6. Development Phases

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
- [ ] `evidence_items` table with tier classification (0–3)

**Unit Tests:**
- Test enum validation for all types
- Test foreign key constraints
- Test unique constraints (idempotency_key, branch name per case)
- Test index performance on common queries
- Test tier assignment validation (0–3 range)

---

#### Objective 1.2: SceneGraph & Evidence Data Structures (Go)
**Description:** Implement Go structs for SceneGraph, Evidence Tiers, and related types.

**Deliverables:**
- [ ] `SceneGraph` struct with JSON marshaling
- [ ] `SceneObject` struct with ObjectType enum
- [ ] `Pose`, `BoundingBox` structs (Proxy Geometry support)
- [ ] `EvidenceCard`, `EvidenceSource` structs with Tier classification
- [ ] `Constraint`, `UncertaintyRegion` structs
- [ ] `Paradox` struct (Sightline / Temporal types)
- [ ] Validation functions for all structs

**Unit Tests:**
- Test JSON serialization/deserialization round-trip
- Test validation (required fields, value ranges)
- Test ObjectType enum parsing
- Test confidence value bounds (0-1)
- Test BoundingBox validity (min < max)
- Test Tier classification assignment logic

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
**Description:** Handle file uploads via presigned URLs with automatic Tier classification.

**Deliverables:**
- [ ] `POST /v1/cases/{caseId}/upload-intent` - Generate presigned URLs
- [ ] Automatic Tier classification based on file type
- [ ] Presigned URL generation with expiry
- [ ] Asset record creation after successful upload
- [ ] Storage key path convention: `cases/{caseId}/{kind}/{batchId}/{filename}`

**Tier Classification Rules:**
- Tier 0: `.pdf`, `.e57`, `.dwg`, `.obj` (floor plans, scans)
- Tier 1: `.mp4`, `.jpg`, `.png`, `.wav` (recordings)
- Tier 2: `.json`, `.csv`, `.log` (electronic logs)
- Tier 3: `.txt`, `.md`, `.docx` (testimonials)

**Unit Tests:**
- Test presigned URL generation
- Test automatic tier classification by file extension
- Test batch upload intent (multiple files)
- Test file size/type validation
- Test storage key format correctness

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
**Description:** Submit witness testimony (Tier 3) and trigger profile updates.

**Deliverables:**
- [ ] `POST /v1/cases/{caseId}/witness-statements` - Submit statements
- [ ] Create `witness_statement` commit with Tier 3 classification
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
**Goal:** Implement AI-powered workers using the Sherlock Pipeline.

#### Objective 4.1: Reconstruction Worker (Scene Reconstruction Engine)
**Description:** Process scan images to generate/update SceneGraph with Proxy Geometry.

**Deliverables:**
- [ ] `ReconstructionInput` validation
- [ ] Scene Reconstruction API client (with interface for mocking)
- [ ] `ReconstructionOutput` to SceneGraph conversion with Proxy Geometry
- [ ] Commit creation with `reconstruction_update` type
- [ ] Tier 0 boundary extraction (impassable walls, doors)
- [ ] Fallback to mock SceneGraph on failure

**Unit Tests:**
- Test input validation (required fields, asset keys exist)
- Test output parsing from mock API response
- Test SceneGraph merge logic (create/update/remove objects)
- Test Proxy Geometry generation (BoundingBox → Box/Cylinder)
- Test commit payload contains correct diff
- Test fallback produces valid mock SceneGraph

---

#### Objective 4.2: Profile Worker
**Description:** Extract structured attributes from witness statements (Tier 3).

**Deliverables:**
- [ ] Statement parsing with Gemini 3 Pro
- [ ] Attribute extraction (age, height, build, etc.)
- [ ] Confidence aggregation from multiple sources (weighted by Tier)
- [ ] Conflict detection between statements
- [ ] Trigger ImageGen job for portrait

**Unit Tests:**
- Test attribute extraction from sample statements
- Test confidence weighted by source credibility and Tier
- Test conflict detection (contradicting descriptions)
- Test ImageGen job triggered after profile update
- Test incremental attribute update (merge with existing)

---

#### Objective 4.3: ImageGen Worker (Nano Banana Pro)
**Description:** Generate suspect portraits with Localized Paint-to-Edit.

**Deliverables:**
- [ ] `ImageGenInput` validation
- [ ] Nano Banana Pro API client (with interface for mocking)
- [ ] Localized Paint-to-Edit for precise portrait modifications
- [ ] Image upload to Supabase Storage
- [ ] Thumbnail generation

**Unit Tests:**
- Test input validation per gen_type
- Test portrait prompt construction from attributes
- Test localized edit regions map to correct attributes
- Test asset storage key generation
- Test cost calculation correctness

---

#### Objective 4.4: Reasoning Worker (Gemini 3 Pro — Deduction Stage)
**Description:** Perform Spatio-Temporal Deduction: Perspective Validation, Discrepancy Detection, and Paradox Alerts.

**Deliverables:**
- [ ] `ReasoningInput` validation (SceneGraph + all Tier evidence)
- [ ] Gemini 3 Pro API client with Thought Signatures config
- [ ] **Perspective Validation**: Simulate witness POV, check line-of-sight occlusion against Proxy Geometry
- [ ] **Discrepancy Detection**: Cross-reference Tier 3 vs Tier 0/1/2
- [ ] **Paradox generation**: Sightline Paradox & Temporal Paradox structs
- [ ] Hard Anchor Mapping: anchor fuzzy Tier 3 events to Tier 1 timestamps
- [ ] Prompt template rendering with Reliability Hierarchy context
- [ ] Commit creation with `reasoning_result` type
- [ ] WebSocket streaming for thinking chunks

**Unit Tests:**
- Test input validation (SceneGraph required)
- Test Thought Signatures configuration
- Test Perspective Validation detects known occlusion cases
- Test Sightline Paradox generation when wall blocks line-of-sight
- Test Temporal Paradox generation when logs contradict testimony
- Test Hard Anchor Mapping calibrates fuzzy timestamps
- Test prompt template rendering with tiered evidence
- Test streaming chunk delivery

---

### Phase 5: Frontend Core
**Goal:** Build core UI with three view modes aligned to the Sherlock Pipeline.

#### Objective 5.1: State Management (Zustand)
**Description:** Implement global state stores for app data.

**Deliverables:**
- [ ] `useCaseStore` - Current case, commits, snapshot
- [ ] `useJobStore` - Active jobs, progress tracking
- [ ] `useViewModeStore` - Evidence/Reasoning/Simulation mode toggle
- [ ] `useSelectionStore` - Selected objects, commits, trajectory segments
- [ ] `useEvidenceStore` - Evidence items organized by Tier (0–3)

**Unit Tests:**
- Test store initialization with default values
- Test state updates (add commit, update job progress)
- Test derived state (filtered commits, active jobs count)
- Test store reset on case change
- Test evidence filtering by Tier

---

#### Objective 5.2: Evidence Mode — Timeline & Archive
**Description:** Display commit history, Evidence Archive organized by Tier.

**Deliverables:**
- [ ] Commit list with type icons
- [ ] Evidence Archive sidebar grouped by Tier (0–3)
- [ ] Commit summary and timestamp display
- [ ] Diff view (objects added/updated/removed)
- [ ] Click-to-select for state replay
- [ ] Auto-scroll on new commit

**Unit Tests:**
- Test commit list renders correct count
- Test type icon mapping
- Test Evidence Archive groups by Tier correctly
- Test diff calculation display
- Test selection state updates on click
- Test real-time subscription updates list

---

#### Objective 5.3: Scene View (three.js) with Proxy Geometry
**Description:** 3D visualization of SceneGraph with Proxy Geometry and Evidence Markers.

**Deliverables:**
- [ ] Scene setup (camera, lights, controls)
- [ ] Object rendering from SceneGraph with Proxy Geometry (Box/Cylinder)
- [ ] Object type-based materials/colors
- [ ] Evidence markers with Tier-based styling
- [ ] Hover highlight with source commit info
- [ ] Click-to-select object → Evidence card in right panel
- [ ] Uncertainty region overlay (semi-transparent amber)

**Unit Tests:**
- Test scene initializes without errors
- Test objects render at correct positions
- Test object count matches SceneGraph
- Test selection state sync with store
- Test camera bounds fit scene bounds

---

#### Objective 5.4: Reasoning Mode — Multi-track Timeline & Paradox Display
**Description:** Display Reasoning results: motion path ghosts, paradox alerts, discrepancy highlights.

**Deliverables:**
- [ ] Multi-track Timeline organized by Scene Volumes and Stakeholders
- [ ] Motion Path Ghost visualization (trajectory overlays in 3D)
- [ ] Click time-block → trigger cross-scene ghost animation
- [ ] Paradox Alert markers (Sightline / Temporal)
- [ ] Discrepancy highlight nodes in scene + timeline
- [ ] Evidence reference links

**Unit Tests:**
- Test multi-track timeline renders all tracks
- Test paradox alert markers display correctly
- Test time-block click triggers ghost animation
- Test discrepancy nodes highlighted in both scene and timeline
- Test evidence ref links navigate correctly

---

#### Objective 5.5: Simulation Mode — 4D Replay
**Description:** 4D panoramic truth replay with Paradox Alerts.

**Deliverables:**
- [ ] Timeline scrubber for 4D replay
- [ ] Play/pause/speed controls
- [ ] Auto-detected Paradox Alerts highlighted during playback
- [ ] Camera follow mode for stakeholders
- [ ] POV simulation (witness perspective camera)

**Unit Tests:**
- Test playback controls work correctly
- Test paradox alerts appear at correct timestamps
- Test camera follow tracks selected stakeholder
- Test POV camera renders from witness position

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
- [ ] Evidence list compilation (organized by Tier)
- [ ] Paradox summary with evidence references
- [ ] Trajectory explanation export
- [ ] Download link generation

**Unit Tests:**
- Test export job creation
- Test HTML template renders all sections
- Test evidence list organized by Tier
- Test paradox summary includes all detected paradoxes
- Test download URL is valid and accessible

---

#### Objective 6.4: Full Integration Flow
**Description:** End-to-end smoke tests for critical paths.

**E2E Tests:**
- [ ] Upload Tier 0 scan → Reconstruction → Proxy Geometry → SceneGraph display
- [ ] Upload Tier 1 CCTV → Hard Anchor extraction → Timeline events
- [ ] Submit Tier 3 witness statement → Profile update → Portrait generation
- [ ] Trigger Deduction → Paradox detection → Discrepancy highlights
- [ ] Simulation Mode → 4D replay with Paradox Alerts
- [ ] Create branch → Modify constraint → Compare results
- [ ] Export report → Download → Verify content

---

## 7. API Response Formats

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

## 8. Testing Strategy

### Unit Test Coverage Targets
- **Data Layer:** 90%+ coverage on validation, serialization
- **API Layer:** 85%+ coverage on handlers, middleware
- **Worker Layer:** 80%+ coverage with mocked AI clients
- **Frontend Stores:** 90%+ coverage on state logic

### Integration Test Focus
- API → Database round-trips
- Worker → Job lifecycle
- Real-time subscription delivery
- Evidence Tier classification accuracy

### E2E Test Focus
- Critical user flows (upload → deduction → paradox detection)
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

## 9. MVP Checklist

### P0 (Must Have for Demo)
- [ ] Phase 1: All objectives (Data Layer with Tier classification)
- [ ] Phase 2: Objectives 2.1, 2.2, 2.3 (Core APIs with Tier-aware uploads)
- [ ] Phase 3: Objectives 3.1, 3.2 (Worker Infrastructure)
- [ ] Phase 4: Objectives 4.1, 4.4 (Reconstruction + Deduction/Paradox Detection)
- [ ] Phase 5: Objectives 5.1, 5.2, 5.3, 5.4 (Frontend: Evidence + Reasoning modes)
- [ ] Phase 6: Objective 6.1 (Real-time)

### P1 (Strong Bonus)
- [ ] Phase 2: Objective 2.4 (Witness Statements)
- [ ] Phase 4: Objectives 4.2, 4.3 (Profile + ImageGen)
- [ ] Phase 5: Objective 5.5 (Simulation Mode / 4D Replay)
- [ ] Phase 6: Objectives 6.2, 6.3 (Branching + Export)

---

## 10. File Structure (Proposed)

```
SherlockOS/
├── backend/
│   ├── cmd/
│   │   └── server/main.go
│   ├── internal/
│   │   ├── api/           # HTTP handlers
│   │   ├── db/            # Database access
│   │   ├── models/        # Data structures (incl. Tiers, Paradox)
│   │   ├── queue/         # Job queue
│   │   ├── workers/       # AI workers (Sherlock Pipeline stages)
│   │   └── clients/       # External API clients (Gemini 3 Pro, etc.)
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

*Document Version: 2.0 | Created: 2026-02-01 | Updated: 2026-02-09*
