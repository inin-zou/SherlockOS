'use client';

import { useState, useCallback } from 'react';
import {
  classifyFiles,
  groupByTier,
  getJobTypeForTier,
  parseTier2File,
  parseTier3File,
  type ClassifiedFile,
  type EvidenceTier,
} from '@/lib/fileClassifier';
import * as api from '@/lib/api';
import { useStore } from '@/lib/store';
import type { Job, JobType } from '@/lib/types';

export interface UploadProgress {
  fileId: string;
  filename: string;
  tier: EvidenceTier;
  status: 'pending' | 'uploading' | 'processing' | 'done' | 'error';
  progress: number;
  error?: string;
  jobId?: string;
}

export interface UseUploadResult {
  uploadFiles: (files: FileList | File[]) => Promise<void>;
  progress: UploadProgress[];
  isUploading: boolean;
  classifiedFiles: ClassifiedFile[];
  clearProgress: () => void;
}

export function useUpload(caseId: string | undefined): UseUploadResult {
  const [progress, setProgress] = useState<UploadProgress[]>([]);
  const [isUploading, setIsUploading] = useState(false);
  const [classifiedFiles, setClassifiedFiles] = useState<ClassifiedFile[]>([]);

  const { addCommit, setJobs } = useStore();

  const updateProgress = useCallback((fileId: string, updates: Partial<UploadProgress>) => {
    setProgress((prev) =>
      prev.map((p) => (p.fileId === fileId ? { ...p, ...updates } : p))
    );
  }, []);

  const uploadFiles = useCallback(
    async (files: FileList | File[]) => {
      if (!caseId) {
        console.error('No case ID provided');
        return;
      }

      setIsUploading(true);

      // Classify files
      const classified = classifyFiles(files);
      setClassifiedFiles(classified);

      // Initialize progress
      const initialProgress: UploadProgress[] = classified.map((cf, i) => ({
        fileId: `file-${i}-${Date.now()}`,
        filename: cf.file.name,
        tier: cf.tier,
        status: 'pending',
        progress: 0,
      }));
      setProgress(initialProgress);

      // Group by tier for batch processing
      const grouped = groupByTier(classified);

      // Process each tier
      for (const tier of [0, 1, 2, 3] as EvidenceTier[]) {
        const tierFiles = grouped[tier];
        if (tierFiles.length === 0) continue;

        for (let i = 0; i < tierFiles.length; i++) {
          const cf = tierFiles[i];
          const fileId = initialProgress.find((p) => p.filename === cf.file.name)?.fileId;
          if (!fileId) continue;

          try {
            updateProgress(fileId, { status: 'uploading', progress: 10 });

            if (tier === 0 || tier === 1) {
              // Tier 0-1: Upload to storage and trigger job
              await uploadAndProcess(caseId, cf, fileId, updateProgress);
            } else if (tier === 2) {
              // Tier 2: Parse locally and add to timeline
              await processTier2(caseId, cf, fileId, updateProgress);
            } else if (tier === 3) {
              // Tier 3: Submit as witness statement
              await processTier3(caseId, cf, fileId, updateProgress);
            }

            updateProgress(fileId, { status: 'done', progress: 100 });
          } catch (error) {
            updateProgress(fileId, {
              status: 'error',
              error: error instanceof Error ? error.message : 'Upload failed',
            });
          }
        }
      }

      setIsUploading(false);
    },
    [caseId, updateProgress]
  );

  const clearProgress = useCallback(() => {
    setProgress([]);
    setClassifiedFiles([]);
  }, []);

  return {
    uploadFiles,
    progress,
    isUploading,
    classifiedFiles,
    clearProgress,
  };
}

/**
 * Upload file to storage and trigger processing job
 */
async function uploadAndProcess(
  caseId: string,
  cf: ClassifiedFile,
  fileId: string,
  updateProgress: (id: string, updates: Partial<UploadProgress>) => void
): Promise<void> {
  // Get upload intent
  const intent = await api.getUploadIntent(caseId, [
    {
      filename: cf.file.name,
      content_type: cf.file.type || 'application/octet-stream',
      size_bytes: cf.file.size,
    },
  ]);

  updateProgress(fileId, { progress: 30 });

  // Upload to presigned URL
  const uploadInfo = intent.intents[0];
  if (uploadInfo) {
    await api.uploadFile(uploadInfo.presigned_url, cf.file);
  }

  updateProgress(fileId, { progress: 60, status: 'processing' });

  // Trigger appropriate job
  const jobType = getJobTypeForTier(cf.tier);
  if (jobType && uploadInfo) {
    const job = await api.createJob(caseId, jobType as JobType, {
      scan_asset_keys: [uploadInfo.storage_key],
    });
    updateProgress(fileId, { jobId: job.id, progress: 80 });
  }
}

/**
 * Process Tier 2 (Electronic Logs) file
 */
async function processTier2(
  caseId: string,
  cf: ClassifiedFile,
  fileId: string,
  updateProgress: (id: string, updates: Partial<UploadProgress>) => void
): Promise<void> {
  updateProgress(fileId, { progress: 50, status: 'processing' });

  const parsed = await parseTier2File(cf.file);

  // Events would be added to timeline via API
  // For now, log them
  console.log('Parsed Tier 2 events:', parsed.events.length);

  updateProgress(fileId, { progress: 100 });
}

/**
 * Process Tier 3 (Testimonials) file
 */
async function processTier3(
  caseId: string,
  cf: ClassifiedFile,
  fileId: string,
  updateProgress: (id: string, updates: Partial<UploadProgress>) => void
): Promise<void> {
  updateProgress(fileId, { progress: 50, status: 'processing' });

  const statement = await parseTier3File(cf.file);

  // Submit witness statement
  const result = await api.submitWitnessStatements(caseId, [statement]);

  updateProgress(fileId, {
    progress: 100,
    jobId: result.profile_job_id,
  });
}
