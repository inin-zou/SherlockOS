import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { ReasoningPanel } from './ReasoningPanel';

// Mock the API
vi.mock('@/lib/api', () => ({
  triggerReasoning: vi.fn().mockResolvedValue({ id: 'job-123' }),
  getJob: vi.fn().mockResolvedValue({
    id: 'job-123',
    status: 'done',
    progress: 100,
    output: {
      trajectories: [
        {
          id: 'traj-1',
          rank: 1,
          overall_confidence: 0.85,
          segments: [],
        },
      ],
      discrepancies: [
        {
          id: 'd1',
          type: 'timeline_conflict',
          severity: 'high',
          description: 'Test discrepancy',
          sources: ['Witness A'],
          evidence: ['CCTV'],
        },
      ],
    },
  }),
}));

// Mock the store
const mockSetTrajectories = vi.fn();
const mockSetSelectedTrajectoryId = vi.fn();
const mockSetViewMode = vi.fn();

vi.mock('@/lib/store', () => ({
  useStore: () => ({
    trajectories: [],
    setTrajectories: mockSetTrajectories,
    selectedTrajectoryId: null,
    setSelectedTrajectoryId: mockSetSelectedTrajectoryId,
    setViewMode: mockSetViewMode,
  }),
}));

describe('ReasoningPanel', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders component with header', () => {
    render(<ReasoningPanel caseId="case-123" />);
    expect(screen.getByText('Reasoning Engine')).toBeInTheDocument();
  });

  it('shows idle state with start button', () => {
    render(<ReasoningPanel caseId="case-123" />);
    expect(screen.getByText('Ready to Analyze')).toBeInTheDocument();
    expect(screen.getByText('Start Reasoning')).toBeInTheDocument();
  });

  it('shows description in idle state', () => {
    render(<ReasoningPanel caseId="case-123" />);
    expect(
      screen.getByText(/run the reasoning engine to generate trajectory hypotheses/i)
    ).toBeInTheDocument();
  });

  it('has settings button', () => {
    render(<ReasoningPanel caseId="case-123" />);
    const settingsButton = document.querySelector('button[class*="hover:bg"]');
    expect(settingsButton).toBeInTheDocument();
  });

  it('toggles settings panel', () => {
    render(<ReasoningPanel caseId="case-123" />);

    // Find and click settings button (the one with Settings icon)
    const buttons = screen.getAllByRole('button');
    const settingsButton = buttons.find(btn =>
      btn.querySelector('svg')?.classList.contains('lucide-settings')
    );

    if (settingsButton) {
      fireEvent.click(settingsButton);
      expect(screen.getByText('Thinking Budget')).toBeInTheDocument();
      expect(screen.getByText('Max Trajectories')).toBeInTheDocument();
    }
  });

  it('clicking start reasoning triggers API call', async () => {
    const { triggerReasoning } = await import('@/lib/api');

    render(<ReasoningPanel caseId="case-123" />);

    const startButton = screen.getByText('Start Reasoning');
    fireEvent.click(startButton);

    await waitFor(() => {
      expect(triggerReasoning).toHaveBeenCalledWith('case-123', expect.any(Object));
    });
  });

  it('shows thinking steps during reasoning', async () => {
    render(<ReasoningPanel caseId="case-123" />);

    const startButton = screen.getByText('Start Reasoning');
    fireEvent.click(startButton);

    await waitFor(() => {
      expect(screen.getByText('Reasoning in progress...')).toBeInTheDocument();
    });
  });

  it('displays thinking phases', async () => {
    render(<ReasoningPanel caseId="case-123" />);

    const startButton = screen.getByText('Start Reasoning');
    fireEvent.click(startButton);

    await waitFor(() => {
      expect(screen.getByText('Gathering')).toBeInTheDocument();
      expect(screen.getByText('Analyzing')).toBeInTheDocument();
      expect(screen.getByText('Validating')).toBeInTheDocument();
      expect(screen.getByText('Generating')).toBeInTheDocument();
      expect(screen.getByText('Ranking')).toBeInTheDocument();
    });
  });
});
