import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import {
  LoadingSpinner,
  LoadingState,
  Skeleton,
  EvidenceListSkeleton,
  TimelineSkeleton,
  ProfileSkeleton,
  EmptyState,
  EvidenceEmptyState,
  TimelineEmptyState,
  TrajectoriesEmptyState,
  ProfileEmptyState,
  ReasoningEmptyState,
} from './LoadingStates';
import { FileQuestion } from 'lucide-react';

describe('LoadingSpinner', () => {
  it('renders with default size', () => {
    render(<LoadingSpinner />);
    const spinner = document.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
    expect(spinner).toHaveClass('w-6', 'h-6');
  });

  it('renders with small size', () => {
    render(<LoadingSpinner size="sm" />);
    const spinner = document.querySelector('.animate-spin');
    expect(spinner).toHaveClass('w-4', 'h-4');
  });

  it('renders with large size', () => {
    render(<LoadingSpinner size="lg" />);
    const spinner = document.querySelector('.animate-spin');
    expect(spinner).toHaveClass('w-8', 'h-8');
  });

  it('applies custom className', () => {
    render(<LoadingSpinner className="custom-class" />);
    const spinner = document.querySelector('.animate-spin');
    expect(spinner).toHaveClass('custom-class');
  });
});

describe('LoadingState', () => {
  it('renders with default message', () => {
    render(<LoadingState />);
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('renders with custom message', () => {
    render(<LoadingState message="Fetching data..." />);
    expect(screen.getByText('Fetching data...')).toBeInTheDocument();
  });

  it('includes a spinner', () => {
    render(<LoadingState />);
    const spinner = document.querySelector('.animate-spin');
    expect(spinner).toBeInTheDocument();
  });
});

describe('Skeleton', () => {
  it('renders with animate-pulse class', () => {
    render(<Skeleton data-testid="skeleton" />);
    const skeleton = document.querySelector('.animate-pulse');
    expect(skeleton).toBeInTheDocument();
  });

  it('applies custom className', () => {
    render(<Skeleton className="h-4 w-32" />);
    const skeleton = document.querySelector('.animate-pulse');
    expect(skeleton).toHaveClass('h-4', 'w-32');
  });
});

describe('Skeleton composites', () => {
  it('renders EvidenceListSkeleton', () => {
    render(<EvidenceListSkeleton />);
    const skeletons = document.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('renders TimelineSkeleton', () => {
    render(<TimelineSkeleton />);
    const skeletons = document.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('renders ProfileSkeleton', () => {
    render(<ProfileSkeleton />);
    const skeletons = document.querySelectorAll('.animate-pulse');
    expect(skeletons.length).toBeGreaterThan(0);
  });
});

describe('EmptyState', () => {
  it('renders with title', () => {
    render(<EmptyState title="No items" />);
    expect(screen.getByText('No items')).toBeInTheDocument();
  });

  it('renders with description', () => {
    render(<EmptyState title="No items" description="Add some items to get started" />);
    expect(screen.getByText('Add some items to get started')).toBeInTheDocument();
  });

  it('renders with action', () => {
    render(
      <EmptyState
        title="No items"
        action={<button>Add Item</button>}
      />
    );
    expect(screen.getByText('Add Item')).toBeInTheDocument();
  });

  it('renders with custom icon', () => {
    render(<EmptyState title="Test" icon={FileQuestion} />);
    // The icon should render (we can't easily test the icon itself, but the component renders)
    expect(screen.getByText('Test')).toBeInTheDocument();
  });
});

describe('Empty state presets', () => {
  it('renders EvidenceEmptyState', () => {
    render(<EvidenceEmptyState />);
    expect(screen.getByText('No evidence yet')).toBeInTheDocument();
  });

  it('renders TimelineEmptyState', () => {
    render(<TimelineEmptyState />);
    expect(screen.getByText('No timeline events')).toBeInTheDocument();
  });

  it('renders TrajectoriesEmptyState', () => {
    render(<TrajectoriesEmptyState />);
    expect(screen.getByText('No trajectories')).toBeInTheDocument();
  });

  it('renders ProfileEmptyState', () => {
    render(<ProfileEmptyState />);
    expect(screen.getByText('No suspect profile')).toBeInTheDocument();
  });

  it('renders ReasoningEmptyState', () => {
    render(<ReasoningEmptyState />);
    expect(screen.getByText('No analysis yet')).toBeInTheDocument();
  });
});
