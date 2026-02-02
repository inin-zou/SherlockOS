import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { POVSimulation } from './POVSimulation';

// Mock the API
vi.mock('@/lib/api', () => ({
  createJob: vi.fn().mockResolvedValue({ id: 'job-123' }),
  getJob: vi.fn().mockResolvedValue({
    id: 'job-123',
    status: 'done',
    output: {
      trajectories: [
        {
          id: 'traj-1',
          segments: [
            {
              from_position: [0, 0, 0],
              to_position: [5, 0, 3],
              confidence: 0.8,
              explanation: 'Entry through window',
            },
          ],
        },
      ],
    },
  }),
}));

// Mock the store
vi.mock('@/lib/store', () => ({
  useStore: () => ({
    setTrajectories: vi.fn(),
    setSelectedTrajectoryId: vi.fn(),
  }),
}));

describe('POVSimulation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders component with header', () => {
    render(<POVSimulation caseId="case-123" />);
    expect(screen.getByText('POV Simulation')).toBeInTheDocument();
    expect(screen.getByText('Text to Motion')).toBeInTheDocument();
  });

  it('renders text input area', () => {
    render(<POVSimulation caseId="case-123" />);
    expect(screen.getByPlaceholderText(/describe the suspect's movement/i)).toBeInTheDocument();
  });

  it('shows generate button', () => {
    render(<POVSimulation caseId="case-123" />);
    expect(screen.getByText('Generate')).toBeInTheDocument();
  });

  it('disables generate button when prompt is empty', () => {
    render(<POVSimulation caseId="case-123" />);
    const button = screen.getByText('Generate').closest('button');
    expect(button).toBeDisabled();
  });

  it('enables generate button when prompt has text', () => {
    render(<POVSimulation caseId="case-123" />);

    const textarea = screen.getByPlaceholderText(/describe the suspect's movement/i);
    fireEvent.change(textarea, { target: { value: 'Suspect entered through window' } });

    const button = screen.getByText('Generate').closest('button');
    expect(button).not.toBeDisabled();
  });

  it('shows empty state message', () => {
    render(<POVSimulation caseId="case-123" />);
    expect(screen.getByText(/describe suspect movement to generate motion paths/i)).toBeInTheDocument();
  });

  it('updates textarea value on change', () => {
    render(<POVSimulation caseId="case-123" />);

    const textarea = screen.getByPlaceholderText(/describe the suspect's movement/i) as HTMLTextAreaElement;
    fireEvent.change(textarea, { target: { value: 'Test movement description' } });

    expect(textarea.value).toBe('Test movement description');
  });
});
