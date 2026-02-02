'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import * as api from '@/lib/api';
import { useStore } from '@/lib/store';
import { useRealtimeJobs } from './useRealtime';
import type { Job } from '@/lib/types';

const POLL_INTERVAL = 2000; // 2 seconds

export interface UseJobsResult {
  jobs: Job[];
  activeJobs: Job[];
  isLoading: boolean;
  error: string | null;
  refreshJobs: () => Promise<void>;
  pollJob: (jobId: string) => void;
  stopPolling: (jobId: string) => void;
  isRealtimeConnected: boolean;
}

export function useJobs(caseId: string | undefined): UseJobsResult {
  const { jobs, setJobs, updateJob } = useStore();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pollingJobs = useRef<Set<string>>(new Set());
  const intervalIds = useRef<Map<string, NodeJS.Timeout>>(new Map());

  // Try to use realtime subscriptions
  const { isConnected: isRealtimeConnected, subscribe, unsubscribe } = useRealtimeJobs();

  // Get active (non-terminal) jobs
  const activeJobs = jobs.filter(
    (job) => job.status === 'queued' || job.status === 'running'
  );

  // Subscribe to realtime when caseId changes
  useEffect(() => {
    if (caseId) {
      subscribe(caseId);
    }
    return () => {
      unsubscribe();
    };
  }, [caseId, subscribe, unsubscribe]);

  // Refresh jobs list
  const refreshJobs = useCallback(async () => {
    if (!caseId) return;

    setIsLoading(true);
    setError(null);

    try {
      // Note: This assumes an API endpoint to list jobs for a case
      // If not available, jobs are tracked via individual pollJob calls
      setIsLoading(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch jobs');
      setIsLoading(false);
    }
  }, [caseId]);

  // Poll a specific job (used as fallback when realtime is not available)
  const pollJob = useCallback(
    (jobId: string) => {
      // Skip polling if realtime is connected
      if (isRealtimeConnected) {
        return;
      }

      if (pollingJobs.current.has(jobId)) return;

      pollingJobs.current.add(jobId);

      const poll = async () => {
        try {
          const job = await api.getJob(jobId);

          // Update store
          updateJob(jobId, job);

          // Check if job is complete
          if (job.status === 'done' || job.status === 'failed' || job.status === 'canceled') {
            stopPolling(jobId);

            // If job completed successfully, refresh scene data
            if (job.status === 'done' && caseId) {
              // Trigger scene refresh
              const snapshot = await api.getSnapshot(caseId);
              useStore.getState().setSceneGraph(snapshot.scenegraph);
            }
          }
        } catch (err) {
          console.error(`Failed to poll job ${jobId}:`, err);
        }
      };

      // Initial poll
      poll();

      // Set up interval
      const intervalId = setInterval(poll, POLL_INTERVAL);
      intervalIds.current.set(jobId, intervalId);
    },
    [caseId, updateJob, isRealtimeConnected]
  );

  // Stop polling a job
  const stopPolling = useCallback((jobId: string) => {
    pollingJobs.current.delete(jobId);
    const intervalId = intervalIds.current.get(jobId);
    if (intervalId) {
      clearInterval(intervalId);
      intervalIds.current.delete(jobId);
    }
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      intervalIds.current.forEach((intervalId) => clearInterval(intervalId));
      intervalIds.current.clear();
      pollingJobs.current.clear();
    };
  }, []);

  // Auto-poll active jobs only if realtime is not connected
  useEffect(() => {
    if (isRealtimeConnected) {
      // Stop all polling when realtime connects
      intervalIds.current.forEach((intervalId) => clearInterval(intervalId));
      intervalIds.current.clear();
      pollingJobs.current.clear();
      return;
    }

    // Fallback to polling
    activeJobs.forEach((job) => {
      if (!pollingJobs.current.has(job.id)) {
        pollJob(job.id);
      }
    });
  }, [activeJobs, pollJob, isRealtimeConnected]);

  return {
    jobs,
    activeJobs,
    isLoading,
    error,
    refreshJobs,
    pollJob,
    stopPolling,
    isRealtimeConnected,
  };
}

/**
 * Hook to track a single job's progress
 */
export function useJobProgress(jobId: string | undefined): {
  job: Job | null;
  isComplete: boolean;
  isError: boolean;
} {
  const { jobs } = useStore();
  const job = jobId ? jobs.find((j) => j.id === jobId) || null : null;

  return {
    job,
    isComplete: job?.status === 'done',
    isError: job?.status === 'failed',
  };
}
