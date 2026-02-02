import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { CommitTimeline, CommitBadge } from './CommitTimeline';
import type { Commit } from '@/lib/types';

describe('CommitTimeline', () => {
  const mockCommits: Commit[] = [
    {
      id: 'c1',
      case_id: 'case1',
      type: 'upload_scan',
      summary: 'Uploaded 3 scan images',
      payload: { images: 3 },
      created_at: new Date().toISOString(),
    },
    {
      id: 'c2',
      case_id: 'case1',
      type: 'witness_statement',
      summary: 'Added statement from Guard A',
      payload: { source: 'Guard A' },
      created_at: new Date(Date.now() - 3600000).toISOString(),
    },
    {
      id: 'c3',
      case_id: 'case1',
      type: 'reasoning_result',
      summary: 'Generated 2 trajectories',
      payload: { trajectories: 2 },
      created_at: new Date(Date.now() - 7200000).toISOString(),
    },
  ];

  const mockOnCommitSelect = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders commit history header', () => {
    render(<CommitTimeline commits={mockCommits} />);
    expect(screen.getByText('Commit History')).toBeInTheDocument();
  });

  it('renders all commits', () => {
    render(<CommitTimeline commits={mockCommits} />);
    expect(screen.getByText('Uploaded 3 scan images')).toBeInTheDocument();
    expect(screen.getByText('Added statement from Guard A')).toBeInTheDocument();
    expect(screen.getByText('Generated 2 trajectories')).toBeInTheDocument();
  });

  it('shows commit type labels', () => {
    render(<CommitTimeline commits={mockCommits} />);
    expect(screen.getByText('Scan Upload')).toBeInTheDocument();
    expect(screen.getByText('Witness Statement')).toBeInTheDocument();
    expect(screen.getByText('Reasoning')).toBeInTheDocument();
  });

  it('calls onCommitSelect when commit clicked', () => {
    render(
      <CommitTimeline commits={mockCommits} onCommitSelect={mockOnCommitSelect} />
    );

    const firstCommit = screen.getByText('Uploaded 3 scan images');
    fireEvent.click(firstCommit);

    expect(mockOnCommitSelect).toHaveBeenCalledWith(mockCommits[0]);
  });

  it('highlights selected commit', () => {
    render(
      <CommitTimeline
        commits={mockCommits}
        selectedCommitId="c1"
        onCommitSelect={mockOnCommitSelect}
      />
    );

    // Check that the selected commit has the ring highlight
    const selectedButton = screen.getByText('Uploaded 3 scan images').closest('button');
    expect(selectedButton).toHaveClass('ring-1');
  });

  it('expands commit to show payload on click', () => {
    render(<CommitTimeline commits={mockCommits} />);

    // Click to expand
    const commit = screen.getByText('Uploaded 3 scan images');
    fireEvent.click(commit);

    // Should show JSON payload
    expect(screen.getByText(/"images": 3/)).toBeInTheDocument();
  });

  it('renders demo commits when empty array provided', () => {
    render(<CommitTimeline commits={[]} />);
    // Should show demo data
    expect(screen.getByText('Commit History')).toBeInTheDocument();
  });
});

describe('CommitBadge', () => {
  it('renders commit type label', () => {
    const commit: Commit = {
      id: 'c1',
      case_id: 'case1',
      type: 'upload_scan',
      summary: 'Test',
      payload: {},
      created_at: new Date().toISOString(),
    };

    render(<CommitBadge commit={commit} />);
    expect(screen.getByText('Scan Upload')).toBeInTheDocument();
  });

  it('renders different commit types', () => {
    const types = [
      { type: 'witness_statement' as const, label: 'Witness Statement' },
      { type: 'manual_edit' as const, label: 'Manual Edit' },
      { type: 'reconstruction_update' as const, label: 'Reconstruction' },
      { type: 'profile_update' as const, label: 'Profile Update' },
      { type: 'reasoning_result' as const, label: 'Reasoning' },
    ];

    types.forEach(({ type, label }) => {
      const commit: Commit = {
        id: 'c1',
        case_id: 'case1',
        type,
        summary: 'Test',
        payload: {},
        created_at: new Date().toISOString(),
      };

      const { unmount } = render(<CommitBadge commit={commit} />);
      expect(screen.getByText(label)).toBeInTheDocument();
      unmount();
    });
  });
});
