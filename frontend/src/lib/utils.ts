import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatDate(date: string | Date): string {
  const d = new Date(date);
  return new Intl.DateTimeFormat('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }).format(d);
}

export function formatDuration(ms: number): string {
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m`;
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  }
  return `${seconds}s`;
}

export function formatConfidence(confidence: number): string {
  return `${Math.round(confidence * 100)}%`;
}

export function getFileIcon(type: string): string {
  const iconMap: Record<string, string> = {
    pdf: 'FileText',
    image: 'Image',
    video: 'Video',
    audio: 'Music',
    json: 'FileJson',
    text: 'FileText',
    '3d': 'Box',
  };
  return iconMap[type] || 'File';
}

export function getCommitIcon(type: string): string {
  const iconMap: Record<string, string> = {
    upload_scan: 'Upload',
    witness_statement: 'MessageSquare',
    manual_edit: 'Edit',
    reconstruction_update: 'Box',
    profile_update: 'User',
    reasoning_result: 'Brain',
    export_report: 'FileOutput',
    replay_generated: 'Video',
  };
  return iconMap[type] || 'Circle';
}

export function getJobStatusColor(status: string): string {
  const colorMap: Record<string, string> = {
    queued: 'text-gray-400',
    running: 'text-blue-400',
    done: 'text-green-400',
    failed: 'text-red-400',
    canceled: 'text-gray-500',
  };
  return colorMap[status] || 'text-gray-400';
}

export function getConfidenceColor(confidence: number): string {
  if (confidence >= 0.8) return 'text-green-400';
  if (confidence >= 0.5) return 'text-yellow-400';
  return 'text-red-400';
}

export function generateId(): string {
  return Math.random().toString(36).substring(2, 11);
}

export function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * t;
}

export function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

export function mapRange(
  value: number,
  inMin: number,
  inMax: number,
  outMin: number,
  outMax: number
): number {
  return ((value - inMin) * (outMax - outMin)) / (inMax - inMin) + outMin;
}
