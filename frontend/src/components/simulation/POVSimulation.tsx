'use client';

import { useState, useCallback } from 'react';
import {
  Play,
  Pause,
  RotateCcw,
  Send,
  Footprints,
  Clock,
  MapPin,
  Loader2,
  ChevronDown,
  ChevronUp,
  Sparkles,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { cn } from '@/lib/utils';
import { useStore } from '@/lib/store';
import * as api from '@/lib/api';

interface POVSimulationProps {
  caseId: string;
  className?: string;
}

interface MotionPoint {
  id: string;
  position: [number, number, number];
  timestamp: string;
  action: string;
  confidence: number;
}

interface GeneratedPath {
  id: string;
  points: MotionPoint[];
  duration: string;
  description: string;
}

export function POVSimulation({ caseId, className }: POVSimulationProps) {
  const [prompt, setPrompt] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [generatedPaths, setGeneratedPaths] = useState<GeneratedPath[]>([]);
  const [selectedPathId, setSelectedPathId] = useState<string | null>(null);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playbackProgress, setPlaybackProgress] = useState(0);
  const [expandedPathId, setExpandedPathId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const { setTrajectories, setSelectedTrajectoryId } = useStore();

  const handleGenerate = useCallback(async () => {
    if (!prompt.trim()) return;

    setIsGenerating(true);
    setError(null);

    try {
      // Create a simulation job
      const job = await api.createJob(caseId, 'reasoning', {
        simulation_prompt: prompt,
        mode: 'text_to_motion',
      });

      // Poll for completion
      let attempts = 0;
      const maxAttempts = 60;

      while (attempts < maxAttempts) {
        await new Promise((r) => setTimeout(r, 2000));
        const updatedJob = await api.getJob(job.id);

        if (updatedJob.status === 'done') {
          // Parse the output into paths
          const output = updatedJob.output as { trajectories?: any[] };
          if (output?.trajectories) {
            const paths: GeneratedPath[] = output.trajectories.map((t: any, i: number) => ({
              id: `path-${i}`,
              points: t.segments?.flatMap((s: any) => [
                {
                  id: `${i}-start`,
                  position: s.from_position,
                  timestamp: s.time_estimate?.start || '',
                  action: 'move',
                  confidence: s.confidence,
                },
                {
                  id: `${i}-end`,
                  position: s.to_position,
                  timestamp: s.time_estimate?.end || '',
                  action: s.explanation || 'arrive',
                  confidence: s.confidence,
                },
              ]) || [],
              duration: '~5 min',
              description: t.segments?.[0]?.explanation || 'Generated motion path',
            }));
            setGeneratedPaths(paths);
            if (paths.length > 0) {
              setSelectedPathId(paths[0].id);
            }
          }
          break;
        } else if (updatedJob.status === 'failed') {
          throw new Error(updatedJob.error || 'Generation failed');
        }

        attempts++;
      }

      if (attempts >= maxAttempts) {
        throw new Error('Generation timed out');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate motion path');
    } finally {
      setIsGenerating(false);
    }
  }, [caseId, prompt]);

  const handlePlayPause = useCallback(() => {
    setIsPlaying((prev) => !prev);
  }, []);

  const handleReset = useCallback(() => {
    setIsPlaying(false);
    setPlaybackProgress(0);
  }, []);

  const handleApplyPath = useCallback((path: GeneratedPath) => {
    // Convert to trajectory format and apply to store
    const trajectory = {
      id: path.id,
      rank: 1,
      overall_confidence: path.points.reduce((acc, p) => acc + p.confidence, 0) / path.points.length,
      segments: path.points.slice(0, -1).map((p, i) => ({
        id: `seg-${i}`,
        from_position: p.position,
        to_position: path.points[i + 1]?.position || p.position,
        confidence: p.confidence,
        explanation: p.action,
        evidence_refs: [],
      })),
    };
    setTrajectories([trajectory]);
    setSelectedTrajectoryId(path.id);
  }, [setTrajectories, setSelectedTrajectoryId]);

  const selectedPath = generatedPaths.find((p) => p.id === selectedPathId);

  return (
    <div className={cn('bg-[#111114] border border-[#2a2a32] rounded-lg overflow-hidden', className)}>
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-[#2a2a32]">
        <Footprints className="w-4 h-4 text-[#8b5cf6]" />
        <h3 className="text-sm font-medium text-[#f0f0f2]">POV Simulation</h3>
        <span className="ml-auto text-xs text-[#606068]">Text to Motion</span>
      </div>

      {/* Prompt Input */}
      <div className="p-4 border-b border-[#2a2a32]">
        <div className="relative">
          <textarea
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            placeholder="Describe the suspect's movement... e.g., 'Entered through the north window at 22:15, moved to the vault, spent 3 minutes, then exited via the back door'"
            rows={3}
            className={cn(
              'w-full px-3 py-2 bg-[#1f1f24] border border-[#2a2a32] rounded-lg',
              'text-sm text-[#f0f0f2] placeholder:text-[#606068]',
              'focus:outline-none focus:ring-1 focus:ring-[#8b5cf6]',
              'resize-none'
            )}
          />
          <Button
            variant="primary"
            size="sm"
            onClick={handleGenerate}
            isLoading={isGenerating}
            disabled={!prompt.trim() || isGenerating}
            className="absolute bottom-2 right-2"
          >
            <Sparkles className="w-4 h-4" />
            Generate
          </Button>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="px-4 py-2 bg-[#ef4444]/10 border-b border-[#ef4444]/20">
          <p className="text-xs text-[#ef4444]">{error}</p>
        </div>
      )}

      {/* Generated Paths */}
      {generatedPaths.length > 0 && (
        <div className="p-4 space-y-3">
          <h4 className="text-xs font-medium text-[#606068] uppercase tracking-wider">
            Generated Paths ({generatedPaths.length})
          </h4>

          {generatedPaths.map((path) => (
            <div
              key={path.id}
              className={cn(
                'bg-[#1f1f24] rounded-lg overflow-hidden transition-all',
                selectedPathId === path.id && 'ring-1 ring-[#8b5cf6]'
              )}
            >
              <button
                onClick={() => {
                  setSelectedPathId(path.id);
                  setExpandedPathId(expandedPathId === path.id ? null : path.id);
                }}
                className="w-full flex items-center gap-3 p-3 text-left hover:bg-[#2a2a32]/50 transition-colors"
              >
                <div className="w-8 h-8 rounded-lg bg-[#8b5cf6]/20 flex items-center justify-center">
                  <Footprints className="w-4 h-4 text-[#8b5cf6]" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-[#f0f0f2] truncate">{path.description}</p>
                  <div className="flex items-center gap-3 mt-0.5">
                    <span className="text-xs text-[#606068] flex items-center gap-1">
                      <MapPin className="w-3 h-3" />
                      {path.points.length} points
                    </span>
                    <span className="text-xs text-[#606068] flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {path.duration}
                    </span>
                  </div>
                </div>
                {expandedPathId === path.id ? (
                  <ChevronUp className="w-4 h-4 text-[#606068]" />
                ) : (
                  <ChevronDown className="w-4 h-4 text-[#606068]" />
                )}
              </button>

              {/* Expanded details */}
              {expandedPathId === path.id && (
                <div className="px-3 pb-3 space-y-2">
                  {/* Timeline preview */}
                  <div className="flex items-center gap-1">
                    {path.points.map((point, i) => (
                      <div key={point.id} className="flex items-center">
                        <div
                          className="w-2 h-2 rounded-full bg-[#8b5cf6]"
                          title={`${point.action} at ${point.timestamp}`}
                        />
                        {i < path.points.length - 1 && (
                          <div className="w-4 h-0.5 bg-[#8b5cf6]/30" />
                        )}
                      </div>
                    ))}
                  </div>

                  {/* Actions */}
                  <div className="flex items-center gap-2 pt-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={handleReset}
                      className="flex-1"
                    >
                      <RotateCcw className="w-3 h-3" />
                      Reset
                    </Button>
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={handlePlayPause}
                      className="flex-1"
                    >
                      {isPlaying ? (
                        <>
                          <Pause className="w-3 h-3" />
                          Pause
                        </>
                      ) : (
                        <>
                          <Play className="w-3 h-3" />
                          Preview
                        </>
                      )}
                    </Button>
                    <Button
                      variant="primary"
                      size="sm"
                      onClick={() => handleApplyPath(path)}
                      className="flex-1"
                    >
                      <Send className="w-3 h-3" />
                      Apply
                    </Button>
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Empty state */}
      {generatedPaths.length === 0 && !isGenerating && (
        <div className="p-8 text-center">
          <Footprints className="w-10 h-10 text-[#2a2a32] mx-auto mb-3" />
          <p className="text-sm text-[#606068]">
            Describe suspect movement to generate motion paths
          </p>
        </div>
      )}

      {/* Generating state */}
      {isGenerating && (
        <div className="p-8 text-center">
          <Loader2 className="w-8 h-8 text-[#8b5cf6] mx-auto mb-3 animate-spin" />
          <p className="text-sm text-[#a0a0a8]">Generating motion path...</p>
          <p className="text-xs text-[#606068] mt-1">Analyzing scene constraints</p>
        </div>
      )}
    </div>
  );
}
