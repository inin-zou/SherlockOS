import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { VideoAnalyzer } from './VideoAnalyzer';

// Mock the API
vi.mock('@/lib/api', () => ({
  getUploadIntent: vi.fn().mockResolvedValue({
    upload_batch_id: 'batch-1',
    intents: [
      {
        filename: 'test.mp4',
        storage_key: 'key-1',
        presigned_url: 'https://example.com/upload',
        expires_at: new Date().toISOString(),
      },
    ],
  }),
  uploadFile: vi.fn().mockResolvedValue(undefined),
  createJob: vi.fn().mockResolvedValue({ id: 'job-123' }),
  getJob: vi.fn().mockResolvedValue({
    id: 'job-123',
    status: 'done',
    progress: 100,
    output: {
      duration: 30000,
      fps: 30,
      detected_objects: [
        {
          timestamp: 5000,
          duration: 2000,
          type: 'person',
          confidence: 0.85,
          description: 'Person walking',
        },
      ],
    },
  }),
}));

// Mock the store
vi.mock('@/lib/store', () => ({
  useStore: () => ({
    setTimelineTracks: vi.fn(),
    addCommit: vi.fn(),
  }),
}));

describe('VideoAnalyzer', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders component with header', () => {
    render(<VideoAnalyzer caseId="case-123" />);
    expect(screen.getByText('Video Analyzer')).toBeInTheDocument();
    expect(screen.getByText('CCTV to Motion')).toBeInTheDocument();
  });

  it('renders upload zone in idle state', () => {
    render(<VideoAnalyzer caseId="case-123" />);
    expect(screen.getByText('Upload CCTV footage')).toBeInTheDocument();
    expect(screen.getByText('MP4, MOV, AVI supported')).toBeInTheDocument();
  });

  it('has hidden file input', () => {
    render(<VideoAnalyzer caseId="case-123" />);
    const input = document.querySelector('input[type="file"]');
    expect(input).toBeInTheDocument();
    expect(input).toHaveClass('hidden');
  });

  it('accepts video file types', () => {
    render(<VideoAnalyzer caseId="case-123" />);
    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    expect(input.accept).toBe('video/*');
  });

  it('clicking upload zone triggers file input', () => {
    render(<VideoAnalyzer caseId="case-123" />);

    const input = document.querySelector('input[type="file"]') as HTMLInputElement;
    const clickSpy = vi.spyOn(input, 'click');

    const uploadButton = screen.getByText('Upload CCTV footage').closest('button');
    if (uploadButton) {
      fireEvent.click(uploadButton);
    }

    expect(clickSpy).toHaveBeenCalled();
  });
});
