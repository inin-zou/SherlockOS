-- SherlockOS Database Schema
-- Migration: 001_initial_schema
-- Description: Create initial database schema with all tables and indexes

-- ============================================
-- ENUM TYPES
-- ============================================

CREATE TYPE commit_type AS ENUM (
  'upload_scan',
  'witness_statement',
  'manual_edit',
  'reconstruction_update',
  'profile_update',
  'reasoning_result',
  'export_report'
);

CREATE TYPE job_type AS ENUM (
  'reconstruction',
  'imagegen',
  'reasoning',
  'profile',
  'export'
);

CREATE TYPE job_status AS ENUM (
  'queued',
  'running',
  'done',
  'failed',
  'canceled'
);

CREATE TYPE asset_kind AS ENUM (
  'scan_image',
  'generated_image',
  'mesh',
  'pointcloud',
  'portrait',
  'report'
);

-- ============================================
-- TABLES
-- ============================================

-- Cases table
CREATE TABLE cases (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  title           text NOT NULL,
  description     text,
  created_by      uuid,  -- Reference to auth.users (Supabase Auth)
  created_at      timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT cases_title_length CHECK (char_length(title) <= 200)
);

CREATE INDEX idx_cases_created_at ON cases(created_at DESC);

-- Branches table (must be created before commits due to FK)
CREATE TABLE branches (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id         uuid NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  name            text NOT NULL,
  base_commit_id  uuid,  -- Will add FK after commits table is created
  created_at      timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT branches_name_length CHECK (char_length(name) <= 100),
  CONSTRAINT branches_unique_name UNIQUE (case_id, name)
);

CREATE INDEX idx_branches_case_id ON branches(case_id);

-- Commits table (Timeline, append-only)
CREATE TABLE commits (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id           uuid NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  parent_commit_id  uuid REFERENCES commits(id),
  branch_id         uuid REFERENCES branches(id),
  type              commit_type NOT NULL,
  summary           text NOT NULL,
  payload           jsonb NOT NULL DEFAULT '{}',
  created_by        uuid,
  created_at        timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT commits_summary_length CHECK (char_length(summary) <= 500)
);

CREATE INDEX idx_commits_case_id ON commits(case_id, created_at DESC);
CREATE INDEX idx_commits_branch_id ON commits(branch_id) WHERE branch_id IS NOT NULL;
CREATE INDEX idx_commits_payload_job_id ON commits((payload->>'job_id')) WHERE payload->>'job_id' IS NOT NULL;

-- Add foreign key from branches to commits for base_commit_id
ALTER TABLE branches ADD CONSTRAINT branches_base_commit_fk
  FOREIGN KEY (base_commit_id) REFERENCES commits(id);

-- Scene snapshots table (Current State)
CREATE TABLE scene_snapshots (
  case_id     uuid PRIMARY KEY REFERENCES cases(id) ON DELETE CASCADE,
  commit_id   uuid NOT NULL REFERENCES commits(id),
  scenegraph  jsonb NOT NULL,
  updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Suspect profiles table
CREATE TABLE suspect_profiles (
  case_id             uuid PRIMARY KEY REFERENCES cases(id) ON DELETE CASCADE,
  commit_id           uuid NOT NULL REFERENCES commits(id),
  attributes          jsonb NOT NULL DEFAULT '{}',
  portrait_asset_key  text,
  updated_at          timestamptz NOT NULL DEFAULT now()
);

-- Jobs table
CREATE TABLE jobs (
  id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id          uuid NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  type             job_type NOT NULL,
  status           job_status NOT NULL DEFAULT 'queued',
  progress         integer NOT NULL DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
  input            jsonb NOT NULL,
  output           jsonb,
  error            text,
  idempotency_key  text,
  retry_count      integer NOT NULL DEFAULT 0,
  created_at       timestamptz NOT NULL DEFAULT now(),
  updated_at       timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT jobs_idempotency_unique UNIQUE (idempotency_key)
);

CREATE INDEX idx_jobs_case_id ON jobs(case_id, created_at DESC);
CREATE INDEX idx_jobs_status ON jobs(status) WHERE status IN ('queued', 'running');
CREATE INDEX idx_jobs_heartbeat ON jobs(updated_at) WHERE status = 'running';

-- Assets table
CREATE TABLE assets (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  case_id      uuid NOT NULL REFERENCES cases(id) ON DELETE CASCADE,
  kind         asset_kind NOT NULL,
  storage_key  text NOT NULL,
  metadata     jsonb DEFAULT '{}',
  created_at   timestamptz NOT NULL DEFAULT now(),

  CONSTRAINT assets_storage_key_unique UNIQUE (storage_key)
);

CREATE INDEX idx_assets_case_id ON assets(case_id);
CREATE INDEX idx_assets_kind ON assets(case_id, kind);

-- ============================================
-- FUNCTIONS
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for jobs table
CREATE TRIGGER jobs_updated_at
  BEFORE UPDATE ON jobs
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

-- Trigger for scene_snapshots table
CREATE TRIGGER scene_snapshots_updated_at
  BEFORE UPDATE ON scene_snapshots
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

-- Trigger for suspect_profiles table
CREATE TRIGGER suspect_profiles_updated_at
  BEFORE UPDATE ON suspect_profiles
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- COMMENTS
-- ============================================

COMMENT ON TABLE cases IS 'Crime investigation cases';
COMMENT ON TABLE commits IS 'Timeline entries (append-only) for case changes';
COMMENT ON TABLE branches IS 'Hypothesis branches for A/B comparison';
COMMENT ON TABLE scene_snapshots IS 'Current SceneGraph state for each case';
COMMENT ON TABLE suspect_profiles IS 'Structured suspect attributes and portrait reference';
COMMENT ON TABLE jobs IS 'Async processing jobs (reconstruction, reasoning, etc.)';
COMMENT ON TABLE assets IS 'References to files stored in Supabase Storage';
