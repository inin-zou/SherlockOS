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
    'p-2 rounded-lg transition-all',
    'text-[#606068] hover:text-[#f0f0f2] hover:bg-[#1f1f24]'
  );

  return (
    <div className="flex items-center gap-1">
      <button className={buttonClass} onClick={onStepBack} title="Step Back">
        <ChevronLeft className="w-4 h-4" />
      </button>
      <button className={buttonClass} onClick={onSkipToStart} title="Skip to Start">
        <SkipBack className="w-4 h-4" />
      </button>
      <button className={buttonClass} onClick={onStepBack} title="Previous Frame">
        <ChevronLeft className="w-3 h-3" />
      </button>
      <button
        className={cn(buttonClass, 'bg-[#1f1f24]')}
        onClick={onPlayPause}
        title={isPlaying ? 'Pause' : 'Play'}
      >
        {isPlaying ? (
          <Pause className="w-4 h-4" />
        ) : (
          <Play className="w-4 h-4 ml-0.5" />
        )}
      </button>
      <button className={buttonClass} onClick={onStop} title="Stop">
        <Square className="w-3 h-3" />
      </button>
      <button className={buttonClass} onClick={onStepForward} title="Next Frame">
        <ChevronRight className="w-3 h-3" />
      </button>
      <button className={buttonClass} onClick={onSkipToEnd} title="Skip to End">
        <SkipForward className="w-4 h-4" />
      </button>
      <button className={buttonClass} onClick={onStepForward} title="Step Forward">
        <ChevronRight className="w-4 h-4" />
      </button>
    </div>
  );
}

function TrackRow({ track, currentTime, duration }: { track: TimelineTrack; currentTime: number; duration: number }) {
  return (
    <div className="flex items-center h-8 group">
      {/* Track label */}
      <div className="w-44 shrink-0 px-3 text-sm text-[#a0a0a8] truncate">
        {track.name}
      </div>

      {/* Track timeline */}
      <div className="flex-1 h-full relative bg-[#18181c] rounded-sm mx-2">
        {/* Events */}
        {track.events.map((event) => {
          const left = (event.start / duration) * 100;
          const width = ((event.end - event.start) / duration) * 100;

          return (
            <div
              key={event.id}
              className="absolute top-1 bottom-1 rounded-sm cursor-pointer hover:brightness-110 transition-all"
              style={{
                left: `${left}%`,
                width: `${width}%`,
                backgroundColor: track.color,
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
    const x = e.clientX - rect.left - 176; // Subtract label width
    const trackWidth = rect.width - 176;
    const time = Math.max(0, Math.min(duration, (x / trackWidth) * duration));
    setCurrentTime(time);
  };

  const playheadPosition = (currentTime / duration) * 100;

  return (
    <div
      className="bg-[#111114] border-t border-[#1e1e24] flex flex-col"
      style={{ height: timelineHeight }}
    >
      {/* Playhead indicator and ruler */}
      <div className="h-8 flex items-center border-b border-[#1e1e24]">
        <div className="w-44 shrink-0 px-3">
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
        <div className="flex-1 h-full relative mx-2">
          {/* Time ruler */}
          <div className="absolute inset-0 flex items-end">
            {Array.from({ length: 11 }).map((_, i) => (
              <div
                key={i}
                className="absolute bottom-0 text-[10px] text-[#606068]"
                style={{ left: `${i * 10}%` }}
              >
                {i * 10}
              </div>
            ))}
          </div>
        </div>
        <button className="px-3 text-[#606068] hover:text-[#a0a0a8]">
          <Maximize2 className="w-4 h-4" />
        </button>
      </div>

      {/* Tracks */}
      <div
        ref={timelineRef}
        className="flex-1 overflow-y-auto relative"
        onClick={handleTimelineClick}
      >
        {/* Playhead */}
        <div
          className="playhead"
          style={{ left: `calc(176px + ${playheadPosition}% * (100% - 176px - 16px) / 100)` }}
        />

        <div className="py-2 space-y-1">
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
