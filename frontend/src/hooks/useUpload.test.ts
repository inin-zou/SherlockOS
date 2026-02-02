import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useUpload } from './useUpload';

// Mock dependencies
vi.mock('@/lib/api', () => ({
  getUploadIntent: vi.fn(),
  uploadFile: vi.fn(),
  createJob: vi.fn(),
  submitWitnessStatements: vi.fn(),
}));

vi.mock('@/lib/store', () => ({
  useStore: () => ({
    addCommit: vi.fn(),
    setJobs: vi.fn(),
  }),
}));

vi.mock('@/lib/fileClassifier', () => ({
  classifyFiles: vi.fn((files: File[]) =>
    files.map((file, i) => ({
      file,
      tier: i % 4 as 0 | 1 | 2 | 3,
      tierName: ['Environment', 'Ground Truth', 'Electronic Logs', 'Testimonials'][i % 4],
    }))
  ),
  groupByTier: vi.fn((classified) => {
    const grouped = { 0: [], 1: [], 2: [], 3: [] } as Record<number, any[]>;
    classified.forEach((cf: any) => {
      grouped[cf.tier].push(cf);
    });
    return grouped;
  }),
  getJobTypeForTier: vi.fn((tier: number) => (tier <= 1 ? 'reconstruction' : null)),
  parseTier2File: vi.fn().mockResolvedValue({ events: [] }),
  parseTier3File: vi.fn().mockResolvedValue({
    source_name: 'Test Witness',
    content: 'Test statement',
    credibility: 0.7,
  }),
}));

import * as api from '@/lib/api';

describe('useUpload', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('initializes with empty state', () => {
    const { result } = renderHook(() => useUpload('case-123'));

    expect(result.current.progress).toEqual([]);
    expect(result.current.isUploading).toBe(false);
    expect(result.current.classifiedFiles).toEqual([]);
  });

  it('does nothing without caseId', async () => {
    const { result } = renderHook(() => useUpload(undefined));

    const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });
    const fileList = {
      0: mockFile,
      length: 1,
      item: (i: number) => (i === 0 ? mockFile : null),
      [Symbol.iterator]: function* () {
        yield mockFile;
      },
    } as unknown as FileList;

    await act(async () => {
      await result.current.uploadFiles(fileList);
    });

    expect(api.getUploadIntent).not.toHaveBeenCalled();
  });

  it('classifies files and sets progress', async () => {
    (api.getUploadIntent as ReturnType<typeof vi.fn>).mockResolvedValue({
      upload_batch_id: 'batch-1',
      intents: [
        {
          filename: 'test.pdf',
          storage_key: 'key-1',
          presigned_url: 'https://example.com/upload',
          expires_at: new Date().toISOString(),
        },
      ],
    });
    (api.uploadFile as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);
    (api.createJob as ReturnType<typeof vi.fn>).mockResolvedValue({
      id: 'job-1',
      type: 'reconstruction',
      status: 'queued',
    });

    const { result } = renderHook(() => useUpload('case-123'));

    const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });

    await act(async () => {
      await result.current.uploadFiles([mockFile]);
    });

    expect(result.current.classifiedFiles.length).toBeGreaterThan(0);
  });

  it('clears progress when requested', async () => {
    const { result } = renderHook(() => useUpload('case-123'));

    // Set some initial state
    act(() => {
      result.current.clearProgress();
    });

    expect(result.current.progress).toEqual([]);
    expect(result.current.classifiedFiles).toEqual([]);
  });

  it('sets isUploading during upload process', async () => {
    (api.getUploadIntent as ReturnType<typeof vi.fn>).mockImplementation(
      () => new Promise((resolve) => setTimeout(() => resolve({
        upload_batch_id: 'batch-1',
        intents: [{
          filename: 'test.pdf',
          storage_key: 'key-1',
          presigned_url: 'https://example.com/upload',
          expires_at: new Date().toISOString(),
        }],
      }), 50))
    );
    (api.uploadFile as ReturnType<typeof vi.fn>).mockResolvedValue(undefined);
    (api.createJob as ReturnType<typeof vi.fn>).mockResolvedValue({
      id: 'job-1',
      type: 'reconstruction',
      status: 'queued',
    });

    const { result } = renderHook(() => useUpload('case-123'));

    const mockFile = new File(['test'], 'test.pdf', { type: 'application/pdf' });

    // Start upload
    act(() => {
      result.current.uploadFiles([mockFile]);
    });

    // Should be uploading
    expect(result.current.isUploading).toBe(true);

    // Wait for completion
    await waitFor(() => {
      expect(result.current.isUploading).toBe(false);
    });
  });
});
