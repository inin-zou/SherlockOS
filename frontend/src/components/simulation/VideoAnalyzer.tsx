'use client';

import { useState, useCallback, useRef } from 'react';
import {
  Video,
  Upload,
  Play,
  Pause,
  SkipBack,
  SkipForward,
  Loader2,
  Eye,
  Clock,
  MapPin,
  ChevronRight,
  AlertTriangle,
  Check,
  X,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { cn } from '@/lib/utils';
import { useStore } from '@/lib/store';
import * as api from '@/lib/api';

interface VideoAnalyzerProps {
  caseId: string;
  className?: string;
}

interface DetectedMotion {
  id: string;
  timestamp: number;
  duration: number;
  type: 'person' | 'vehicle' | 'object';
  confidence: number;
  bbox: { x: number; y: number; width: number; height: number };
  path: [number, number][];
  description: string;
}

interface AnalysisResult {
  videoId: string;
  filename: string;
  duration: number;
  fps: number;
  motions: DetectedMotion[];
  keyFrames: number[];
}

type AnalysisStatus = 'idle' | 'uploading' | 'analyzing' | 'done' | 'error';

export function VideoAnalyzer({ caseId, className }: VideoAnalyzerProps) {
  const [status, setStatus] = useState<AnalysisStatus>('idle');
  const [progress, setProgress] = useState(0);
  const [result, setResult] = useState<AnalysisResult | null>(null);
  const [selectedMotionId, setSelectedMotionId] = useState<string | null>(null);
  const [videoTime, setVideoTime] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const { setTimelineTracks, addCommit } = useStore();

  const handleFileSelect = useCallback(async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file type
    if (!file.type.startsWith('video/')) {
      setError('Please select a video file');
      return;
    }

    setStatus('uploading');
    setProgress(0);
    setError(null);

    try {
      // Get upload intent
      const intent = await api.getUploadIntent(caseId, [
        {
          filename: file.name,
          content_type: file.type,
          size_bytes: file.size,
        },
      ]);

      setProgress(20);

      // Upload file
      const uploadInfo = intent.intents[0];
      if (uploadInfo) {
        await api.uploadFile(uploadInfo.presigned_url, file);
      }

      setProgress(40);
      setStatus('analyzing');

      // Create scene analysis job for video
      const job = await api.createJob(caseId, 'scene_analysis', {
        video_asset_key: uploadInfo?.storage_key,
        extract_motion: true,
        detect_persons: true,
      });

      // Poll for completion
      let attempts = 0;
      const maxAttempts = 120; // 4 minutes max

      while (attempts < maxAttempts) {
        await new Promise((r) => setTimeout(r, 2000));
        const updatedJob = await api.getJob(job.id);

        setProgress(40 + Math.min(updatedJob.progress * 0.5, 50));

        if (updatedJob.status === 'done') {
          const output = updatedJob.output as any;

          // Parse analysis results
          const analysisResult: AnalysisResult = {
            videoId: uploadInfo?.storage_key || '',
            filename: file.name,
            duration: output?.duration || 0,
            fps: output?.fps || 30,
            motions: (output?.detected_objects || []).map((obj: any, i: number) => ({
              id: `motion-${i}`,
              timestamp: obj.timestamp || i * 1000,
              duration: obj.duration || 1000,
              type: obj.type || 'person',
              confidence: obj.confidence || 0.8,
              bbox: obj.bbox || { x: 0, y: 0, width: 100, height: 100 },
              path: obj.path || [],
              description: obj.description || `Detected ${obj.type || 'motion'}`,
            })),
            keyFrames: output?.key_frames || [],
          };

          setResult(analysisResult);
          setProgress(100);
          setStatus('done');

          // Convert to timeline tracks
          const tracks = analysisResult.motions.map((motion) => ({
            id: motion.id,
            name: motion.description,
            type: motion.type === 'person' ? 'person' as const : 'event' as const,
            color: motion.type === 'person' ? '#8b5cf6' : '#f59e0b',
            events: [{
              id: `${motion.id}-event`,
              start: motion.timestamp,
              end: motion.timestamp + motion.duration,
              label: motion.description,
              confidence: motion.confidence,
            }],
          }));

          setTimelineTracks(tracks);
          break;
        } else if (updatedJob.status === 'failed') {
          throw new Error(updatedJob.error || 'Analysis failed');
        }

        attempts++;
      }

      if (attempts >= maxAttempts) {
        throw new Error('Analysis timed out');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to analyze video');
      setStatus('error');
    }

    // Reset input
    e.target.value = '';
  }, [caseId, setTimelineTracks]);

  const handleSeek = useCallback((timestamp: number) => {
    setVideoTime(timestamp);
  }, []);

  const selectedMotion = result?.motions.find((m) => m.id === selectedMotionId);

  return (
    <div className={cn('bg-[#111114] border border-[#2a2a32] rounded-lg overflow-hidden', className)}>
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-[#2a2a32]">
        <Video className="w-4 h-4 text-[#3b82f6]" />
        <h3 className="text-sm font-medium text-[#f0f0f2]">Video Analyzer</h3>
        <span className="ml-auto text-xs text-[#606068]">CCTV to Motion</span>
      </div>

      {/* Upload Zone */}
      {status === 'idle' && (
        <div className="p-4">
          <input
            ref={fileInputRef}
            type="file"
            accept="video/*"
            onChange={handleFileSelect}
            className="hidden"
          />
          <button
            onClick={() => fileInputRef.current?.click()}
            className={cn(
              'w-full border-2 border-dashed border-[#2a2a32] rounded-lg p-8',
              'flex flex-col items-center justify-center gap-3',
              'hover:border-[#3b82f6]/50 hover:bg-[#1f1f24]/50',
              'transition-all cursor-pointer'
            )}
          >
            <div className="w-12 h-12 rounded-xl bg-[#3b82f6]/20 flex items-center justify-center">
              <Upload className="w-6 h-6 text-[#3b82f6]" />
            </div>
            <div className="text-center">
              <p className="text-sm text-[#f0f0f2]">Upload CCTV footage</p>
              <p className="text-xs text-[#606068] mt-1">MP4, MOV, AVI supported</p>
            </div>
          </button>
        </div>
      )}

      {/* Processing State */}
      {(status === 'uploading' || status === 'analyzing') && (
        <div className="p-8 text-center">
          <Loader2 className="w-10 h-10 text-[#3b82f6] mx-auto mb-4 animate-spin" />
          <p className="text-sm text-[#f0f0f2]">
            {status === 'uploading' ? 'Uploading video...' : 'Analyzing footage...'}
          </p>
          <div className="w-48 h-1.5 bg-[#2a2a32] rounded-full mx-auto mt-3 overflow-hidden">
            <div
              className="h-full bg-[#3b82f6] transition-all duration-300"
              style={{ width: `${progress}%` }}
            />
          </div>
          <p className="text-xs text-[#606068] mt-2">{Math.round(progress)}%</p>
        </div>
      )}

      {/* Error State */}
      {status === 'error' && (
        <div className="p-4">
          <div className="bg-[#ef4444]/10 border border-[#ef4444]/20 rounded-lg p-4 text-center">
            <AlertTriangle className="w-8 h-8 text-[#ef4444] mx-auto mb-2" />
            <p className="text-sm text-[#ef4444]">{error}</p>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                setStatus('idle');
                setError(null);
              }}
              className="mt-3"
            >
              Try Again
            </Button>
          </div>
        </div>
      )}

      {/* Results */}
      {status === 'done' && result && (
        <div className="divide-y divide-[#2a2a32]">
          {/* Video Info */}
          <div className="p-4 bg-[#1f1f24]/50">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-[#22c55e]/20 flex items-center justify-center">
                <Check className="w-5 h-5 text-[#22c55e]" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm text-[#f0f0f2] truncate">{result.filename}</p>
                <p className="text-xs text-[#606068]">
                  {Math.round(result.duration / 1000)}s at {result.fps}fps
                </p>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => {
                  setStatus('idle');
                  setResult(null);
                }}
              >
                <X className="w-4 h-4" />
              </Button>
            </div>
          </div>

          {/* Detected Motions */}
          <div className="p-4">
            <h4 className="text-xs font-medium text-[#606068] uppercase tracking-wider mb-3">
              Detected Motion ({result.motions.length})
            </h4>

            <div className="space-y-2 max-h-64 overflow-y-auto">
              {result.motions.map((motion) => (
                <button
                  key={motion.id}
                  onClick={() => {
                    setSelectedMotionId(motion.id);
                    handleSeek(motion.timestamp);
                  }}
                  className={cn(
                    'w-full flex items-center gap-3 p-3 rounded-lg text-left transition-all',
                    'hover:bg-[#1f1f24]',
                    selectedMotionId === motion.id && 'bg-[#1f1f24] ring-1 ring-[#3b82f6]'
                  )}
                >
                  <div
                    className={cn(
                      'w-8 h-8 rounded-lg flex items-center justify-center',
                      motion.type === 'person' ? 'bg-[#8b5cf6]/20' : 'bg-[#f59e0b]/20'
                    )}
                  >
                    <Eye
                      className={cn(
                        'w-4 h-4',
                        motion.type === 'person' ? 'text-[#8b5cf6]' : 'text-[#f59e0b]'
                      )}
                    />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm text-[#f0f0f2] truncate">{motion.description}</p>
                    <div className="flex items-center gap-3 mt-0.5">
                      <span className="text-xs text-[#606068] flex items-center gap-1">
                        <Clock className="w-3 h-3" />
                        {Math.round(motion.timestamp / 1000)}s
                      </span>
                      <span className="text-xs text-[#606068]">
                        {Math.round(motion.confidence * 100)}% conf
                      </span>
                    </div>
                  </div>
                  <ChevronRight className="w-4 h-4 text-[#606068]" />
                </button>
              ))}
            </div>

            {result.motions.length === 0 && (
              <div className="text-center py-8">
                <Eye className="w-8 h-8 text-[#2a2a32] mx-auto mb-2" />
                <p className="text-sm text-[#606068]">No motion detected</p>
              </div>
            )}
          </div>

          {/* Selected Motion Details */}
          {selectedMotion && (
            <div className="p-4 bg-[#1f1f24]/50">
              <h4 className="text-xs font-medium text-[#606068] uppercase tracking-wider mb-3">
                Motion Path
              </h4>
              <div className="flex items-center gap-2 flex-wrap">
                {selectedMotion.path.map((point, i) => (
                  <div key={i} className="flex items-center">
                    <div
                      className="w-6 h-6 rounded-full bg-[#3b82f6]/20 border border-[#3b82f6] flex items-center justify-center"
                      title={`Point ${i + 1}: (${point[0]}, ${point[1]})`}
                    >
                      <span className="text-[10px] text-[#3b82f6]">{i + 1}</span>
                    </div>
                    {i < selectedMotion.path.length - 1 && (
                      <div className="w-6 h-0.5 bg-[#3b82f6]/30" />
                    )}
                  </div>
                ))}
              </div>
              <Button
                variant="primary"
                size="sm"
                className="w-full mt-3"
                onClick={() => {
                  // Apply motion path to 3D scene
                  console.log('Applying motion path:', selectedMotion);
                }}
              >
                <MapPin className="w-4 h-4" />
                Apply to Scene
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
