import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useRealtimeJobs, useRealtimeSnapshot } from './useRealtime';

// Mock channel
const mockChannel = {
  on: vi.fn().mockReturnThis(),
  subscribe: vi.fn((callback) => {
    // Simulate subscription status change
    if (callback) callback('SUBSCRIBED');
    return mockChannel;
  }),
};

// Mock Supabase client
const mockSupabase = {
  channel: vi.fn(() => mockChannel),
  removeChannel: vi.fn(),
};

// Track whether supabase should return null
let supabaseEnabled = true;

vi.mock('@/lib/supabase', () => ({
  getSupabase: vi.fn(() => (supabaseEnabled ? mockSupabase : null)),
}));

// Mock store
const mockUpdateJob = vi.fn();
const mockSetSceneGraph = vi.fn();
const mockSetCommits = vi.fn();

vi.mock('@/lib/store', () => ({
  useStore: vi.fn(() => ({
    updateJob: mockUpdateJob,
    setSceneGraph: mockSetSceneGraph,
    setCommits: mockSetCommits,
  })),
}));

// Mock API
vi.mock('@/lib/api', () => ({
  getSnapshot: vi.fn().mockResolvedValue({ scenegraph: {} }),
  getTimeline: vi.fn().mockResolvedValue({ commits: [] }),
}));

describe('useRealtimeJobs', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    supabaseEnabled = true;
  });

  it('starts with disconnected state', () => {
    const { result } = renderHook(() => useRealtimeJobs());

    expect(result.current.isConnected).toBe(false);
    expect(result.current.error).toBe(null);
  });

  it('provides subscribe and unsubscribe functions', () => {
    const { result } = renderHook(() => useRealtimeJobs());

    expect(typeof result.current.subscribe).toBe('function');
    expect(typeof result.current.unsubscribe).toBe('function');
  });

  it('creates channel when subscribing', () => {
    const { result } = renderHook(() => useRealtimeJobs());

    act(() => {
      result.current.subscribe('case-123');
    });

    expect(mockSupabase.channel).toHaveBeenCalledWith('jobs:case_id=eq.case-123');
    expect(mockChannel.on).toHaveBeenCalledWith(
      'postgres_changes',
      expect.objectContaining({
        event: '*',
        schema: 'public',
        table: 'jobs',
        filter: 'case_id=eq.case-123',
      }),
      expect.any(Function)
    );
    expect(mockChannel.subscribe).toHaveBeenCalled();
  });

  it('sets connected state when subscription succeeds', () => {
    const { result } = renderHook(() => useRealtimeJobs());

    act(() => {
      result.current.subscribe('case-123');
    });

    expect(result.current.isConnected).toBe(true);
    expect(result.current.error).toBe(null);
  });

  it('calls removeChannel when unsubscribing after subscribe', () => {
    const { result } = renderHook(() => useRealtimeJobs());

    act(() => {
      result.current.subscribe('case-123');
    });

    act(() => {
      result.current.unsubscribe();
    });

    expect(mockSupabase.removeChannel).toHaveBeenCalledWith(mockChannel);
  });

  it('sets error when Supabase not configured', () => {
    supabaseEnabled = false;

    const { result } = renderHook(() => useRealtimeJobs());

    act(() => {
      result.current.subscribe('case-123');
    });

    expect(result.current.error).toBe('Supabase not configured');
    expect(result.current.isConnected).toBe(false);
  });
});

describe('useRealtimeSnapshot', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    supabaseEnabled = true;
  });

  it('does not create channel without caseId', () => {
    renderHook(() => useRealtimeSnapshot(undefined));

    expect(mockSupabase.channel).not.toHaveBeenCalled();
  });

  it('creates channel for commits when caseId provided', () => {
    renderHook(() => useRealtimeSnapshot('case-456'));

    expect(mockSupabase.channel).toHaveBeenCalledWith('commits:case_id=eq.case-456');
    expect(mockChannel.on).toHaveBeenCalledWith(
      'postgres_changes',
      expect.objectContaining({
        event: 'INSERT',
        schema: 'public',
        table: 'commits',
        filter: 'case_id=eq.case-456',
      }),
      expect.any(Function)
    );
  });

  it('reports connection status', () => {
    const { result } = renderHook(() => useRealtimeSnapshot('case-456'));

    expect(result.current.isConnected).toBe(true);
  });

  it('handles missing Supabase gracefully', () => {
    supabaseEnabled = false;

    const { result } = renderHook(() => useRealtimeSnapshot('case-456'));

    expect(result.current.isConnected).toBe(false);
    expect(mockSupabase.channel).not.toHaveBeenCalled();
  });
});
