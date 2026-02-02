import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ModeSelector } from './ModeSelector';

// Mock the store
const mockSetViewMode = vi.fn();
let mockViewMode = 'evidence';

vi.mock('@/lib/store', () => ({
  useStore: () => ({
    viewMode: mockViewMode,
    setViewMode: mockSetViewMode,
  }),
}));

describe('ModeSelector', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockViewMode = 'evidence';
  });

  it('renders all three mode buttons', () => {
    render(<ModeSelector />);
    expect(screen.getByText('Evidence')).toBeInTheDocument();
    expect(screen.getByText('Simulation')).toBeInTheDocument();
    expect(screen.getByText('Reasoning')).toBeInTheDocument();
  });

  it('shows evidence mode as active by default', () => {
    render(<ModeSelector />);
    const evidenceButton = screen.getByText('Evidence').closest('button');
    expect(evidenceButton).toHaveClass('text-[#3b82f6]');
  });

  it('calls setViewMode when clicking a mode button', () => {
    render(<ModeSelector />);

    const simulationButton = screen.getByText('Simulation');
    fireEvent.click(simulationButton);

    expect(mockSetViewMode).toHaveBeenCalledWith('simulation');
  });

  it('calls setViewMode for reasoning mode', () => {
    render(<ModeSelector />);

    const reasoningButton = screen.getByText('Reasoning');
    fireEvent.click(reasoningButton);

    expect(mockSetViewMode).toHaveBeenCalledWith('reasoning');
  });

  it('shows mode description', () => {
    render(<ModeSelector />);
    expect(screen.getByText('Upload & analyze evidence')).toBeInTheDocument();
  });

  it('renders icons for each mode', () => {
    render(<ModeSelector />);
    // Check that buttons have icons (SVG elements)
    const buttons = screen.getAllByRole('button');
    buttons.forEach(button => {
      expect(button.querySelector('svg')).toBeInTheDocument();
    });
  });
});
