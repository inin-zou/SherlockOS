import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ModePanel } from './ModePanel';

// Mock the store
let mockViewMode = 'evidence';

vi.mock('@/lib/store', () => ({
  useStore: () => ({
    viewMode: mockViewMode,
    trajectories: [],
    selectedTrajectoryId: null,
    setTrajectories: vi.fn(),
    setSelectedTrajectoryId: vi.fn(),
    setViewMode: vi.fn(),
    suspectProfile: null,
    sceneGraph: null,
  }),
}));

// Mock child components
vi.mock('./RightPanel', () => ({
  RightPanel: () => <div data-testid="right-panel">RightPanel</div>,
}));

vi.mock('@/components/simulation/POVSimulation', () => ({
  POVSimulation: ({ caseId }: { caseId: string }) => (
    <div data-testid="pov-simulation">POVSimulation: {caseId}</div>
  ),
}));

vi.mock('@/components/simulation/VideoAnalyzer', () => ({
  VideoAnalyzer: ({ caseId }: { caseId: string }) => (
    <div data-testid="video-analyzer">VideoAnalyzer: {caseId}</div>
  ),
}));

vi.mock('@/components/reasoning/ReasoningPanel', () => ({
  ReasoningPanel: ({ caseId }: { caseId: string }) => (
    <div data-testid="reasoning-panel">ReasoningPanel: {caseId}</div>
  ),
}));

describe('ModePanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockViewMode = 'evidence';
  });

  it('renders RightPanel in evidence mode', () => {
    mockViewMode = 'evidence';
    render(<ModePanel caseId="case-123" />);
    expect(screen.getByTestId('right-panel')).toBeInTheDocument();
  });

  it('renders simulation components in simulation mode', () => {
    mockViewMode = 'simulation';
    render(<ModePanel caseId="case-123" />);
    expect(screen.getByTestId('pov-simulation')).toBeInTheDocument();
    expect(screen.getByTestId('video-analyzer')).toBeInTheDocument();
  });

  it('passes caseId to simulation components', () => {
    mockViewMode = 'simulation';
    render(<ModePanel caseId="case-123" />);
    expect(screen.getByText('POVSimulation: case-123')).toBeInTheDocument();
    expect(screen.getByText('VideoAnalyzer: case-123')).toBeInTheDocument();
  });

  it('renders reasoning panel in reasoning mode', () => {
    mockViewMode = 'reasoning';
    render(<ModePanel caseId="case-123" />);
    expect(screen.getByTestId('reasoning-panel')).toBeInTheDocument();
  });

  it('passes caseId to reasoning panel', () => {
    mockViewMode = 'reasoning';
    render(<ModePanel caseId="case-123" />);
    expect(screen.getByText('ReasoningPanel: case-123')).toBeInTheDocument();
  });

  it('handles undefined caseId gracefully', () => {
    mockViewMode = 'simulation';
    render(<ModePanel caseId={undefined} />);
    expect(screen.getByText('POVSimulation:')).toBeInTheDocument();
  });

  it('defaults to RightPanel for unknown modes', () => {
    mockViewMode = 'unknown' as any;
    render(<ModePanel caseId="case-123" />);
    expect(screen.getByTestId('right-panel')).toBeInTheDocument();
  });
});
