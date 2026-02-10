# SherlockOS Frontend Design Document

> **Version:** 1.0 | **Last Updated:** 2026-02

## Overview

SherlockOS is a crime investigation assistant built on **World Model + Timeline + Explainable Reasoning**. This document maps the PRD features to concrete UI components and interactions.

---

## Design System

### Color Palette

| Role | Hex | Usage |
|------|-----|-------|
| Background Primary | `#111114` | Main app background |
| Background Secondary | `#1f1f24` | Cards, panels |
| Background Tertiary | `#2a2a32` | Hover states, borders |
| Background Dark | `#0a0a0c` | Scene viewer, deep panels |
| Text Primary | `#ffffff` | Headings, important text |
| Text Secondary | `#8b8b96` | Body text, descriptions |
| Text Muted | `#606068` | Timestamps, metadata |
| Accent Blue | `#3b82f6` | Primary actions, selection |
| Accent Purple | `#8b5cf6` | Reasoning, trajectories |
| Accent Green | `#22c55e` | Success, confirmed evidence |
| Accent Amber | `#f59e0b` | Warnings, uncertainty |
| Accent Red | `#ef4444` | Errors, discrepancies |

### Typography

| Element | Style |
|---------|-------|
| H1 | 24px, semibold, white |
| H2 | 18px, semibold, white |
| H3 | 14px, semibold, white |
| Body | 14px, regular, `#8b8b96` |
| Caption | 12px, regular, `#606068` |
| Code | 13px, monospace, `#8b8b96` |

### Spacing

- Base unit: 4px
- Common values: 4, 8, 12, 16, 24, 32, 48px

### Components

- Borders: 1px `#2a2a32`
- Border radius: 6px (small), 8px (medium), 12px (large)
- Shadows: Minimal, use borders instead
- Transitions: 150-200ms ease

---

