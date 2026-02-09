'use client';

import { useState, useRef } from 'react';
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  ChevronLeft,
  ChevronRight,
  Square,
  Maximize2,
} from 'lucide-react';
import { useStore } from '@/lib/store';
import { cn } from '@/lib/utils';
import { TimelineEmptyState } from '@/components/ui/LoadingStates';

interface TimelineTrack {
  id: string;
  name: string;
  color: string;
  events: Array<{
    id: string;
    start: number;
    end: number;
    label?: string;
  }>;
}

// Demo tracks
const demoTracks: TimelineTrack[] = [
  {
    id: 'primary',
    name: 'Primary Residence',
    color: '#6366f1',
    events: [
      { id: 'e1', start: 0, end: 100, label: 'Initial Entry' },
    ],
  },
  {
    id: 'gate',
    name: 'External Gate',
    color: '#8b5cf6',
    events: [
      { id: 'e2', start: 10, end: 20 },
      { id: 'e3', start: 85, end: 95 },
    ],
  },
  {
    id: 'downtown',
    name: 'Downtown Hub',
    color: '#a855f7',
    events: [
      { id: 'e4', start: 25, end: 40 },
    ],
  },
  {
    id: 'suspect',
    name: 'Suspect B (Ex-Partner)',
    color: '#c084fc',
    events: [
      { id: 'e5', start: 15, end: 35 },
      { id: 'e6', start: 60, end: 75 },
    ],
  },
  {
    id: 'victim',
    name: 'Nova Welsh (Victim)',
    color: '#e879f9',
    events: [
      { id: 'e7', start: 0, end: 50 },
    ],
  },
];

function PlaybackControls({
  isPlaying,
  onPlayPause,
  onStepBack,
  onStepForward,
  onSkipToStart,
  onSkipToEnd,
  onStop,
}: {
  isPlaying: boolean;
  onPlayPause: () => void;
  onStepBack: () => void;
  onStepForward: () => void;
  onSkipToStart: () => void;
  onSkipToEnd: () => void;
  onStop: () => void;
}) {
  const buttonClass = cn(
    'p-1.5 rounded transition-all',
    'text-[#a0a0a8] hover:text-[#f0f0f2] hover:bg-[#27272a]'
  );

  return (
    <div className="flex items-center gap-2 px-3 py-1.5 bg-[#111114]/80 backdrop-blur-sm border border-[#27272a] rounded-full shadow-lg">
      <button className={buttonClass} onClick={onSkipToStart} title="Start">
        <SkipBack className="w-3 h-3" />
      </button>
      <button className={buttonClass} onClick={onStepBack} title="Step Back">
        <ChevronLeft className="w-3 h-3" />
      </button>
      <button
        className={cn(buttonClass, 'text-[#f0f0f2]')}
        onClick={onPlayPause}
        title={isPlaying ? 'Pause' : 'Play'}
      >
        {isPlaying ? (
          <Pause className="w-3 h-3" />
        ) : (
          <Play className="w-3 h-3 fill-current" />
        )}
      </button>
      <button className={buttonClass} onClick={onStepForward} title="Step Forward">
        <ChevronRight className="w-3 h-3" />
      </button>
      <button className={buttonClass} onClick={onSkipToEnd} title="End">
        <SkipForward className="w-3 h-3" />
      </button>
    </div>
  );
}

function TrackRow({ track, currentTime, duration }: { track: TimelineTrack; currentTime: number; duration: number }) {
  return (
    <div className="flex items-center h-10 group px-4">
      {/* Track label */}
      <div className="w-40 shrink-0 pr-4 text-xs font-medium text-[#a0a0a8] truncate">
        {track.name}
      </div>

      {/* Track timeline */}
      <div className="flex-1 h-6 relative bg-[#1f1f24] rounded-md mx-2 overflow-hidden">
        {/* Events */}
        {track.events.map((event) => {
          const left = (event.start / duration) * 100;
          const width = ((event.end - event.start) / duration) * 100;

          return (
            <div
              key={event.id}
              className="absolute top-1 bottom-1 rounded cursor-pointer hover:brightness-110 transition-all bg-[#fde047]"
              style={{
                left: `${left}%`,
                width: `${width}%`,
              }}
              title={event.label}
            />
          );
        })}
      </div>
    </div>
  );
}

export function Timeline() {
  const { isPlaying, setIsPlaying, currentTime, setCurrentTime, timelineHeight } = useStore();
  const [duration] = useState(100);
  const timelineRef = useRef<HTMLDivElement>(null);

  const tracks = demoTracks;

  const handlePlayPause = () => setIsPlaying(!isPlaying);
  const handleStop = () => {
    setIsPlaying(false);
    setCurrentTime(0);
  };
  const handleStepBack = () => setCurrentTime(Math.max(0, currentTime - 1));
  const handleStepForward = () => setCurrentTime(Math.min(duration, currentTime + 1));
  const handleSkipToStart = () => setCurrentTime(0);
  const handleSkipToEnd = () => setCurrentTime(duration);

  const handleTimelineClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!timelineRef.current) return;
    const rect = timelineRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left - 184; // 160px label + 24px padding/gap roughly
    const trackWidth = rect.width - 200; // Approximate available width
    // Simplified click handling for now - would need precise math based on layout
    const percent = Math.max(0, Math.min(1, (e.clientX - rect.left - 180) / (rect.width - 200)));
    setCurrentTime(percent * duration);
  };

  const playheadPosition = (currentTime / duration) * 100;

  return (
    <div
      className="bg-[#09090B] border-t border-[#1e1e24] flex flex-col relative"
      style={{ height: timelineHeight }}
    >
      {/* Floating Controls */}
      <div className="absolute top-2 left-1/2 -translate-x-1/2 z-20">
        <PlaybackControls
          isPlaying={isPlaying}
          onPlayPause={handlePlayPause}
          onStop={handleStop}
          onStepBack={handleStepBack}
          onStepForward={handleStepForward}
          onSkipToStart={handleSkipToStart}
          onSkipToEnd={handleSkipToEnd}
        />
      </div>

      {/* Tracks Container */}
      <div
        ref={timelineRef}
        className="flex-1 overflow-y-auto relative pt-12 pb-4"
        onClick={handleTimelineClick}
      >
        {/* Playhead Line - spans full height */}
        <div
          className="absolute top-12 bottom-0 z-10 pointer-events-none"
          style={{ 
            left: `calc(160px + 24px + ${playheadPosition}% * (100% - 200px) / 100)`, // Adjust based on layout
            transform: 'translateX(-50%)'
          }}
        >
          {/* Triangle Indicator */}
          <div className="w-0 h-0 border-l-[6px] border-l-transparent border-r-[6px] border-r-transparent border-t-[8px] border-t-[#fde047] absolute -top-1 left-1/2 -translate-x-1/2" />
          {/* Line */}
          <div className="w-px h-full bg-[#fde047]" />
        </div>

        <div className="space-y-1">
          {tracks.map((track) => (
            <TrackRow
              key={track.id}
              track={track}
              currentTime={currentTime}
              duration={duration}
            />
          ))}
        </div>
      </div>
    </div>
  );
}
