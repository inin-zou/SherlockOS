import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useJobProgress } from './useJobs';

const mockJobs = [
  { id: 'job-1', type: 'reconstruction', status: 'running', progress: 50, created_at: new Date().toISOString() },
  { id: 'job-2', type: 'reasoning', status: 'queued', progress: 0, created_at: new Date().toISOString() },
  { id: 'job-3', type: 'profile', status: 'done', progress: 100, created_at: new Date().toISOString() },
  { id: 'job-4', type: 'profile', status: 'failed', progress: 0, error: 'Processing error', created_at: new Date().toISOString() },
];

// Mock store
vi.mock('@/lib/store', () => ({
  useStore: () => ({
    jobs: mockJobs,
    setJobs: vi.fn(),
    updateJob: vi.fn(),
  }),
}));

describe('useJobProgress', () => {
  it('returns null for undefined jobId', () => {
    const { result } = renderHook(() => useJobProgress(undefined));

    expect(result.current.job).toBeNull();
    expect(result.current.isComplete).toBe(false);
    expect(result.current.isError).toBe(false);
  });

  it('finds job in store by id', () => {
    const { result } = renderHook(() => useJobProgress('job-1'));

    expect(result.current.job).toEqual(mockJobs[0]);
    expect(result.current.job?.status).toBe('running');
    expect(result.current.job?.progress).toBe(50);
  });

  it('returns isComplete true when job is done', () => {
    const { result } = renderHook(() => useJobProgress('job-3'));

    expect(result.current.job).toEqual(mockJobs[2]);
    expect(result.current.isComplete).toBe(true);
    expect(result.current.isError).toBe(false);
  });

  it('returns isError true when job failed', () => {
    const { result } = renderHook(() => useJobProgress('job-4'));

    expect(result.current.job).toEqual(mockJobs[3]);
    expect(result.current.isComplete).toBe(false);
    expect(result.current.isError).toBe(true);
  });

  it('returns null for non-existent job', () => {
    const { result } = renderHook(() => useJobProgress('non-existent'));

    expect(result.current.job).toBeNull();
    expect(result.current.isComplete).toBe(false);
    expect(result.current.isError).toBe(false);
  });

  it('correctly identifies queued jobs as not complete', () => {
    const { result } = renderHook(() => useJobProgress('job-2'));

    expect(result.current.job?.status).toBe('queued');
    expect(result.current.isComplete).toBe(false);
    expect(result.current.isError).toBe(false);
  });
});