## Application Layout

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  HEADER (56px)                                                              │
│  Logo | Case Title | [Evidence] [Simulation] [Reasoning] | Jobs | Export   │
├──────────────┬──────────────────────────────────────────────┬───────────────┤
│   SIDEBAR    │        MAIN SCENE AREA                       │  RIGHT PANEL  │
│   (280px)    │                                              │   (320px)     │
│              │  ┌────────────────────────────────────────┐  │               │
│  Evidence    │  │                                        │  │   Mode-       │
│  Archive     │  │          3D Scene Viewer               │  │   Specific    │
│              │  │                                        │  │   Content     │
│  ─────────   │  │    Objects + Trajectories +            │  │               │
│  Tier 0      │  │    Evidence Markers +                  │  │  ──────────   │
│  Tier 1      │  │    Uncertainty Regions                 │  │  Evidence:    │
│  Tier 2      │  │                                        │  │   Details     │
│  Tier 3      │  └────────────────────────────────────────┘  │               │
│              │                                              │  Simulation:  │
│  ─────────   │  ┌────────────────────────────────────────┐  │   Controls    │
│  Upload      │  │        MODE SELECTOR BAR               │  │               │
│  Drop Zone   │  │   [Evidence] [Simulation] [Reasoning]  │  │  Reasoning:   │
│              │  └────────────────────────────────────────┘  │   Trajectories│
│              ├──────────────────────────────────────────────┴───────────────┤
│              │  TIMELINE (200px)                                            │
│              │  ┌─ Commits: [upload] [witness] [reasoning] [profile] ────┐  │
│              │  ├─ Tracks:  ████░░░░████░░████ (events with markers) ────┤  │
│              │  └─ Time Scrubber ─────────────────────────────────────────┘  │
└──────────────┴──────────────────────────────────────────────────────────────┘
```

---

## Feature Mapping: PRD → UI Components

### 1. Crime Scene World State (PRD 5.2)

**PRD Feature:** Upload multi-angle photos → reconstruct SceneGraph → interactive 3D view

| PRD Requirement | UI Component | Location | Behavior |
|-----------------|--------------|----------|----------|
| Upload scan images | `DropZone` | Sidebar bottom | Drag & drop files, show upload progress |
| Tier classification | `FileClassifier` | Automatic | Detect file type → assign Tier 0-3 |
| Reconstruction job | `JobProgress` | Header badge + panel | Show progress %, estimated time |
| SceneGraph rendering | `SceneViewer` | Main area | Three.js objects from API snapshot |
| Object interaction | `ObjectTooltip` | On hover | Show label, confidence, source commits |
| Evidence markers | `EvidenceMarker` | 3D scene | Numbered markers (1, 2, 3...) |
| Uncertainty regions | `UncertaintyOverlay` | 3D scene | Semi-transparent amber zones |

**User Flow:**
1. Drag files to sidebar → Files classified by type
2. Upload progress shown → Job created automatically
3. SceneGraph updates → 3D view refreshes
4. Click objects → See evidence details in right panel

### 2. Timeline (PRD 5.1)

**PRD Feature:** Version-controlled event stream with commits

| PRD Requirement | UI Component | Location | Behavior |
|-----------------|--------------|----------|----------|
| Commit list | `CommitList` | Timeline panel | Scrollable list with type icons |
| Commit types | `CommitIcon` | Each commit | Color-coded: upload, witness, reasoning |
| Commit diff | `CommitDiff` | On expand | Show changes: added/updated/removed |
| Time navigation | `TimeScrubber` | Timeline bottom | Drag to view historical state |
| Playback | `PlaybackControls` | Timeline | Play/pause animation of changes |

**Commit Type Icons:**
- `upload_scan` → Camera icon (blue)
- `witness_statement` → User icon (purple)
- `reconstruction_update` → Cube icon (green)
- `profile_update` → Person icon (amber)
- `reasoning_result` → Brain icon (purple)
- `export_report` → Download icon (gray)

### 3. Suspect Profiling (PRD 5.6)

**PRD Feature:** Input witness statements → structured attributes → portrait generation

| PRD Requirement | UI Component | Location | Behavior |
|-----------------|--------------|----------|----------|
| Statement input | `WitnessForm` | Right panel (Evidence mode) | Text area + credibility slider |
| Attribute display | `AttributeList` | Right panel | Age, height, build with confidence bars |
| Confidence scores | `ConfidenceBar` | Each attribute | Visual bar 0-100% |
| Conflict indicators | `ConflictBadge` | Conflicting attributes | Red badge with source tooltip |
| Portrait display | `PortraitCard` | Right panel | Generated image with edit button |
| Multi-version | `PortraitCompare` | Modal | Side-by-side portraits |

**Attribute Structure:**
```
Age Range:      [25-35]  ████████░░ 80%   (Guard, Driver)
Height:         [175-185cm] ███████░░░ 70%   (Guard) ⚠️ Conflict: Neighbor
Build:          Slim     █████████░ 90%   (Guard)
Hair:           Short, Dark ███████░░░ 70%   (Driver)
Facial Hair:    Beard    ██████░░░░ 60%   (Driver)
```

### 4. Trajectory Reasoning (PRD 5.3, 9.5)

**PRD Feature:** Gemini 2.5 Flash analyzes scene → generates movement hypotheses

| PRD Requirement | UI Component | Location | Behavior |
|-----------------|--------------|----------|----------|
| Trigger reasoning | `ReasoningButton` | Right panel | "Analyze Trajectories" |
| Trajectory display | `TrajectoryPath` | 3D scene | Animated path with waypoints |
| Trajectory list | `TrajectoryList` | Right panel | Ranked list with confidence |
| Segment click | `TrajectorySegment` | 3D scene | Highlight related evidence |
| Evidence refs | `EvidenceRefList` | On segment click | List of supporting evidence |
| Explanation | `ExplanationCard` | Right panel | Natural language summary |
| Thinking summary | `ThinkingSummary` | Collapsible | Model's reasoning process |

**Trajectory Visualization:**
```
Entry Point (Window) ──────> Desk Area ──────> Filing Cabinet ──────> Exit (Door)
     [1]                        [2]                 [3]                [4]
   t=23:30                   t=23:32             t=23:35            t=23:40
  Conf: 92%                 Conf: 85%           Conf: 78%          Conf: 88%
