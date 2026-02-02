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
        <div className="flex-1 overflow-y-auto">
          {/* POV Simulation */}
          <div className="border-b border-[#1e1e24]">
            <POVSimulation caseId={caseId || ''} />
          </div>

          {/* Video Analyzer */}
          <div>
            <VideoAnalyzer caseId={caseId || ''} />
          </div>
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
