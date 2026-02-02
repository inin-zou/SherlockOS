'use client';

import { useEffect, useRef, useCallback, useState } from 'react';
import { getSupabase, RealtimeChannel } from '@/lib/supabase';
import { useStore } from '@/lib/store';
import * as api from '@/lib/api';
import type { Job } from '@/lib/types';

interface RealtimeJobPayload {
  id: string;
  case_id: string;
  type: string;
  status: 'queued' | 'running' | 'done' | 'failed' | 'canceled';
  progress: number;
  input: Record<string, unknown>;
  output?: Record<string, unknown>;
  error?: string;
  created_at: string;
  updated_at: string;
}

export interface UseRealtimeJobsResult {
  isConnected: boolean;
  error: string | null;
  subscribe: (caseId: string) => void;
  unsubscribe: () => void;
}

/**
 * Hook to subscribe to realtime job updates via Supabase
 * Falls back gracefully if Supabase is not configured
 */
export function useRealtimeJobs(): UseRealtimeJobsResult {
  const { updateJob, setSceneGraph } = useStore();
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const channelRef = useRef<RealtimeChannel | null>(null);
  const caseIdRef = useRef<string | null>(null);

  const handleJobChange = useCallback(
    async (payload: { new: RealtimeJobPayload; old: RealtimeJobPayload | null }) => {
      const job = payload.new;

      // Update job in store
      const mappedJob: Job = {
        id: job.id,
        case_id: job.case_id,
        type: job.type as Job['type'],
        status: job.status,
        progress: job.progress,
        input: job.input,
        output: job.output,
        error: job.error,
        created_at: job.created_at,
        updated_at: job.updated_at,
      };

      updateJob(job.id, mappedJob);

      // If job completed successfully, refresh scene data
      if (job.status === 'done' && caseIdRef.current) {
        try {
          const snapshot = await api.getSnapshot(caseIdRef.current);
          setSceneGraph(snapshot.scenegraph);
        } catch (err) {
          console.error('Failed to refresh scene after job completion:', err);
        }
      }
    },
    [updateJob, setSceneGraph]
  );

  const subscribe = useCallback(
    (caseId: string) => {
      const supabase = getSupabase();
      if (!supabase) {
        setError('Supabase not configured');
        return;
      }

      // Unsubscribe from previous channel if exists
      if (channelRef.current) {
        supabase.removeChannel(channelRef.current);
      }

      caseIdRef.current = caseId;

      // Create a new channel for this case's jobs
      const channel = supabase
        .channel(`jobs:case_id=eq.${caseId}`)
        .on(
          'postgres_changes',
          {
            event: '*',
            schema: 'public',
            table: 'jobs',
            filter: `case_id=eq.${caseId}`,
          },
          (payload) => {
            handleJobChange(payload as { new: RealtimeJobPayload; old: RealtimeJobPayload | null });
          }
        )
        .subscribe((status) => {
          if (status === 'SUBSCRIBED') {
            setIsConnected(true);
            setError(null);
          } else if (status === 'CLOSED') {
            setIsConnected(false);
          } else if (status === 'CHANNEL_ERROR') {
            setIsConnected(false);
            setError('Failed to connect to realtime channel');
          }
        });

      channelRef.current = channel;
    },
    [handleJobChange]
  );

  const unsubscribe = useCallback(() => {
    const supabase = getSupabase();
    if (supabase && channelRef.current) {
      supabase.removeChannel(channelRef.current);
      channelRef.current = null;
      caseIdRef.current = null;
      setIsConnected(false);
    }
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      unsubscribe();
    };
  }, [unsubscribe]);

  return {
    isConnected,
    error,
    subscribe,
    unsubscribe,
  };
}

/**
 * Hook to subscribe to realtime scene/snapshot updates
 */
export function useRealtimeSnapshot(caseId: string | undefined): {
  isConnected: boolean;
} {
  const { setSceneGraph, setCommits } = useStore();
  const [isConnected, setIsConnected] = useState(false);
  const channelRef = useRef<RealtimeChannel | null>(null);

  useEffect(() => {
    if (!caseId) return;

    const supabase = getSupabase();
    if (!supabase) return;

    // Subscribe to commits table for timeline updates
    const channel = supabase
      .channel(`commits:case_id=eq.${caseId}`)
      .on(
        'postgres_changes',
        {
          event: 'INSERT',
          schema: 'public',
          table: 'commits',
          filter: `case_id=eq.${caseId}`,
        },
        async () => {
          // Refresh timeline on new commit
          try {
            const timeline = await api.getTimeline(caseId);
            setCommits(timeline.commits);

            // Also refresh scene snapshot
            const snapshot = await api.getSnapshot(caseId);
            setSceneGraph(snapshot.scenegraph);
          } catch (err) {
            console.error('Failed to refresh after new commit:', err);
          }
        }
      )
      .subscribe((status) => {
        setIsConnected(status === 'SUBSCRIBED');
      });

    channelRef.current = channel;

    return () => {
      if (channelRef.current) {
        supabase.removeChannel(channelRef.current);
        channelRef.current = null;
      }
    };
  }, [caseId, setSceneGraph, setCommits]);

  return { isConnected };
}
