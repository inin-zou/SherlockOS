'use client';

import { Scan, Play, Brain } from 'lucide-react';
import { useStore, type ViewMode } from '@/lib/store';
import { cn } from '@/lib/utils';

const modes: { id: ViewMode; label: string; icon: typeof Scan; description: string }[] = [
  {
    id: 'evidence',
    label: 'Evidence',
    icon: Scan,
    description: 'Upload & analyze evidence',
  },
  {
    id: 'simulation',
    label: 'Simulation',
    icon: Play,
    description: 'Generate motion paths',
  },
  {
    id: 'reasoning',
    label: 'Reasoning',
    icon: Brain,
    description: 'AI-powered analysis',
  },
];

export function ModeSelector() {
  const { viewMode, setViewMode } = useStore();

  return (
    <div className="flex items-center gap-1 px-4 py-2 border-b border-[#1e1e24] bg-[#111114]">
      {modes.map((mode) => {
        const Icon = mode.icon;
        const isActive = viewMode === mode.id;

        return (
          <button
            key={mode.id}
            onClick={() => setViewMode(mode.id)}
            className={cn(
              'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all',
              isActive
                ? 'bg-[#3b82f6]/10 text-[#3b82f6] border border-[#3b82f6]/30'
                : 'text-[#606068] hover:text-[#a0a0a8] hover:bg-[#1f1f24]'
            )}
            title={mode.description}
          >
            <Icon className="w-4 h-4" />
            {mode.label}
          </button>
        );
      })}

      {/* Mode description */}
      <div className="ml-auto text-xs text-[#606068]">
        {modes.find((m) => m.id === viewMode)?.description}
      </div>
    </div>
  );
}
