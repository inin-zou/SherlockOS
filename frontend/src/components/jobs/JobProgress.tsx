'use client';

import { useState } from 'react';
import {
  Loader2,
  CheckCircle,
  XCircle,
  Clock,
  ChevronDown,
  ChevronUp,
  Box,
  Brain,
  User,
  Image,
  Video,
  FileOutput,
  Scan,
} from 'lucide-react';
import { cn, formatDuration } from '@/lib/utils';
import type { Job, JobType, JobStatus } from '@/lib/types';

interface JobProgressProps {
  jobs: Job[];
  className?: string;
}

const JOB_TYPE_CONFIG: Record<
  JobType,
  { label: string; icon: React.ComponentType<{ className?: string; style?: React.CSSProperties }>; color: string }
> = {
  reconstruction: { label: 'Reconstruction', icon: Box, color: '#3b82f6' },
  reasoning: { label: 'Reasoning', icon: Brain, color: '#8b5cf6' },
  profile: { label: 'Profile', icon: User, color: '#10b981' },
  imagegen: { label: 'Image Gen', icon: Image, color: '#f59e0b' },
  replay: { label: 'Replay', icon: Video, color: '#ef4444' },
  export: { label: 'Export', icon: FileOutput, color: '#06b6d4' },
  scene_analysis: { label: 'Scene Analysis', icon: Scan, color: '#ec4899' },
  asset3d: { label: '3D Asset', icon: Box, color: '#84cc16' },
};

const STATUS_CONFIG: Record<
  JobStatus,
  { label: string; icon: React.ComponentType<{ className?: string; style?: React.CSSProperties }>; color: string }
> = {
  queued: { label: 'Queued', icon: Clock, color: '#606068' },
  running: { label: 'Running', icon: Loader2, color: '#3b82f6' },
  done: { label: 'Done', icon: CheckCircle, color: '#22c55e' },
  failed: { label: 'Failed', icon: XCircle, color: '#ef4444' },
  canceled: { label: 'Canceled', icon: XCircle, color: '#606068' },
};

export function JobProgress({ jobs, className }: JobProgressProps) {
  const [isExpanded, setIsExpanded] = useState(true);

  const activeJobs = jobs.filter((j) => j.status === 'queued' || j.status === 'running');
  const completedJobs = jobs.filter((j) => j.status === 'done' || j.status === 'failed');

  if (jobs.length === 0) return null;

  return (
    <div className={cn('bg-[#111114] border border-[#2a2a32] rounded-lg overflow-hidden', className)}>
      {/* Header */}
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-center justify-between px-4 py-3 hover:bg-[#1f1f24] transition-colors"
      >
        <div className="flex items-center gap-3">
          {activeJobs.length > 0 ? (
            <Loader2 className="w-4 h-4 text-[#3b82f6] animate-spin" />
          ) : (
            <CheckCircle className="w-4 h-4 text-[#22c55e]" />
          )}
          <span className="text-sm font-medium text-[#f0f0f2]">
            {activeJobs.length > 0
              ? `${activeJobs.length} job${activeJobs.length > 1 ? 's' : ''} processing`
              : 'All jobs complete'}
          </span>
        </div>
        {isExpanded ? (
          <ChevronUp className="w-4 h-4 text-[#606068]" />
        ) : (
          <ChevronDown className="w-4 h-4 text-[#606068]" />
        )}
      </button>

      {/* Job list */}
      {isExpanded && (
        <div className="border-t border-[#2a2a32] divide-y divide-[#1e1e24]">
          {jobs.map((job) => (
            <JobItem key={job.id} job={job} />
          ))}
        </div>
      )}
    </div>
  );
}

function JobItem({ job }: { job: Job }) {
  const typeConfig = JOB_TYPE_CONFIG[job.type] || {
    label: job.type,
    icon: Box,
    color: '#606068',
  };
  const statusConfig = STATUS_CONFIG[job.status];

  const TypeIcon = typeConfig.icon;
  const StatusIcon = statusConfig.icon;

  const isActive = job.status === 'queued' || job.status === 'running';

  return (
    <div className="px-4 py-3 flex items-center gap-4">
      {/* Type icon */}
      <div
        className="w-8 h-8 rounded-lg flex items-center justify-center"
        style={{ backgroundColor: `${typeConfig.color}20` }}
      >
        <TypeIcon className="w-4 h-4" style={{ color: typeConfig.color }} />
      </div>

      {/* Info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-[#f0f0f2]">{typeConfig.label}</span>
          <span
            className="text-xs px-1.5 py-0.5 rounded"
            style={{
              backgroundColor: `${statusConfig.color}20`,
              color: statusConfig.color,
            }}
          >
            {statusConfig.label}
          </span>
        </div>
        {job.error && (
          <p className="text-xs text-[#ef4444] mt-0.5 truncate">{job.error}</p>
        )}
      </div>

      {/* Progress */}
      {isActive && (
        <div className="flex items-center gap-3">
          <div className="w-24 h-1.5 bg-[#2a2a32] rounded-full overflow-hidden">
            <div
              className="h-full bg-[#3b82f6] transition-all duration-300"
              style={{ width: `${job.progress}%` }}
            />
          </div>
          <span className="text-xs text-[#606068] w-8">{job.progress}%</span>
        </div>
      )}

      {/* Status icon */}
      <StatusIcon
        className={cn('w-4 h-4', isActive && 'animate-spin')}
        style={{ color: statusConfig.color }}
      />
    </div>
  );
}

/**
 * Compact job indicator for header
 */
export function JobIndicator({
  activeCount,
  onClick,
}: {
  activeCount: number;
  onClick?: () => void;
}) {
  if (activeCount === 0) return null;

  return (
    <button
      onClick={onClick}
      className={cn(
        'flex items-center gap-2 px-3 py-1.5 rounded-lg',
        'bg-[#3b82f6]/10 hover:bg-[#3b82f6]/20 transition-colors'
      )}
    >
      <Loader2 className="w-3.5 h-3.5 text-[#3b82f6] animate-spin" />
      <span className="text-xs font-medium text-[#3b82f6]">
        {activeCount} job{activeCount > 1 ? 's' : ''}
      </span>
    </button>
  );
}
