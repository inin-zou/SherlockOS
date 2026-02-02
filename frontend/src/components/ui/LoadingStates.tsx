'use client';

import { Loader2, FolderOpen, FileQuestion, Brain, Route, User } from 'lucide-react';
import { cn } from '@/lib/utils';

interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function LoadingSpinner({ size = 'md', className }: LoadingSpinnerProps) {
  const sizeClasses = {
    sm: 'w-4 h-4',
    md: 'w-6 h-6',
    lg: 'w-8 h-8',
  };

  return (
    <Loader2
      className={cn(
        sizeClasses[size],
        'text-[#3b82f6] animate-spin',
        className
      )}
    />
  );
}

interface LoadingStateProps {
  message?: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function LoadingState({ message = 'Loading...', size = 'md', className }: LoadingStateProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center gap-3 p-6', className)}>
      <LoadingSpinner size={size} />
      <span className="text-sm text-[#606068]">{message}</span>
    </div>
  );
}

interface SkeletonProps {
  className?: string;
}

export function Skeleton({ className }: SkeletonProps) {
  return (
    <div
      className={cn(
        'bg-[#1f1f24] rounded animate-pulse',
        className
      )}
    />
  );
}

export function EvidenceListSkeleton() {
  return (
    <div className="space-y-3 p-3">
      {[1, 2, 3].map((i) => (
        <div key={i} className="space-y-2">
          <Skeleton className="h-4 w-24" />
          <div className="space-y-1.5">
            {[1, 2].map((j) => (
              <div key={j} className="flex items-center gap-2 p-2 bg-[#1f1f24] rounded">
                <Skeleton className="w-8 h-8 rounded" />
                <div className="flex-1 space-y-1">
                  <Skeleton className="h-3 w-3/4" />
                  <Skeleton className="h-2 w-1/2" />
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

export function TimelineSkeleton() {
  return (
    <div className="flex gap-3 p-3 overflow-x-auto">
      {[1, 2, 3, 4, 5].map((i) => (
        <div key={i} className="flex-shrink-0 w-32 space-y-2">
          <Skeleton className="h-16 w-full rounded-lg" />
          <Skeleton className="h-3 w-3/4" />
          <Skeleton className="h-2 w-1/2" />
        </div>
      ))}
    </div>
  );
}

export function ProfileSkeleton() {
  return (
    <div className="space-y-4 p-4">
      <Skeleton className="aspect-square w-full rounded-lg" />
      <div className="space-y-3">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="flex justify-between items-center">
            <Skeleton className="h-3 w-16" />
            <Skeleton className="h-3 w-24" />
          </div>
        ))}
      </div>
    </div>
  );
}

interface EmptyStateProps {
  icon?: React.ComponentType<{ className?: string }>;
  title: string;
  description?: string;
  action?: React.ReactNode;
  className?: string;
}

export function EmptyState({
  icon: Icon = FileQuestion,
  title,
  description,
  action,
  className,
}: EmptyStateProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center p-6 text-center', className)}>
      <div className="w-12 h-12 rounded-xl bg-[#1f1f24] flex items-center justify-center mb-4">
        <Icon className="w-6 h-6 text-[#606068]" />
      </div>
      <h4 className="text-sm font-medium text-[#a0a0a8] mb-1">{title}</h4>
      {description && (
        <p className="text-xs text-[#606068] max-w-sm mb-4">{description}</p>
      )}
      {action}
    </div>
  );
}

export function EvidenceEmptyState() {
  return (
    <EmptyState
      icon={FolderOpen}
      title="No evidence yet"
      description="Upload crime scene photos, videos, or documents to begin the investigation"
    />
  );
}

export function TimelineEmptyState() {
  return (
    <EmptyState
      icon={FileQuestion}
      title="No timeline events"
      description="Events will appear here as you add evidence and run analysis"
      className="py-8"
    />
  );
}

export function TrajectoriesEmptyState() {
  return (
    <EmptyState
      icon={Route}
      title="No trajectories"
      description="Run the reasoning engine to generate movement hypotheses"
    />
  );
}

export function ProfileEmptyState() {
  return (
    <EmptyState
      icon={User}
      title="No suspect profile"
      description="Add witness statements to build a suspect profile"
    />
  );
}

export function ReasoningEmptyState() {
  return (
    <EmptyState
      icon={Brain}
      title="No analysis yet"
      description="Run reasoning to analyze evidence and generate insights"
    />
  );
}
