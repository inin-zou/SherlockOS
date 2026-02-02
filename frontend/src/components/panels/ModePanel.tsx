'use client';

import { useStore } from '@/lib/store';
import { POVSimulation } from '@/components/simulation/POVSimulation';
import { VideoAnalyzer } from '@/components/simulation/VideoAnalyzer';
import { ReasoningPanel } from '@/components/reasoning/ReasoningPanel';
import { RightPanel } from './RightPanel';

interface ModePanelProps {
  caseId: string | undefined;
}

export function ModePanel({ caseId }: ModePanelProps) {
  const { viewMode } = useStore();

  if (viewMode === 'evidence') {
    return <RightPanel />;
  }

  if (viewMode === 'simulation') {
    return (
      <div className="w-80 border-l border-[#1e1e24] bg-[#111114] flex flex-col overflow-hidden">
        <div className="flex-1 overflow-y-auto p-3 space-y-3">
          {/* POV Simulation */}
          <POVSimulation caseId={caseId || ''} />

          {/* Video Analyzer */}
          <VideoAnalyzer caseId={caseId || ''} />
        </div>
      </div>
    );
  }

  if (viewMode === 'reasoning') {
    return (
      <div className="w-80 border-l border-[#1e1e24] bg-[#111114] flex flex-col overflow-hidden">
        <ReasoningPanel caseId={caseId || ''} />
      </div>
    );
  }

  return <RightPanel />;
}
