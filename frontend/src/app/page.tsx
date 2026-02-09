'use client';

import { useEffect, useState, useRef } from 'react';
import dynamic from 'next/dynamic';
import { Header } from '@/components/layout/Header';
import { Sidebar } from '@/components/layout/Sidebar';
import { NavigationRail } from '@/components/layout/NavigationRail'; // Import NavigationRail
import { ModeSelector } from '@/components/layout/ModeSelector';
import { Timeline } from '@/components/timeline/Timeline';
import { ModePanel } from '@/components/panels/ModePanel';
import { JobProgress } from '@/components/jobs/JobProgress';
import { DropOverlay } from '@/components/evidence/DropZone';
import { ErrorBoundary, SceneErrorBoundary, CompactErrorBoundary } from '@/components/ui/ErrorBoundary';
import { useStore } from '@/lib/store';
import { useUpload } from '@/hooks/useUpload';
import { useJobs } from '@/hooks/useJobs';
import * as api from '@/lib/api';

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

export default function DashboardPage() {
  const { currentCase, setCurrentCase, setCases, setCommits, setSceneGraph, jobs, setJobs } = useStore();
  const [isDragOver, setIsDragOver] = useState(false);
  const [showJobPanel, setShowJobPanel] = useState(false);

  // Initialize case on mount (use first case or create demo)
  useEffect(() => {
    const initCase = async () => {
      try {
        const cases = await api.getCases();
        setCases(cases); // Store cases list for case switching
        if (cases.length > 0) {
          setCurrentCase(cases[0]);
        } else {
          // Create a demo case
          const newCase = await api.createCase('Demo Investigation', 'Auto-created demo case');
          setCases([newCase]);
          setCurrentCase(newCase);
        }
      } catch (err) {
        console.error('Failed to initialize case:', err);
      }
    };

    if (!currentCase) {
      initCase();
    }
  }, [currentCase, setCurrentCase, setCases]);

  // Load case data when case changes
  useEffect(() => {
    if (!currentCase?.id) return;

    const loadCaseData = async () => {
      try {
        // Load timeline
        const timeline = await api.getTimeline(currentCase.id);
        setCommits(timeline.commits);

        // Load scene snapshot
        const snapshot = await api.getSnapshot(currentCase.id);
        setSceneGraph(snapshot.scenegraph);
      } catch (err) {
        console.error('Failed to load case data:', err);
      }
    };

    loadCaseData();
  }, [currentCase?.id, setCommits, setSceneGraph]);

  // Upload hook
  const { uploadFiles, progress, isUploading } = useUpload(currentCase?.id);

  // Jobs hook
  const { activeJobs, pollJob } = useJobs(currentCase?.id);

  // Track upload jobs
  useEffect(() => {
    progress.forEach((p) => {
      if (p.jobId && !jobs.find((j) => j.id === p.jobId)) {
        // Add job to store and start polling
        const now = new Date().toISOString();
        setJobs([...jobs, {
          id: p.jobId,
          case_id: currentCase?.id || '',
          type: 'reconstruction',
          status: 'queued',
          progress: 0,
          input: {},
          created_at: now,
          updated_at: now,
        }]);
        pollJob(p.jobId);
      }
    });
  }, [progress, jobs, setJobs, pollJob, currentCase?.id]);

  // Show job panel automatically when jobs become active
  // Track previous count to only trigger on 0 -> >0 transition
  const prevJobCountRef = useRef(0);
  useEffect(() => {
    const prevCount = prevJobCountRef.current;
    const currentCount = activeJobs.length;
    prevJobCountRef.current = currentCount;

    // Only auto-show when transitioning from no jobs to having jobs
    if (prevCount === 0 && currentCount > 0) {
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

  return (
    <div className="h-screen flex bg-[#0a0a0c] overflow-hidden">
      {/* Navigation Rail */}
      <NavigationRail />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
        {/* Header */}
        <Header activeJobCount={activeJobs.length} onJobsClick={() => setShowJobPanel(!showJobPanel)} />

        {/* Content with Sidebar */}
        <div className="flex-1 flex overflow-hidden">
          {/* Sidebar with upload */}
          <ErrorBoundary name="Sidebar">
            <Sidebar
              caseId={currentCase?.id}
              onUpload={uploadFiles}
              uploadProgress={progress}
              isUploading={isUploading}
            />
          </ErrorBoundary>

          {/* Scene + Timeline */}
          <div className="flex-1 flex flex-col overflow-hidden relative">
            {/* Mode Selector */}
            <CompactErrorBoundary name="ModeSelector">
              <ModeSelector />
            </CompactErrorBoundary>

            {/* Job Progress Panel */}
            {showJobPanel && jobs.length > 0 && (
              <div className="absolute top-16 right-4 z-20 w-80">
                <CompactErrorBoundary name="JobProgress">
                  <JobProgress jobs={jobs} />
                </CompactErrorBoundary>
              </div>
            )}

            {/* 3D Scene Viewer */}
            <div className="flex-1 relative">
              <SceneErrorBoundary name="SceneViewer">
                <SceneViewer />
              </SceneErrorBoundary>
            </div>

            {/* Timeline */}
            <CompactErrorBoundary name="Timeline">
              <Timeline />
            </CompactErrorBoundary>
          </div>

          {/* Right Panel - Mode-specific */}
          <ErrorBoundary name="ModePanel">
            <ModePanel caseId={currentCase?.id} />
          </ErrorBoundary>
        </div>
      </div>

      {/* Full-screen drop overlay */}
      <DropOverlay isVisible={isDragOver} onDrop={handleFileDrop} />
    </div>
  );
}
