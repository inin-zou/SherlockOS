'use client';

import { useState } from 'react';
import {
  Upload,
  MessageSquare,
  Edit,
  Box,
  User,
  Brain,
  FileOutput,
  Video,
  GitCommit,
  ChevronRight,
} from 'lucide-react';
import { cn, formatDate } from '@/lib/utils';
import type { Commit, CommitType } from '@/lib/types';

interface CommitTimelineProps {
  commits: Commit[];
  onCommitSelect?: (commit: Commit) => void;
  selectedCommitId?: string;
  className?: string;
}

const COMMIT_TYPE_CONFIG: Record<
  CommitType,
  { label: string; icon: React.ComponentType<{ className?: string; style?: React.CSSProperties }>; color: string }
> = {
  upload_scan: { label: 'Scan Upload', icon: Upload, color: '#3b82f6' },
  witness_statement: { label: 'Witness Statement', icon: MessageSquare, color: '#8b5cf6' },
  manual_edit: { label: 'Manual Edit', icon: Edit, color: '#f59e0b' },
  reconstruction_update: { label: 'Reconstruction', icon: Box, color: '#10b981' },
  profile_update: { label: 'Profile Update', icon: User, color: '#ec4899' },
  reasoning_result: { label: 'Reasoning', icon: Brain, color: '#6366f1' },
  export_report: { label: 'Export', icon: FileOutput, color: '#06b6d4' },
  replay_generated: { label: 'Replay', icon: Video, color: '#ef4444' },
};

export function CommitTimeline({
  commits,
  onCommitSelect,
  selectedCommitId,
  className,
}: CommitTimelineProps) {
  const [expandedId, setExpandedId] = useState<string | null>(null);

  // Demo commits if none provided
  const displayCommits: Commit[] =
    commits.length > 0
      ? commits
      : [
          {
            id: 'c1',
            case_id: 'demo',
            type: 'upload_scan',
            summary: 'Uploaded 3 scan images from north wing',
            payload: { images: 3 },
            created_at: new Date(Date.now() - 3600000).toISOString(),
          },
          {
            id: 'c2',
            case_id: 'demo',
            type: 'reconstruction_update',
            summary: 'Scene reconstructed with 12 objects detected',
            payload: { objects: 12 },
            created_at: new Date(Date.now() - 3000000).toISOString(),
          },
          {
            id: 'c3',
            case_id: 'demo',
            type: 'witness_statement',
            summary: 'Added statement from Security Guard A',
            payload: { source: 'Security Guard A', credibility: 0.8 },
            created_at: new Date(Date.now() - 2400000).toISOString(),
          },
          {
            id: 'c4',
            case_id: 'demo',
            type: 'reasoning_result',
            summary: 'Generated 3 trajectory hypotheses',
            payload: { trajectories: 3, top_confidence: 0.85 },
            created_at: new Date(Date.now() - 1800000).toISOString(),
          },
        ];

  return (
    <div className={cn('space-y-1', className)}>
      <h3 className="text-xs font-medium text-[#606068] uppercase tracking-wider px-2 mb-2">
        Commit History
      </h3>

      <div className="space-y-1">
        {displayCommits.map((commit, index) => {
          const config = COMMIT_TYPE_CONFIG[commit.type] || {
            label: commit.type,
            icon: GitCommit,
            color: '#606068',
          };
          const Icon = config.icon;
          const isSelected = commit.id === selectedCommitId;
          const isExpanded = commit.id === expandedId;

          return (
            <div key={commit.id}>
              <button
                onClick={() => {
                  onCommitSelect?.(commit);
                  setExpandedId(isExpanded ? null : commit.id);
                }}
                className={cn(
                  'w-full flex items-start gap-3 px-3 py-2 rounded-lg transition-all',
                  'hover:bg-[#1f1f24] text-left',
                  isSelected && 'bg-[#1f1f24] ring-1 ring-[#3b82f6]'
                )}
              >
                {/* Timeline connector */}
                <div className="flex flex-col items-center pt-1">
                  <div
                    className="w-6 h-6 rounded-full flex items-center justify-center"
                    style={{ backgroundColor: `${config.color}20` }}
                  >
                    <Icon className="w-3 h-3" style={{ color: config.color }} />
                  </div>
                  {index < displayCommits.length - 1 && (
                    <div className="w-px flex-1 bg-[#2a2a32] mt-1" />
                  )}
                </div>

                {/* Content */}
                <div className="flex-1 min-w-0 pb-2">
                  <div className="flex items-center gap-2">
                    <span className="text-xs font-medium" style={{ color: config.color }}>
                      {config.label}
                    </span>
                    <span className="text-xs text-[#606068]">
                      {formatDate(commit.created_at)}
                    </span>
                  </div>
                  <p className="text-sm text-[#a0a0a8] mt-0.5 line-clamp-2">
                    {commit.summary}
                  </p>

                  {/* Expanded details */}
                  {isExpanded && commit.payload && (
                    <div className="mt-2 p-2 bg-[#18181c] rounded text-xs font-mono text-[#606068]">
                      <pre className="whitespace-pre-wrap">
                        {JSON.stringify(commit.payload, null, 2)}
                      </pre>
                    </div>
                  )}
                </div>

                <ChevronRight
                  className={cn(
                    'w-4 h-4 text-[#606068] transition-transform mt-1',
                    isExpanded && 'rotate-90'
                  )}
                />
              </button>
            </div>
          );
        })}
      </div>
    </div>
  );
}

/**
 * Compact commit indicator for inline display
 */
export function CommitBadge({ commit }: { commit: Commit }) {
  const config = COMMIT_TYPE_CONFIG[commit.type] || {
    label: commit.type,
    icon: GitCommit,
    color: '#606068',
  };
  const Icon = config.icon;

  return (
    <div
      className="inline-flex items-center gap-1.5 px-2 py-1 rounded text-xs"
      style={{ backgroundColor: `${config.color}15`, color: config.color }}
    >
      <Icon className="w-3 h-3" />
      <span>{config.label}</span>
    </div>
  );
}
