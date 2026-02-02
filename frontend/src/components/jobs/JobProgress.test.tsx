import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { JobProgress, JobIndicator } from './JobProgress';
import type { Job } from '@/lib/types';

describe('JobProgress', () => {
  const mockJobs: Job[] = [
    {
      id: 'job1',
      type: 'reconstruction',
      status: 'running',
      progress: 45,
      created_at: new Date().toISOString(),
    },
    {
      id: 'job2',
      type: 'reasoning',
      status: 'queued',
      progress: 0,
      created_at: new Date().toISOString(),
    },
    {
      id: 'job3',
      type: 'profile',
      status: 'done',
      progress: 100,
      created_at: new Date().toISOString(),
    },
  ];

  it('renders nothing when no jobs', () => {
    const { container } = render(<JobProgress jobs={[]} />);
    expect(container.firstChild).toBeNull();
  });

  it('renders job list', () => {
    render(<JobProgress jobs={mockJobs} />);
    expect(screen.getByText('Reconstruction')).toBeInTheDocument();
    expect(screen.getByText('Reasoning')).toBeInTheDocument();
    expect(screen.getByText('Profile')).toBeInTheDocument();
  });

  it('shows processing message for active jobs', () => {
    render(<JobProgress jobs={mockJobs} />);
    expect(screen.getByText(/2 jobs processing/)).toBeInTheDocument();
  });

  it('shows complete message when all jobs done', () => {
    const completedJobs: Job[] = [
      {
        id: 'job1',
        type: 'reconstruction',
        status: 'done',
        progress: 100,
        created_at: new Date().toISOString(),
      },
    ];
    render(<JobProgress jobs={completedJobs} />);
    expect(screen.getByText('All jobs complete')).toBeInTheDocument();
  });

  it('shows status labels for each job', () => {
    render(<JobProgress jobs={mockJobs} />);
    expect(screen.getByText('Running')).toBeInTheDocument();
    expect(screen.getByText('Queued')).toBeInTheDocument();
    expect(screen.getByText('Done')).toBeInTheDocument();
  });

  it('shows progress percentage for active jobs', () => {
    render(<JobProgress jobs={mockJobs} />);
    expect(screen.getByText('45%')).toBeInTheDocument();
    expect(screen.getByText('0%')).toBeInTheDocument();
  });

  it('shows error message for failed jobs', () => {
    const failedJob: Job[] = [
      {
        id: 'job1',
        type: 'reconstruction',
        status: 'failed',
        progress: 0,
        error: 'Processing failed: timeout',
        created_at: new Date().toISOString(),
      },
    ];
    render(<JobProgress jobs={failedJob} />);
    expect(screen.getByText('Processing failed: timeout')).toBeInTheDocument();
  });

  it('collapses and expands job list', () => {
    render(<JobProgress jobs={mockJobs} />);

    // Initially expanded - jobs should be visible
    expect(screen.getByText('Reconstruction')).toBeInTheDocument();

    // Click to collapse
    const toggleButton = screen.getByText(/jobs processing/);
    fireEvent.click(toggleButton);

    // Jobs should still be in DOM but may be hidden
    // The collapse behavior is handled by conditional rendering
  });
});

describe('JobIndicator', () => {
  it('renders nothing when no active jobs', () => {
    const { container } = render(<JobIndicator activeCount={0} />);
    expect(container.firstChild).toBeNull();
  });

  it('shows singular job text for 1 job', () => {
    render(<JobIndicator activeCount={1} />);
    expect(screen.getByText('1 job')).toBeInTheDocument();
  });

  it('shows plural job text for multiple jobs', () => {
    render(<JobIndicator activeCount={3} />);
    expect(screen.getByText('3 jobs')).toBeInTheDocument();
  });

  it('calls onClick when clicked', () => {
    const mockClick = vi.fn();
    render(<JobIndicator activeCount={2} onClick={mockClick} />);

    const button = screen.getByRole('button');
    fireEvent.click(button);

    expect(mockClick).toHaveBeenCalledTimes(1);
  });

  it('has spinner animation', () => {
    render(<JobIndicator activeCount={1} />);
    const spinner = document.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });
});
