# SherlockOS

**AI-powered Spatio-Temporal Deduction Engine for Crime Scene Reconstruction**

> *"SherlockOS doesn't tell you who did it. It shows you what is impossible, so only the truth remains."*

Built for the **Google Gemini 3 Hackathon 2026**.

---

## What is SherlockOS?

SherlockOS is a detective assistance system that takes messy, multimodal evidence and reconciles it into a single, structured 3D world model. It maps evidence from multiple sources onto a unified spatial-temporal canvas, leverages AI to discover **Spatio-Temporal Paradoxes**, and reconstructs the logical truth of a case.

### Core Innovation: Reliability Hierarchy

| Tier | Name | Definition | Example |
|------|------|------------|---------|
| **Tier 0** | Environment | Physical infrastructure / impassable boundaries | Floor plans, 3D reconstruction |
| **Tier 1** | Ground Truth | Raw visual/audio recordings | CCTV, dash-cam footage |
| **Tier 2** | Electronic Logs | Digitally-triggered records | Smart-lock logs, Wi-Fi connections |
| **Tier 3** | Testimonials | Subjective descriptions | Witness statements |

Hard Anchors (Tier 0-1) correct fuzzy Soft Events (Tier 2-3); contradictions surface as **Paradox Alerts**.

---

## The Sherlock Pipeline

```
Ingestion → Model → Deduction → Simulation
```

1. **Ingestion** — Upload multi-modal evidence (images, video, witness statements, sensor logs). Auto-classify into reliability tiers.
2. **Model** — Reconstruct the physical environment in 3D via Gaussian splatting. Generate proxy geometry for spatial reasoning.
3. **Deduction** — AI plots motion paths, simulates perspectives, and flags contradictions where testimony breaks the laws of physics.
4. **Simulation** — Replay the most probable version of events in 3D, combining all validated evidence.

---

## AI Services

| Service | Provider | Purpose |
|---------|----------|---------|
| **Reasoning** | Gemini 2.5 Flash | Trajectory hypothesis generation with thinking |
| **Profile Extraction** | Gemini 2.5 Flash | Extract suspect attributes from witness statements |
| **Image Generation** | Gemini (Nano Banana) | Suspect portraits, POV scenes, evidence boards |
| **Portrait Chat** | Gemini (Nano Banana) | Multi-turn iterative portrait refinement |
| **3D Reconstruction** | Modal (HunyuanWorld-Mirror) | Gaussian splatting from images/video |
| **Video Replay** | Modal (HY-World-1.5) | Camera trajectory video generation |
| **Scene Analysis** | Gemini 3 Pro Vision | Object detection from crime scene images |
| **3D Assets** | Replicate (Hunyuan3D-2) | Evidence 3D model generation |

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| **Frontend** | Next.js 16, TypeScript, React 19, three.js, Zustand, Tailwind CSS |
| **Backend** | Go 1.24, chi router, pgx |
| **Database** | Supabase (Postgres + Storage + Realtime) |
| **Queue** | Redis Streams (in-memory fallback) |
| **3D Rendering** | Gaussian splatting via @react-three/drei |
| **Infrastructure** | Modal.com (GPU compute), Replicate |

---

## Project Structure

```
SherlockOS/
├── frontend/          # Next.js frontend application
├── backend/           # Go backend service
├── modal/             # Modal.com 3D reconstruction deployment
├── scripts/           # Utility scripts
├── demo/              # Demo data and assets
├── PRD.md             # Product Requirements Document
└── README.md
```

---

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 18+
- PostgreSQL (via Supabase)
- Redis (optional)

### 1. Backend

```bash
cd backend
cp .env.example .env
# Edit .env with your API keys (GEMINI_API_KEY, SUPABASE_URL, etc.)
make deps
make run
```

### 2. Frontend

```bash
cd frontend
npm install
npm run dev
```

The app will be available at `http://localhost:3000` with the backend API at `http://localhost:8080`.

### Environment Variables

| Variable | Description |
|----------|-------------|
| `GEMINI_API_KEY` | Google Gemini API key |
| `SUPABASE_URL` | Supabase project URL |
| `SUPABASE_ANON_KEY` | Supabase anonymous key |
| `SUPABASE_SECRET_KEY` | Supabase service role key |
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Redis connection URL (optional) |
| `MODAL_MIRROR_URL` | Modal HunyuanWorld-Mirror endpoint |
| `MODAL_WORLDPLAY_URL` | Modal HY-World-1.5 endpoint |
| `REPLICATE_API_TOKEN` | Replicate API token |

---

## Key Features

- **Multi-modal evidence upload** with automatic tier classification
- **3D crime scene reconstruction** via Gaussian splatting
- **Suspect portrait generation** with multi-turn conversational refinement
- **Spatio-temporal paradox detection** — find where testimony contradicts physics
- **Timeline-based investigation versioning** with branching hypotheses
- **POV simulation** — validate witness perspectives against the 3D model
- **Evidence board generation** with AI-powered reasoning

---

## License

This project was built for the **Google Gemini 3 Hackathon 2026**. It is provided as-is for demonstration and educational purposes under the [MIT License](LICENSE).
