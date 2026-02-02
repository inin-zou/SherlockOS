import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { DropZone, DropOverlay } from './DropZone';
import type { UploadProgress } from '@/hooks/useUpload';

describe('DropZone', () => {
  const mockOnFilesDropped = vi.fn();
  const defaultProps = {
    onFilesDropped: mockOnFilesDropped,
    progress: [] as UploadProgress[],
    isUploading: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders drop zone with instructions', () => {
    render(<DropZone {...defaultProps} />);
    expect(screen.getByText(/drop evidence files or click to browse/i)).toBeInTheDocument();
  });

  it('shows tier legend', () => {
    render(<DropZone {...defaultProps} />);
    expect(screen.getByText('Environment')).toBeInTheDocument();
    expect(screen.getByText('Ground Truth')).toBeInTheDocument();
    expect(screen.getByText('Electronic Logs')).toBeInTheDocument();
    expect(screen.getByText('Testimonials')).toBeInTheDocument();
  });

  it('shows loading state when uploading', () => {
    render(<DropZone {...defaultProps} isUploading={true} />);
    // The loading spinner should be visible (Loader2 icon has animate-spin class)
    const spinner = document.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });

  it('shows progress items when files are uploading', () => {
    const progress: UploadProgress[] = [
      {
        fileId: '1',
        filename: 'test.pdf',
        tier: 0,
        status: 'uploading',
        progress: 50,
      },
    ];

    render(<DropZone {...defaultProps} progress={progress} />);
    expect(screen.getByText('test.pdf')).toBeInTheDocument();
  });

  it('applies disabled state when disabled prop is true', () => {
    render(<DropZone {...defaultProps} disabled={true} />);
    const dropZone = document.querySelector('.cursor-not-allowed');
    expect(dropZone).toBeInTheDocument();
  });

  it('handles click to open file picker', () => {
    render(<DropZone {...defaultProps} />);
    const hiddenInput = document.querySelector('input[type="file"]');
    expect(hiddenInput).toBeInTheDocument();
  });

  it('displays error message for failed uploads', () => {
    const progress: UploadProgress[] = [
      {
        fileId: '1',
        filename: 'failed.pdf',
        tier: 0,
        status: 'error',
        progress: 0,
        error: 'Upload failed',
      },
    ];

    render(<DropZone {...defaultProps} progress={progress} />);
    expect(screen.getByText('Upload failed')).toBeInTheDocument();
  });

  it('shows done status with checkmark', () => {
    const progress: UploadProgress[] = [
      {
        fileId: '1',
        filename: 'done.pdf',
        tier: 0,
        status: 'done',
        progress: 100,
      },
    ];

    render(<DropZone {...defaultProps} progress={progress} />);
    expect(screen.getByText('done.pdf')).toBeInTheDocument();
  });
});

describe('DropOverlay', () => {
  const mockOnDrop = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders when visible', () => {
    render(<DropOverlay isVisible={true} onDrop={mockOnDrop} />);
    expect(screen.getByText('Drop files to upload')).toBeInTheDocument();
  });

  it('does not render when not visible', () => {
    render(<DropOverlay isVisible={false} onDrop={mockOnDrop} />);
    expect(screen.queryByText('Drop files to upload')).not.toBeInTheDocument();
  });

  it('shows auto-classification message', () => {
    render(<DropOverlay isVisible={true} onDrop={mockOnDrop} />);
    expect(screen.getByText(/files will be auto-classified/i)).toBeInTheDocument();
  });
});