```

### 5. View Modes (PRD 5.3)

| Mode | Description | Right Panel Content | Scene Overlay |
|------|-------------|---------------------|---------------|
| **Evidence** | Default mode for evidence collection | Suspect profile, evidence details | Object highlights, evidence markers |
| **Simulation** | Generate and test movement paths | Motion input, POV controls | Trajectory animation, camera follow |
| **Reasoning** | Analyze and compare hypotheses | Trajectory list, explanations | Multiple paths, discrepancy highlights |

### 6. Explain Mode (PRD 5.4)

**PRD Feature:** Interactive reasoning trace - click segment → see evidence chain

| Interaction | Result |
|-------------|--------|
| Click trajectory segment | Highlight path, pulse related objects |
| Hover evidence card | Show contributing commits in timeline |
| Click "Why?" button | Expand explanation with evidence refs |
| Click evidence ref | Jump to commit in timeline |

### 7. Hypothesis Branches (PRD 5.5)

| PRD Requirement | UI Component | Location | Behavior |
|-----------------|--------------|----------|----------|
| Create branch | `BranchButton` | Timeline header | Fork from selected commit |
| Branch selector | `BranchDropdown` | Timeline header | Switch between branches |
| Branch compare | `BranchCompare` | Modal | Side-by-side trajectory comparison |
| Merge branch | `MergeBranch` | Branch menu | Merge conclusions to main |

### 8. Export Report (PRD 5.8)

| PRD Requirement | UI Component | Location | Behavior |
|-----------------|--------------|----------|----------|
| Export button | `ExportButton` | Header | Trigger export job |
| Format selection | `ExportOptions` | Dropdown | HTML (default), PDF (future) |
| Progress | `ExportProgress` | Modal | Show generation progress |
| Download | `DownloadLink` | Modal | Link to generated report |

**Report Contents:**
- Scene screenshots (top-down, key areas)
- Evidence list with summaries
- Trajectory hypotheses with explanations
- Suspect profile and portrait
- Uncertainty areas
- Next-step suggestions

---

## Page Routes

| Route | Page | Description |
|-------|------|-------------|
| `/` | Dashboard | Auto-loads first case or creates demo |
| `/cases` | Cases List | View and manage all cases |
| `/cases/[id]` | Case Detail | Full investigation workspace |

---

## Component Hierarchy

```
App
├── Header
│   ├── Logo
│   ├── CaseTitle
│   ├── ModeSelector
│   ├── JobsBadge
│   └── ExportButton
├── Sidebar
│   ├── EvidenceArchive
│   │   ├── TierSection (×4)
│   │   └── EvidenceItem
│   ├── UploadArea
│   │   └── DropZone
│   └── ResizeHandle
├── MainArea
│   ├── SceneViewer
│   │   ├── ThreeCanvas
│   │   ├── SceneObjects
│   │   ├── TrajectoryPaths
│   │   ├── EvidenceMarkers
│   │   └── UncertaintyOverlay
│   └── Timeline
│       ├── CommitList
│       ├── TrackView
│       └── TimeScrubber
├── RightPanel
│   ├── EvidencePanel (Evidence mode)
│   │   ├── SuspectProfile
│   │   ├── WitnessForm
│   │   └── EvidenceDetails
│   ├── SimulationPanel (Simulation mode)
│   │   ├── MotionInput
│   │   ├── VideoUpload
│   │   └── POVControls
│   └── ReasoningPanel (Reasoning mode)
│       ├── TrajectoryList
│       ├── ExplanationCard
│       └── NextStepSuggestions
└── Modals
    ├── JobProgress
    ├── BranchCompare
    ├── ExportProgress
    └── EvidenceDetail
```

---

## State Management (Zustand)

```typescript
interface AppState {
  // Case
  currentCase: Case | null;
  cases: Case[];

  // Scene
  sceneGraph: SceneGraph | null;
  selectedObjectIds: string[];

  // Timeline
  commits: Commit[];
  currentTime: number;
  isPlaying: boolean;

  // Jobs
  jobs: Job[];

  // Reasoning
  trajectories: Trajectory[];
  selectedTrajectoryId: string | null;

  // Profile
  suspectProfile: SuspectProfile | null;

