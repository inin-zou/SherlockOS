'use client';

import { useEffect, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import dynamic from 'next/dynamic';
import { Header } from '@/components/layout/Header';
import { Sidebar } from '@/components/layout/Sidebar';
import { ModeSelector } from '@/components/layout/ModeSelector';
import { Timeline } from '@/components/timeline/Timeline';
import { ModePanel } from '@/components/panels/ModePanel';
import { JobProgress } from '@/components/jobs/JobProgress';
import { DropOverlay } from '@/components/evidence/DropZone';
import { useStore } from '@/lib/store';
import { useUpload } from '@/hooks/useUpload';
import { useJobs } from '@/hooks/useJobs';
import * as api from '@/lib/api';
import { ArrowLeft, Loader2, AlertCircle } from 'lucide-react';

// Dynamically import SceneViewer to avoid SSR issues with Three.js
const SceneViewer = dynamic(
  () => import('@/components/scene/SceneViewer').then((mod) => mod.SceneViewer),
  {
    ssr: false,
    loading: () => (
      <div className="w-full h-full flex items-center justify-center bg-[#0a0a0c]">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 border-2 border-[#3b82f6] border-t-transparent rounded-full animate-spin" />
          <span className="text-sm text-[#606068]">Loading 3D scene...</span>
        </div>
      </div>
    ),
  }
);

export default function CaseDetailPage() {
  const params = useParams();
  const router = useRouter();
  const caseId = params.id as string;

  const { currentCase, setCurrentCase, commits, setCommits, setSceneGraph, jobs, setJobs } = useStore();
  const [isDragOver, setIsDragOver] = useState(false);
  const [showJobPanel, setShowJobPanel] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Load case data on mount
  useEffect(() => {
    if (!caseId) return;

    const loadCase = async () => {
      setIsLoading(true);
      setError(null);

      try {
        // Load case info
        const caseData = await api.getCase(caseId);
        setCurrentCase(caseData);

        // Load timeline
        const timeline = await api.getTimeline(caseId);
        setCommits(timeline.commits || []);

        // Load scene snapshot
        try {
          const snapshot = await api.getSnapshot(caseId);
          setSceneGraph(snapshot.scenegraph);
        } catch {
          // Scene might not exist yet
          setSceneGraph(null);
        }
      } catch (err) {
        console.error('Failed to load case:', err);
        setError(err instanceof Error ? err.message : 'Failed to load case');
      } finally {
        setIsLoading(false);
      }
    };

    loadCase();
  }, [caseId, setCurrentCase, setCommits, setSceneGraph]);

  // Upload hook
  const { uploadFiles, progress, isUploading } = useUpload(caseId);

  // Jobs hook
  const { activeJobs, pollJob } = useJobs(caseId);

  // Track upload jobs
  useEffect(() => {
    progress.forEach((p) => {
      if (p.jobId && !jobs.find((j) => j.id === p.jobId)) {
        setJobs([...jobs, {
          id: p.jobId,
          type: 'reconstruction',
          status: 'queued',
          progress: 0,
          created_at: new Date().toISOString()
        } as any]);
        pollJob(p.jobId);
      }
    });
  }, [progress, jobs, setJobs, pollJob]);

  // Show job panel when jobs are active
  useEffect(() => {
    if (activeJobs.length > 0) {
      setShowJobPanel(true);
    }
  }, [activeJobs.length]);

  // Handle global drag events
  useEffect(() => {
    const handleDragEnter = (e: DragEvent) => {
      e.preventDefault();
      if (e.dataTransfer?.types.includes('Files')) {
        setIsDragOver(true);
      }
    };

    const handleDragLeave = (e: DragEvent) => {
      e.preventDefault();
      if (e.relatedTarget === null) {
        setIsDragOver(false);
      }
    };

    const handleDrop = (e: DragEvent) => {
      e.preventDefault();
      setIsDragOver(false);
    };

    window.addEventListener('dragenter', handleDragEnter);
    window.addEventListener('dragleave', handleDragLeave);
    window.addEventListener('drop', handleDrop);

    return () => {
      window.removeEventListener('dragenter', handleDragEnter);
      window.removeEventListener('dragleave', handleDragLeave);
      window.removeEventListener('drop', handleDrop);
    };
  }, []);

  const handleFileDrop = (files: FileList) => {
    setIsDragOver(false);
    uploadFiles(files);
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="h-screen flex flex-col bg-[#111114]">
        <Header activeJobCount={0} onJobsClick={() => {}} />
        <div className="flex-1 flex items-center justify-center">
          <div className="flex flex-col items-center gap-4">
            <Loader2 className="w-10 h-10 text-[#3b82f6] animate-spin" />
            <p className="text-[#8b8b96]">Loading case...</p>
          </div>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="h-screen flex flex-col bg-[#111114]">
        <Header activeJobCount={0} onJobsClick={() => {}} />
        <div className="flex-1 flex items-center justify-center">
          <div className="flex flex-col items-center gap-4 max-w-md text-center">
            <div className="w-16 h-16 rounded-full bg-[#ef4444]/10 flex items-center justify-center">
              <AlertCircle className="w-8 h-8 text-[#ef4444]" />
            </div>
            <h2 className="text-xl font-semibold text-white">Case Not Found</h2>
            <p className="text-[#8b8b96]">{error}</p>
            <button
              onClick={() => router.push('/')}
              className="flex items-center gap-2 px-4 py-2 bg-[#1f1f24] hover:bg-[#2a2a32] rounded-lg transition-colors text-sm"
            >
              <ArrowLeft className="w-4 h-4" />
              Back to Dashboard
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-screen flex flex-col overflow-hidden">
      {/* Header */}
      <Header activeJobCount={activeJobs.length} onJobsClick={() => setShowJobPanel(!showJobPanel)} />

      {/* Main content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar with upload */}
        <Sidebar
          caseId={caseId}
          onUpload={uploadFiles}
          uploadProgress={progress}
          isUploading={isUploading}
        />

        {/* Scene + Timeline */}
        <div className="flex-1 flex flex-col overflow-hidden relative">
          {/* Case Title Bar */}
          <div className="h-12 bg-[#0a0a0c] border-b border-[#1f1f24] flex items-center px-4 gap-3">
            <button
              onClick={() => router.push('/')}
              className="p-1.5 hover:bg-[#1f1f24] rounded-md transition-colors"
              title="Back to Dashboard"
            >
              <ArrowLeft className="w-4 h-4 text-[#606068]" />
            </button>
            <div className="h-4 w-px bg-[#1f1f24]" />
            <div className="flex-1 min-w-0">
              <h1 className="text-sm font-medium text-white truncate">
                {currentCase?.title || 'Untitled Case'}
              </h1>
              {currentCase?.description && (
                <p className="text-xs text-[#606068] truncate">{currentCase.description}</p>
              )}
            </div>
            <div className="flex items-center gap-2 text-xs text-[#606068]">
              <span>{commits.length} commits</span>
              <span className="w-1 h-1 rounded-full bg-[#606068]" />
              <span>ID: {caseId.slice(0, 8)}...</span>
            </div>
          </div>

          {/* Mode Selector */}
          <ModeSelector />

          {/* Job Progress Panel */}
          {showJobPanel && jobs.length > 0 && (
            <div className="absolute top-28 right-4 z-20 w-80">
              <JobProgress jobs={jobs} />
            </div>
          )}

          {/* 3D Scene Viewer */}
          <div className="flex-1 relative">
            <SceneViewer />
          </div>

          {/* Timeline */}
          <Timeline />
        </div>

        {/* Right Panel - Mode-specific */}
        <ModePanel caseId={caseId} />
      </div>

      {/* Full-screen drop overlay */}
      <DropOverlay isVisible={isDragOver} onDrop={handleFileDrop} />
    </div>
  );
}
