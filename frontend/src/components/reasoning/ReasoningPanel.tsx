'use client';

import { useState, useCallback, useEffect } from 'react';
import {
  Brain,
  Play,
  Loader2,
  CheckCircle,
  AlertTriangle,
  ChevronDown,
  ChevronUp,
  Lightbulb,
  Route,
  Eye,
  Clock,
  Sparkles,
  Settings,
} from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { cn, formatConfidence } from '@/lib/utils';
import { useStore } from '@/lib/store';
import * as api from '@/lib/api';
import type { Trajectory, TrajectorySegment } from '@/lib/types';

interface ReasoningPanelProps {
  caseId: string;
  className?: string;
}

type ReasoningStatus = 'idle' | 'thinking' | 'done' | 'error';

interface ThinkingStep {
  id: string;
  phase: string;
  description: string;
  status: 'pending' | 'active' | 'done';
  duration?: number;
}

interface Discrepancy {
  id: string;
  type: 'timeline_conflict' | 'line_of_sight' | 'physical_impossible' | 'testimony_mismatch';
  severity: 'low' | 'medium' | 'high';
  description: string;
  sources: string[];
  evidence: string[];
}

export function ReasoningPanel({ caseId, className }: ReasoningPanelProps) {
  const [status, setStatus] = useState<ReasoningStatus>('idle');
  const [thinkingSteps, setThinkingSteps] = useState<ThinkingStep[]>([]);
  const [discrepancies, setDiscrepancies] = useState<Discrepancy[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [jobId, setJobId] = useState<string | null>(null);
  const [showSettings, setShowSettings] = useState(false);
  const [thinkingBudget, setThinkingBudget] = useState(10000);
  const [maxTrajectories, setMaxTrajectories] = useState(5);

  const {
    trajectories,
    setTrajectories,
    selectedTrajectoryId,
    setSelectedTrajectoryId,
    setViewMode,
  } = useStore();

  const initializeThinkingSteps = useCallback(() => {
    setThinkingSteps([
      { id: '1', phase: 'Gathering', description: 'Collecting all evidence and witness statements', status: 'pending' },
      { id: '2', phase: 'Analyzing', description: 'Cross-referencing timeline and locations', status: 'pending' },
      { id: '3', phase: 'Validating', description: 'Checking physical constraints and line of sight', status: 'pending' },
      { id: '4', phase: 'Generating', description: 'Building trajectory hypotheses', status: 'pending' },
      { id: '5', phase: 'Ranking', description: 'Scoring paths by evidence support', status: 'pending' },
    ]);
  }, []);

  const handleStartReasoning = useCallback(async () => {
    setStatus('thinking');
    setError(null);
    initializeThinkingSteps();

    try {
      // Trigger reasoning API
      const job = await api.triggerReasoning(caseId, {
        thinking_budget: thinkingBudget,
        max_trajectories: maxTrajectories,
      });

      setJobId(job.id);

      // Poll for progress
      let attempts = 0;
      const maxAttempts = 150; // 5 minutes max
      let currentStep = 0;

      while (attempts < maxAttempts) {
        await new Promise((r) => setTimeout(r, 2000));
        const updatedJob = await api.getJob(job.id);

        // Update thinking steps based on progress
        const progress = updatedJob.progress;
        const newStep = Math.floor(progress / 20);

        if (newStep > currentStep) {
          setThinkingSteps((prev) =>
            prev.map((step, i) => ({
              ...step,
              status: i < newStep ? 'done' : i === newStep ? 'active' : 'pending',
            }))
          );
          currentStep = newStep;
        }

        if (updatedJob.status === 'done') {
          // Mark all steps as done
          setThinkingSteps((prev) => prev.map((s) => ({ ...s, status: 'done' })));

          // Parse results
          const output = updatedJob.output as any;

          if (output?.trajectories) {
            const parsedTrajectories: Trajectory[] = output.trajectories.map((t: any, i: number) => ({
              id: t.id || `traj-${i}`,
              rank: t.rank || i + 1,
              overall_confidence: t.overall_confidence || 0.5,
              segments: t.segments || [],
            }));
            setTrajectories(parsedTrajectories);

            if (parsedTrajectories.length > 0) {
              setSelectedTrajectoryId(parsedTrajectories[0].id);
            }
          }

          if (output?.discrepancies) {
            setDiscrepancies(output.discrepancies);
          }

          setStatus('done');
          setViewMode('reasoning');
          break;
        } else if (updatedJob.status === 'failed') {
          throw new Error(updatedJob.error || 'Reasoning failed');
        }

        attempts++;
      }

      if (attempts >= maxAttempts) {
        throw new Error('Reasoning timed out');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Reasoning failed');
      setStatus('error');
    }
  }, [caseId, thinkingBudget, maxTrajectories, initializeThinkingSteps, setTrajectories, setSelectedTrajectoryId, setViewMode]);

  const getDiscrepancyIcon = (type: Discrepancy['type']) => {
    switch (type) {
      case 'timeline_conflict':
        return Clock;
      case 'line_of_sight':
        return Eye;
      case 'physical_impossible':
        return AlertTriangle;
      case 'testimony_mismatch':
        return Lightbulb;
      default:
        return AlertTriangle;
    }
  };

  const getSeverityColor = (severity: Discrepancy['severity']) => {
    switch (severity) {
      case 'high':
        return '#ef4444';
      case 'medium':
        return '#f59e0b';
      case 'low':
        return '#22c55e';
    }
  };

  return (
    <div className={cn('bg-[#111114] border border-[#2a2a32] rounded-lg overflow-hidden', className)}>
      {/* Header */}
      <div className="flex items-center gap-2 px-4 py-3 border-b border-[#2a2a32]">
        <Brain className="w-4 h-4 text-[#6366f1]" />
        <h3 className="text-sm font-medium text-[#f0f0f2]">Reasoning Engine</h3>
        <button
          onClick={() => setShowSettings(!showSettings)}
          className="ml-auto p-1 hover:bg-[#1f1f24] rounded transition-colors text-[#606068] hover:text-[#a0a0a8]"
        >
          <Settings className="w-4 h-4" />
        </button>
      </div>

      {/* Settings */}
      {showSettings && (
        <div className="p-4 border-b border-[#2a2a32] bg-[#1f1f24]/50 space-y-3">
          <div>
            <label className="text-xs text-[#606068] block mb-1">Thinking Budget</label>
            <input
              type="range"
              min="5000"
              max="50000"
              step="5000"
              value={thinkingBudget}
              onChange={(e) => setThinkingBudget(Number(e.target.value))}
              className="w-full"
            />
            <span className="text-xs text-[#a0a0a8]">{thinkingBudget.toLocaleString()} tokens</span>
          </div>
          <div>
            <label className="text-xs text-[#606068] block mb-1">Max Trajectories</label>
            <input
              type="range"
              min="1"
              max="10"
              value={maxTrajectories}
              onChange={(e) => setMaxTrajectories(Number(e.target.value))}
              className="w-full"
            />
            <span className="text-xs text-[#a0a0a8]">{maxTrajectories}</span>
          </div>
        </div>
      )}

      {/* Idle State */}
      {status === 'idle' && (
        <div className="p-6 text-center">
          <div className="w-16 h-16 rounded-2xl bg-[#6366f1]/20 flex items-center justify-center mx-auto mb-4">
            <Sparkles className="w-8 h-8 text-[#6366f1]" />
          </div>
          <h4 className="text-sm font-medium text-[#f0f0f2] mb-2">Ready to Analyze</h4>
          <p className="text-xs text-[#606068] mb-4">
            Run the reasoning engine to generate trajectory hypotheses and identify discrepancies.
          </p>
          <Button variant="primary" onClick={handleStartReasoning}>
            <Play className="w-4 h-4" />
            Start Reasoning
          </Button>
        </div>
      )}

      {/* Thinking State */}
      {status === 'thinking' && (
        <div className="p-4 space-y-3">
          <div className="flex items-center gap-3 mb-4">
            <Loader2 className="w-5 h-5 text-[#6366f1] animate-spin" />
            <span className="text-sm text-[#f0f0f2]">Reasoning in progress...</span>
          </div>

          {thinkingSteps.map((step) => (
            <div
              key={step.id}
              className={cn(
                'flex items-center gap-3 p-3 rounded-lg transition-all',
                step.status === 'active' && 'bg-[#6366f1]/10 border border-[#6366f1]/30',
                step.status === 'done' && 'bg-[#22c55e]/10',
                step.status === 'pending' && 'opacity-50'
              )}
            >
              {step.status === 'done' ? (
                <CheckCircle className="w-5 h-5 text-[#22c55e]" />
              ) : step.status === 'active' ? (
                <Loader2 className="w-5 h-5 text-[#6366f1] animate-spin" />
              ) : (
                <div className="w-5 h-5 rounded-full border-2 border-[#2a2a32]" />
              )}
              <div className="flex-1">
                <p className="text-sm text-[#f0f0f2]">{step.phase}</p>
                <p className="text-xs text-[#606068]">{step.description}</p>
              </div>
            </div>
          ))}
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
      {status === 'done' && (
        <div className="divide-y divide-[#2a2a32]">
          {/* Success Header */}
          <div className="p-4 bg-[#22c55e]/10">
            <div className="flex items-center gap-3">
              <CheckCircle className="w-5 h-5 text-[#22c55e]" />
              <div>
                <p className="text-sm text-[#f0f0f2]">Analysis Complete</p>
                <p className="text-xs text-[#606068]">
                  {trajectories.length} trajectories, {discrepancies.length} discrepancies found
                </p>
              </div>
            </div>
          </div>

          {/* Trajectories */}
          {trajectories.length > 0 && (
            <div className="p-4">
              <h4 className="text-xs font-medium text-[#606068] uppercase tracking-wider mb-3 flex items-center gap-2">
                <Route className="w-4 h-4" />
                Trajectory Hypotheses
              </h4>

              <div className="space-y-2">
                {trajectories.map((trajectory) => (
                  <button
                    key={trajectory.id}
                    onClick={() => setSelectedTrajectoryId(trajectory.id)}
                    className={cn(
                      'w-full flex items-center gap-3 p-3 rounded-lg text-left transition-all',
                      'hover:bg-[#1f1f24]',
                      selectedTrajectoryId === trajectory.id && 'bg-[#1f1f24] ring-1 ring-[#6366f1]'
                    )}
                  >
                    <div className="w-8 h-8 rounded-lg bg-[#6366f1]/20 flex items-center justify-center">
                      <span className="text-sm font-bold text-[#6366f1]">#{trajectory.rank}</span>
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm text-[#f0f0f2]">
                        Hypothesis #{trajectory.rank}
                      </p>
                      <p className="text-xs text-[#606068]">
                        {trajectory.segments.length} segments
                      </p>
                    </div>
                    <span
                      className={cn(
                        'text-xs px-2 py-0.5 rounded',
                        trajectory.overall_confidence >= 0.7
                          ? 'bg-[#22c55e]/20 text-[#22c55e]'
                          : trajectory.overall_confidence >= 0.5
                          ? 'bg-[#f59e0b]/20 text-[#f59e0b]'
                          : 'bg-[#ef4444]/20 text-[#ef4444]'
                      )}
                    >
                      {formatConfidence(trajectory.overall_confidence)}
                    </span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Discrepancies */}
          {discrepancies.length > 0 && (
            <div className="p-4">
              <h4 className="text-xs font-medium text-[#606068] uppercase tracking-wider mb-3 flex items-center gap-2">
                <AlertTriangle className="w-4 h-4 text-[#ef4444]" />
                Discrepancies Found
              </h4>

              <div className="space-y-2">
                {discrepancies.map((d) => {
                  const Icon = getDiscrepancyIcon(d.type);
                  return (
                    <div
                      key={d.id}
                      className={cn(
                        'p-3 rounded-lg border-l-2',
                        d.severity === 'high' && 'bg-[#ef4444]/10 border-[#ef4444]',
                        d.severity === 'medium' && 'bg-[#f59e0b]/10 border-[#f59e0b]',
                        d.severity === 'low' && 'bg-[#22c55e]/10 border-[#22c55e]'
                      )}
                    >
                      <div className="flex items-start gap-2">
                        <Icon
                          className="w-4 h-4 shrink-0 mt-0.5"
                          style={{ color: getSeverityColor(d.severity) }}
                        />
                        <div>
                          <p className="text-sm text-[#f0f0f2]">{d.description}</p>
                          <p className="text-xs text-[#606068] mt-1">
                            {d.sources.join(' vs ')}
                          </p>
                        </div>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Run Again */}
          <div className="p-4">
            <Button
              variant="ghost"
              size="sm"
              className="w-full"
              onClick={() => {
                setStatus('idle');
                setDiscrepancies([]);
              }}
            >
              <Play className="w-4 h-4" />
              Run Again
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