  // UI
  viewMode: 'evidence' | 'simulation' | 'reasoning';
  sidebarWidth: number;
  timelineHeight: number;
}
```

---

## API Integration

### Endpoints Used

| Endpoint | Method | Usage |
|----------|--------|-------|
| `/v1/cases` | GET | List all cases |
| `/v1/cases` | POST | Create new case |
| `/v1/cases/{id}` | GET | Get case details |
| `/v1/cases/{id}/timeline` | GET | Get commits |
| `/v1/cases/{id}/snapshot` | GET | Get SceneGraph |
| `/v1/cases/{id}/upload-intent` | POST | Get presigned URLs |
| `/v1/cases/{id}/jobs` | POST | Create job |
| `/v1/cases/{id}/witness-statements` | POST | Add statements |
| `/v1/cases/{id}/reasoning` | POST | Trigger reasoning |
| `/v1/cases/{id}/export` | POST | Generate report |
| `/v1/jobs/{id}` | GET | Check job status |

### Real-time Updates

Via Supabase Realtime:
- Subscribe to `commits` table → Update timeline
- Subscribe to `jobs` table → Update progress

---

## Responsive Behavior

| Breakpoint | Layout Change |
|------------|---------------|
| Desktop (>1200px) | Full 3-column layout |
| Tablet (768-1200px) | Collapsible sidebar, narrower panels |
| Mobile (<768px) | Single column, bottom sheet panels |

---

## Accessibility

- All interactive elements keyboard navigable
- ARIA labels on icons and buttons
- Color contrast minimum 4.5:1
- Focus indicators visible
- Screen reader announcements for job status

---

## Performance

- Lazy load 3D viewer (dynamic import)
- Virtual scrolling for long commit lists
- Debounced API calls for real-time search
- Image optimization for evidence thumbnails
- Progressive loading for large scenes

---

## Testing Requirements

### Unit Tests
- All components with Jest + React Testing Library
- Store actions and selectors
- API utility functions

### Integration Tests
- Upload flow → Job creation → Scene update
- Witness input → Profile update → Portrait generation
- Reasoning trigger → Trajectory display

### E2E Tests (Playwright)
- Full investigation workflow
- Export report generation
- Branch comparison

---

## File Structure

```
frontend/
├── src/
│   ├── app/
│   │   ├── page.tsx              # Dashboard
│   │   ├── cases/
│   │   │   ├── page.tsx          # Cases list
│   │   │   └── [id]/
│   │   │       └── page.tsx      # Case detail
│   │   ├── layout.tsx
│   │   └── globals.css
│   ├── components/
│   │   ├── evidence/             # Evidence components
│   │   ├── jobs/                 # Job progress
│   │   ├── layout/               # Header, Sidebar
│   │   ├── panels/               # Right panel modes
│   │   ├── reasoning/            # Trajectory, explanations
│   │   ├── scene/                # 3D viewer
│   │   ├── simulation/           # Motion, POV
│   │   ├── timeline/             # Commits, tracks
│   │   └── ui/                   # Shared UI components
│   ├── hooks/
│   │   ├── useUpload.ts
│   │   ├── useJobs.ts
│   │   └── useRealtime.ts
│   ├── lib/
│   │   ├── api.ts                # API client
│   │   ├── store.ts              # Zustand store
│   │   └── types.ts              # TypeScript types
│   └── __tests__/                # Test files
├── public/
└── package.json
```

---

## Implementation Checklist

### Phase 1: Core Layout ✅
- [x] Header with mode selector
- [x] Sidebar with evidence archive
- [x] Main scene area
- [x] Timeline panel
- [x] Right panel structure

### Phase 2: Data Integration
- [x] Cases list page
- [x] Case detail page with routing
- [x] API client for all endpoints
- [ ] Supabase realtime subscriptions
- [ ] Job progress polling

### Phase 3: 3D Scene
- [ ] SceneGraph → Three.js objects
- [ ] Object selection and highlighting
- [ ] Evidence markers
- [ ] Trajectory path rendering
- [ ] Camera controls

### Phase 4: Features
- [ ] File upload with tier classification
- [ ] Witness statement form
- [ ] Suspect profile display
- [ ] Reasoning trigger and display
- [ ] Export functionality

### Phase 5: Polish
- [ ] Error boundaries
- [ ] Loading states
- [ ] Empty states
- [ ] Animations and transitions
- [ ] Mobile responsiveness

---

*Document maintained alongside SherlockOS.md PRD*
