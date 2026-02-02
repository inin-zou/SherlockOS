'use client';

import { useState, useCallback, useRef } from 'react';
import { Upload, File, X, CheckCircle, AlertCircle, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { TIER_CONFIG, type EvidenceTier } from '@/lib/fileClassifier';
import type { UploadProgress } from '@/hooks/useUpload';

interface DropZoneProps {
  onFilesDropped: (files: FileList) => void;
  progress: UploadProgress[];
  isUploading: boolean;
  disabled?: boolean;
  className?: string;
}

export function DropZone({
  onFilesDropped,
  progress,
  isUploading,
  disabled,
  className,
}: DropZoneProps) {
  const [isDragging, setIsDragging] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (!disabled) {
      setIsDragging(true);
    }
  }, [disabled]);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setIsDragging(false);

      if (disabled) return;

      const files = e.dataTransfer.files;
      if (files.length > 0) {
        onFilesDropped(files);
      }
    },
    [disabled, onFilesDropped]
  );

  const handleClick = useCallback(() => {
    if (!disabled) {
      fileInputRef.current?.click();
    }
  }, [disabled]);

  const handleFileInput = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const files = e.target.files;
      if (files && files.length > 0) {
        onFilesDropped(files);
      }
      // Reset input
      e.target.value = '';
    },
    [onFilesDropped]
  );

  const hasProgress = progress.length > 0;

  return (
    <div className={cn('space-y-3', className)}>
      {/* Drop area */}
      <div
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={handleClick}
        className={cn(
          'relative border-2 border-dashed rounded-lg p-6',
          'flex flex-col items-center justify-center gap-3',
          'cursor-pointer transition-all duration-200',
          isDragging
            ? 'border-[#3b82f6] bg-[#3b82f6]/10'
            : 'border-[#2a2a32] hover:border-[#3b82f6]/50 hover:bg-[#1f1f24]/50',
          disabled && 'opacity-50 cursor-not-allowed',
          hasProgress && 'pb-2'
        )}
      >
        <input
          ref={fileInputRef}
          type="file"
          multiple
          onChange={handleFileInput}
          className="hidden"
          disabled={disabled}
        />

        {isUploading ? (
          <Loader2 className="w-8 h-8 text-[#3b82f6] animate-spin" />
        ) : (
          <Upload
            className={cn(
              'w-8 h-8 transition-colors',
              isDragging ? 'text-[#3b82f6]' : 'text-[#606068]'
            )}
          />
        )}

        <div className="text-center">
          <p className="text-sm text-[#a0a0a8]">
            {isDragging ? 'Drop files here' : 'Drop evidence files or click to browse'}
          </p>
          <p className="text-xs text-[#606068] mt-1">
            Auto-classified into Environment, Ground Truth, Logs, or Testimonials
          </p>
        </div>

        {/* Tier legend */}
        <div className="flex flex-wrap gap-2 mt-2 justify-center">
          {([0, 1, 2, 3] as EvidenceTier[]).map((tier) => (
            <div
              key={tier}
              className="flex items-center gap-1.5 text-xs text-[#606068]"
            >
              <div
                className="w-2 h-2 rounded-full"
                style={{ backgroundColor: TIER_CONFIG[tier].color }}
              />
              <span>{TIER_CONFIG[tier].name}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Progress list */}
      {hasProgress && (
        <div className="space-y-2 max-h-48 overflow-y-auto">
          {progress.map((p) => (
            <ProgressItem key={p.fileId} progress={p} />
          ))}
        </div>
      )}
    </div>
  );
}

function ProgressItem({ progress }: { progress: UploadProgress }) {
  const tierConfig = TIER_CONFIG[progress.tier];

  const statusIcon = {
    pending: <File className="w-4 h-4 text-[#606068]" />,
    uploading: <Loader2 className="w-4 h-4 text-[#3b82f6] animate-spin" />,
    processing: <Loader2 className="w-4 h-4 text-[#f59e0b] animate-spin" />,
    done: <CheckCircle className="w-4 h-4 text-[#22c55e]" />,
    error: <AlertCircle className="w-4 h-4 text-[#ef4444]" />,
  };

  return (
    <div className="flex items-center gap-3 px-3 py-2 bg-[#1f1f24] rounded-lg">
      {/* Tier indicator */}
      <div
        className="w-1 h-8 rounded-full"
        style={{ backgroundColor: tierConfig.color }}
      />

      {/* File info */}
      <div className="flex-1 min-w-0">
        <p className="text-sm text-[#f0f0f2] truncate">{progress.filename}</p>
        <div className="flex items-center gap-2 mt-0.5">
          <span className="text-xs text-[#606068]">{tierConfig.name}</span>
          {progress.status === 'error' && progress.error && (
            <span className="text-xs text-[#ef4444]">{progress.error}</span>
          )}
        </div>
      </div>

      {/* Progress bar */}
      {(progress.status === 'uploading' || progress.status === 'processing') && (
        <div className="w-20 h-1.5 bg-[#2a2a32] rounded-full overflow-hidden">
          <div
            className="h-full bg-[#3b82f6] transition-all duration-300"
            style={{ width: `${progress.progress}%` }}
          />
        </div>
      )}

      {/* Status icon */}
      {statusIcon[progress.status]}
    </div>
  );
}

/**
 * Full-screen drop overlay for drag events
 */
export function DropOverlay({
  isVisible,
  onDrop,
}: {
  isVisible: boolean;
  onDrop: (files: FileList) => void;
}) {
  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      const files = e.dataTransfer.files;
      if (files.length > 0) {
        onDrop(files);
      }
    },
    [onDrop]
  );

  if (!isVisible) return null;

  return (
    <div
      onDragOver={handleDragOver}
      onDrop={handleDrop}
      className="fixed inset-0 z-50 bg-[#0a0a0c]/90 backdrop-blur-sm flex items-center justify-center"
    >
      <div className="text-center">
        <div className="w-20 h-20 mx-auto mb-4 rounded-2xl bg-[#3b82f6]/20 border-2 border-dashed border-[#3b82f6] flex items-center justify-center">
          <Upload className="w-10 h-10 text-[#3b82f6]" />
        </div>
        <h3 className="text-xl font-medium text-[#f0f0f2]">Drop files to upload</h3>
        <p className="text-sm text-[#a0a0a8] mt-2">
          Files will be auto-classified into evidence tiers
        </p>
      </div>
    </div>
  );
}
