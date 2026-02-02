import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { RightPanel } from './RightPanel';

// Mock the store
vi.mock('@/lib/store', () => ({
  useStore: () => ({
    suspectProfile: null,
    trajectories: [],
    sceneGraph: { evidence: [] },
  }),
}));

describe('RightPanel', () => {
  it('renders tab buttons', () => {
    render(<RightPanel />);
    expect(screen.getByText('Suspect')).toBeInTheDocument();
    expect(screen.getByText('Evidence')).toBeInTheDocument();
    expect(screen.getByText('Reasoning')).toBeInTheDocument();
  });

  it('shows Suspect tab by default', () => {
    render(<RightPanel />);
    expect(screen.getByText('Attributes')).toBeInTheDocument();
  });

  it('switches to Evidence tab when clicked', () => {
    render(<RightPanel />);

    fireEvent.click(screen.getByText('Evidence'));
    expect(screen.getByText(/Evidence Cards/)).toBeInTheDocument();
  });

  it('switches to Reasoning tab when clicked', () => {
    render(<RightPanel />);

    fireEvent.click(screen.getByText('Reasoning'));
    expect(screen.getByText(/Discrepancies/)).toBeInTheDocument();
    expect(screen.getByText(/Trajectory Hypotheses/)).toBeInTheDocument();
  });

  describe('Suspect Tab', () => {
    it('shows demo attributes when no profile', () => {
      render(<RightPanel />);

      expect(screen.getByText('Age')).toBeInTheDocument();
      expect(screen.getByText('Height')).toBeInTheDocument();
      expect(screen.getByText('Build')).toBeInTheDocument();
      expect(screen.getByText('Hair')).toBeInTheDocument();
    });

    it('shows distinctive features section', () => {
      render(<RightPanel />);

      expect(screen.getByText('Distinctive Features')).toBeInTheDocument();
    });
  });

  describe('Evidence Tab', () => {
    it('shows demo evidence cards', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Evidence'));

      // Demo evidence items
      expect(screen.getByText('Forced entry marks on window')).toBeInTheDocument();
      expect(screen.getByText('Footprints near vault')).toBeInTheDocument();
      expect(screen.getByText('CCTV timestamp gap')).toBeInTheDocument();
    });

    it('expands evidence card on click', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Evidence'));

      // Click to expand first evidence
      fireEvent.click(screen.getByText('Forced entry marks on window'));

      // Should show description
      expect(screen.getByText(/Tool marks consistent with pry bar/)).toBeInTheDocument();
    });

    it('shows confidence levels', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Evidence'));

      expect(screen.getByText(/85% confidence/)).toBeInTheDocument();
      expect(screen.getByText(/72% confidence/)).toBeInTheDocument();
      expect(screen.getByText(/95% confidence/)).toBeInTheDocument();
    });
  });

  describe('Reasoning Tab', () => {
    it('shows discrepancies section', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Reasoning'));

      expect(screen.getByText(/Discrepancies \(2\)/)).toBeInTheDocument();
    });

    it('shows timeline conflict discrepancy', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Reasoning'));

      expect(screen.getByText(/Witness A claims suspect at gate at 22:10/)).toBeInTheDocument();
    });

    it('shows line of sight discrepancy', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Reasoning'));

      expect(screen.getByText(/Witness B claimed to see suspect from lobby/)).toBeInTheDocument();
    });

    it('shows Tier 3 source labels', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Reasoning'));

      // Multiple discrepancies have "Tier 3:" labels
      const tier3Labels = screen.getAllByText('Tier 3:');
      expect(tier3Labels.length).toBeGreaterThan(0);
      expect(screen.getByText('Witness A statement')).toBeInTheDocument();
    });

    it('shows trajectory hypotheses', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Reasoning'));

      expect(screen.getByText(/Trajectory Hypotheses \(1\)/)).toBeInTheDocument();
      expect(screen.getByText('Hypothesis #1')).toBeInTheDocument();
    });

    it('shows trajectory segments', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Reasoning'));

      expect(screen.getByText('Entry through north window')).toBeInTheDocument();
      expect(screen.getByText('Movement to vault area')).toBeInTheDocument();
    });
  });
});
