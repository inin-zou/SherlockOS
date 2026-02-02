import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  getCases,
  getCase,
  createCase,
  getTimeline,
  getSnapshot,
  getUploadIntent,
  uploadFile,
  createJob,
  getJob,
  submitWitnessStatements,
  triggerReasoning,
  triggerExport,
  createBranch,
  generateMotionPath,
  analyzeVideo,
  getSuspectProfile,
  getAssetUrl,
  getExportResult,
  ApiError,
} from './api';

// Mock fetch globally
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('API Client', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockFetch.mockReset();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('ApiError', () => {
    it('creates error with code and message', () => {
      const error = new ApiError('NOT_FOUND', 'Resource not found');
      expect(error.code).toBe('NOT_FOUND');
      expect(error.message).toBe('Resource not found');
      expect(error.name).toBe('ApiError');
    });

    it('includes details when provided', () => {
      const error = new ApiError('VALIDATION_ERROR', 'Invalid input', { field: 'title' });
      expect(error.details).toEqual({ field: 'title' });
    });
  });

  describe('getCases', () => {
    it('returns list of cases', async () => {
      const cases = [
        { id: '1', title: 'Case 1' },
        { id: '2', title: 'Case 2' },
      ];
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: cases }),
      });

      const result = await getCases();
      expect(result).toEqual(cases);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/cases'),
        expect.objectContaining({
          headers: { 'Content-Type': 'application/json' },
        })
      );
    });

    it('throws ApiError on failure', async () => {
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({
          success: false,
          error: { code: 'INTERNAL_ERROR', message: 'Server error' },
        }),
      });

      await expect(getCases()).rejects.toThrow(ApiError);
    });
  });

  describe('getCase', () => {
    it('returns single case by ID', async () => {
      const caseData = { id: 'case-123', title: 'Test Case' };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: caseData }),
      });

      const result = await getCase('case-123');
      expect(result).toEqual(caseData);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/cases/case-123'),
        expect.any(Object)
      );
    });
  });

  describe('createCase', () => {
    it('creates a new case', async () => {
      const newCase = { id: 'new-id', title: 'New Case', description: 'Desc' };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: newCase }),
      });

      const result = await createCase('New Case', 'Desc');
      expect(result).toEqual(newCase);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/cases'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ title: 'New Case', description: 'Desc' }),
        })
      );
    });

    it('creates case without description', async () => {
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: { id: '1', title: 'Case' } }),
      });

      await createCase('Case');
      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          body: JSON.stringify({ title: 'Case', description: undefined }),
        })
      );
    });
  });

  describe('getTimeline', () => {
    it('returns commits for a case', async () => {
      const commits = [
        { id: 'c1', type: 'upload_scan' },
        { id: 'c2', type: 'witness_statement' },
      ];
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: commits }),
      });

      const result = await getTimeline('case-123');
      expect(result.commits).toEqual(commits);
    });

    it('includes pagination parameters', async () => {
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: [] }),
      });

      await getTimeline('case-123', 'cursor-abc', 25);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('cursor=cursor-abc'),
        expect.any(Object)
      );
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('limit=25'),
        expect.any(Object)
      );
    });

    it('returns empty array when data is null', async () => {
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: null }),
      });

      const result = await getTimeline('case-123');
      expect(result.commits).toEqual([]);
    });
  });

  describe('getSnapshot', () => {
    it('returns scene snapshot', async () => {
      const snapshot = {
        case_id: 'case-123',
        commit_id: 'commit-456',
        scenegraph: { objects: [] },
        updated_at: '2024-01-01T00:00:00Z',
      };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: snapshot }),
      });

      const result = await getSnapshot('case-123');
      expect(result).toEqual(snapshot);
    });
  });

  describe('getUploadIntent', () => {
    it('returns presigned URLs for uploads', async () => {
      const intent = {
        upload_batch_id: 'batch-123',
        intents: [
          { filename: 'image.jpg', presigned_url: 'https://example.com/upload', storage_key: 'key' },
        ],
      };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: intent }),
      });

      const result = await getUploadIntent('case-123', [
        { filename: 'image.jpg', content_type: 'image/jpeg', size_bytes: 1024 },
      ]);
      expect(result).toEqual(intent);
    });
  });

  describe('uploadFile', () => {
    it('uploads file to presigned URL', async () => {
      mockFetch.mockResolvedValueOnce({ ok: true });

      const file = new File(['content'], 'test.jpg', { type: 'image/jpeg' });
      await uploadFile('https://example.com/upload', file);

      expect(mockFetch).toHaveBeenCalledWith(
        'https://example.com/upload',
        expect.objectContaining({
          method: 'PUT',
          body: file,
          headers: { 'Content-Type': 'image/jpeg' },
        })
      );
    });

    it('throws error on upload failure', async () => {
      mockFetch.mockResolvedValueOnce({ ok: false, statusText: 'Forbidden' });

      const file = new File(['content'], 'test.jpg', { type: 'image/jpeg' });
      await expect(uploadFile('https://example.com/upload', file)).rejects.toThrow('Upload failed: Forbidden');
    });
  });

  describe('createJob', () => {
    it('creates a job', async () => {
      const job = { id: 'job-123', type: 'reconstruction', status: 'queued' };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: job }),
      });

      const result = await createJob('case-123', 'reconstruction', { files: ['file1'] });
      expect(result).toEqual(job);
    });

    it('includes idempotency key when provided', async () => {
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: { id: 'job-123' } }),
      });

      await createJob('case-123', 'reconstruction', {}, 'idem-key');
      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            'Idempotency-Key': 'idem-key',
          }),
        })
      );
    });
  });

  describe('getJob', () => {
    it('returns job by ID', async () => {
      const job = { id: 'job-123', status: 'running', progress: 50 };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: job }),
      });

      const result = await getJob('job-123');
      expect(result).toEqual(job);
    });
  });

  describe('submitWitnessStatements', () => {
    it('submits statements and returns commit info', async () => {
      const response = { commit_id: 'commit-123', profile_job_id: 'job-456' };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: response }),
      });

      const result = await submitWitnessStatements('case-123', [
        { source_name: 'John', content: 'I saw something', credibility: 0.8 },
      ]);
      expect(result).toEqual(response);
    });
  });

  describe('triggerReasoning', () => {
    it('triggers reasoning job', async () => {
      const job = { id: 'job-123', type: 'reasoning', status: 'queued' };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: job }),
      });

      const result = await triggerReasoning('case-123', { thinking_budget: 10000 });
      expect(result).toEqual(job);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/reasoning'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ thinking_budget: 10000 }),
        })
      );
    });

    it('works without options', async () => {
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: { id: 'job' } }),
      });

      await triggerReasoning('case-123');
      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          body: JSON.stringify({}),
        })
      );
    });
  });

  describe('triggerExport', () => {
    it('triggers export job with default format', async () => {
      const job = { id: 'job-123', type: 'export', status: 'queued' };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: job }),
      });

      const result = await triggerExport('case-123');
      expect(result).toEqual(job);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          body: JSON.stringify({ format: 'html' }),
        })
      );
    });

    it('supports PDF format', async () => {
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: { id: 'job' } }),
      });

      await triggerExport('case-123', 'pdf');
      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          body: JSON.stringify({ format: 'pdf' }),
        })
      );
    });
  });

  describe('createBranch', () => {
    it('creates hypothesis branch', async () => {
      const branch = { id: 'branch-123', name: 'Hypothesis A' };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: branch }),
      });

      const result = await createBranch('case-123', 'Hypothesis A', 'commit-456');
      expect(result).toEqual(branch);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/branches'),
        expect.objectContaining({
          body: JSON.stringify({ name: 'Hypothesis A', base_commit_id: 'commit-456' }),
        })
      );
    });
  });

  describe('generateMotionPath', () => {
    it('generates motion path from prompt', async () => {
      const job = { id: 'job-123', type: 'reasoning' };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: job }),
      });

      const result = await generateMotionPath('case-123', 'Enter through window', { max_paths: 3 });
      expect(result).toEqual(job);
      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          body: expect.stringContaining('text_to_motion'),
        })
      );
    });
  });

  describe('analyzeVideo', () => {
    it('creates video analysis job', async () => {
      const job = { id: 'job-123', type: 'scene_analysis' };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: job }),
      });

      const result = await analyzeVideo('case-123', 'videos/test.mp4', {
        detect_persons: true,
        extract_motion: true,
      });
      expect(result).toEqual(job);
    });
  });

  describe('getSuspectProfile', () => {
    it('returns suspect profile', async () => {
      const profile = {
        case_id: 'case-123',
        attributes: { age: '25-35' },
        portrait_asset_key: 'portraits/suspect.png',
      };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: profile }),
      });

      const result = await getSuspectProfile('case-123');
      expect(result).toEqual(profile);
    });

    it('returns null when profile does not exist', async () => {
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({
          success: false,
          error: { code: 'NOT_FOUND', message: 'Not found' },
        }),
      });

      const result = await getSuspectProfile('case-123');
      expect(result).toBeNull();
    });
  });

  describe('getAssetUrl', () => {
    it('constructs storage URL', () => {
      const url = getAssetUrl('portraits/suspect.png');
      expect(url).toContain('storage/v1/object/public/assets/portraits/suspect.png');
    });
  });

  describe('getExportResult', () => {
    it('returns export result when job is done', async () => {
      const job = {
        id: 'job-123',
        status: 'done',
        output: { report_asset_key: 'reports/report.html', report_url: 'https://example.com/report' },
      };
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({ success: true, data: job }),
      });

      const result = await getExportResult('job-123');
      expect(result).toEqual(job.output);
    });

    it('throws error when job not done', async () => {
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({
          success: true,
          data: { id: 'job-123', status: 'running' },
        }),
      });

      await expect(getExportResult('job-123')).rejects.toThrow('Export not ready');
    });

    it('throws error when output is null', async () => {
      mockFetch.mockResolvedValueOnce({
        json: () => Promise.resolve({
          success: true,
          data: { id: 'job-123', status: 'done', output: null },
        }),
      });

      await expect(getExportResult('job-123')).rejects.toThrow('Export not ready');
    });
  });
});
