import type {
  Case,
  Commit,
  Job,
  SceneGraph,
  ApiResponse,
  JobType,
} from './types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/v1';

class ApiError extends Error {
  code: string;
  details?: Record<string, unknown>;

  constructor(code: string, message: string, details?: Record<string, unknown>) {
    super(message);
    this.code = code;
    this.details = details;
    this.name = 'ApiError';
  }
}

async function request<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${API_BASE}${endpoint}`;

  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  const data: ApiResponse<T> = await response.json();

  if (!data.success || data.error) {
    throw new ApiError(
      data.error?.code || 'UNKNOWN_ERROR',
      data.error?.message || 'An unknown error occurred',
      data.error?.details
    );
  }

  return data.data as T;
}

// Cases
export async function getCases(): Promise<Case[]> {
  return request<Case[]>('/cases');
}

export async function getCase(caseId: string): Promise<Case> {
  return request<Case>(`/cases/${caseId}`);
}

export async function createCase(title: string, description?: string): Promise<Case> {
  return request<Case>('/cases', {
    method: 'POST',
    body: JSON.stringify({ title, description }),
  });
}

// Timeline
export async function getTimeline(
  caseId: string,
  cursor?: string,
  limit = 50
): Promise<{ commits: Commit[]; cursor?: string }> {
  const params = new URLSearchParams({ limit: String(limit) });
  if (cursor) params.set('cursor', cursor);

  // API returns array directly, wrap it for compatibility
  const commits = await request<Commit[]>(
    `/cases/${caseId}/timeline?${params}`
  );
  return { commits: commits || [] };
}

// Scene Snapshot
export async function getSnapshot(caseId: string): Promise<{
  case_id: string;
  commit_id: string;
  scenegraph: SceneGraph;
  updated_at: string;
}> {
  return request(`/cases/${caseId}/snapshot`);
}

// Upload Intent
export async function getUploadIntent(
  caseId: string,
  files: Array<{ filename: string; content_type: string; size_bytes: number }>
): Promise<{
  upload_batch_id: string;
  intents: Array<{
    filename: string;
    storage_key: string;
    presigned_url: string;
    expires_at: string;
  }>;
}> {
  return request(`/cases/${caseId}/upload-intent`, {
    method: 'POST',
    body: JSON.stringify({ files }),
  });
}

// Upload file to presigned URL
export async function uploadFile(
  presignedUrl: string,
  file: File
): Promise<void> {
  const response = await fetch(presignedUrl, {
    method: 'PUT',
    body: file,
    headers: {
      'Content-Type': file.type,
    },
  });

  if (!response.ok) {
    throw new Error(`Upload failed: ${response.statusText}`);
  }
}

// Jobs
export async function createJob(
  caseId: string,
  type: JobType,
  input: Record<string, unknown>,
  idempotencyKey?: string
): Promise<Job> {
  const headers: Record<string, string> = {};
  if (idempotencyKey) {
    headers['Idempotency-Key'] = idempotencyKey;
  }

  return request<Job>(`/cases/${caseId}/jobs`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ type, input }),
  });
}

export async function getJob(jobId: string): Promise<Job> {
  return request<Job>(`/jobs/${jobId}`);
}

// Witness Statements
export async function submitWitnessStatements(
  caseId: string,
  statements: Array<{
    source_name: string;
    content: string;
    credibility: number;
  }>
): Promise<{ commit_id: string; profile_job_id?: string }> {
  return request(`/cases/${caseId}/witness-statements`, {
    method: 'POST',
    body: JSON.stringify({ statements }),
  });
}

// Reasoning
export async function triggerReasoning(
  caseId: string,
  options?: {
    thinking_budget?: number;
    max_trajectories?: number;
    constraints_override?: unknown[];
  }
): Promise<Job> {
  return request<Job>(`/cases/${caseId}/reasoning`, {
    method: 'POST',
    body: JSON.stringify(options || {}),
  });
}

// Export
export async function triggerExport(
  caseId: string,
  format: 'html' | 'pdf' = 'html'
): Promise<Job> {
  return request<Job>(`/cases/${caseId}/export`, {
    method: 'POST',
    body: JSON.stringify({ format }),
  });
}

// Branches
export async function createBranch(
  caseId: string,
  name: string,
  baseCommitId: string
): Promise<{ id: string; name: string }> {
  return request(`/cases/${caseId}/branches`, {
    method: 'POST',
    body: JSON.stringify({ name, base_commit_id: baseCommitId }),
  });
}

// Simulation - Text to Motion
export async function generateMotionPath(
  caseId: string,
  prompt: string,
  options?: {
    constraints?: unknown[];
    max_paths?: number;
  }
): Promise<Job> {
  return request<Job>(`/cases/${caseId}/jobs`, {
    method: 'POST',
    body: JSON.stringify({
      type: 'reasoning',
      input: {
        simulation_prompt: prompt,
        mode: 'text_to_motion',
        ...options,
      },
    }),
  });
}

// Video Analysis
export async function analyzeVideo(
  caseId: string,
  videoAssetKey: string,
  options?: {
    extract_motion?: boolean;
    detect_persons?: boolean;
    detect_objects?: boolean;
  }
): Promise<Job> {
  return request<Job>(`/cases/${caseId}/jobs`, {
    method: 'POST',
    body: JSON.stringify({
      type: 'scene_analysis',
      input: {
        video_asset_key: videoAssetKey,
        ...options,
      },
    }),
  });
}

// Suspect Profile
export async function getSuspectProfile(caseId: string): Promise<{
  case_id: string;
  commit_id: string;
  attributes: Record<string, unknown>;
  portrait_asset_key?: string;
  updated_at: string;
} | null> {
  try {
    return await request(`/cases/${caseId}/suspect-profile`);
  } catch (e) {
    // Profile may not exist yet
    return null;
  }
}

// Get asset URL (for portrait, etc.)
export function getAssetUrl(storageKey: string): string {
  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL || 'https://hdfaugwofzqqdjuzcsin.supabase.co';
  return `${supabaseUrl}/storage/v1/object/public/assets/${storageKey}`;
}

// Get export result (after job completes)
export async function getExportResult(jobId: string): Promise<{
  report_asset_key?: string;
  report_url?: string;
}> {
  const job = await getJob(jobId);
  if (job.status === 'done' && job.output) {
    return job.output as { report_asset_key?: string; report_url?: string };
  }
  throw new Error('Export not ready');
}

export { ApiError };
