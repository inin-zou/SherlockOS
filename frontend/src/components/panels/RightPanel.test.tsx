import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { RightPanel } from './RightPanel';

// Create a mockable store state
const createMockStore = (overrides = {}) => ({
  suspectProfile: null,
  trajectories: [],
  sceneGraph: { evidence: [] },
  ...overrides,
});

// Default mock for empty state tests
vi.mock('@/lib/store', () => ({
  useStore: vi.fn(() => createMockStore()),
}));

import { useStore } from '@/lib/store';

describe('RightPanel', () => {
  beforeEach(() => {
    vi.mocked(useStore).mockReturnValue(createMockStore());
  });

  it('renders tab buttons', () => {
    render(<RightPanel />);
    expect(screen.getByText('Suspect')).toBeInTheDocument();
    expect(screen.getByText('Evidence')).toBeInTheDocument();
    expect(screen.getByText('Reasoning')).toBeInTheDocument();
  });

  it('shows Suspect tab by default', () => {
    render(<RightPanel />);
    // Should show empty state when no profile
    expect(screen.getByText('No suspect profile')).toBeInTheDocument();
  });

  it('switches to Evidence tab when clicked', () => {
    render(<RightPanel />);

    fireEvent.click(screen.getByText('Evidence'));
    // Should show empty state when no evidence
    expect(screen.getByText('No evidence yet')).toBeInTheDocument();
  });

  it('switches to Reasoning tab when clicked', () => {
    render(<RightPanel />);

    fireEvent.click(screen.getByText('Reasoning'));
    // Should show empty state when no trajectories
    expect(screen.getByText('No trajectories')).toBeInTheDocument();
  });

  describe('Suspect Tab - Empty State', () => {
    it('shows empty state when no profile', () => {
      render(<RightPanel />);

      expect(screen.getByText('No suspect profile')).toBeInTheDocument();
      expect(screen.getByText('Add witness statements to build a suspect profile')).toBeInTheDocument();
    });
  });

  describe('Suspect Tab - With Data', () => {
    const mockProfile = {
      id: 'profile-1',
      attributes: {
        age_range: { min: 25, max: 35, confidence: 0.7 },
        height_range_cm: { min: 170, max: 180, confidence: 0.8 },
        build: { value: 'average', confidence: 0.6 },
        hair: { style: 'short', color: 'dark', confidence: 0.75 },
        distinctive_features: [
          { description: 'Walks with slight limp', confidence: 0.65 },
        ],
      },
    };

    beforeEach(() => {
      vi.mocked(useStore).mockReturnValue(createMockStore({
        suspectProfile: mockProfile,
      }));
    });

    it('shows attributes when profile exists', () => {
      render(<RightPanel />);

      expect(screen.getByText('Age')).toBeInTheDocument();
      expect(screen.getByText('Height')).toBeInTheDocument();
      expect(screen.getByText('Build')).toBeInTheDocument();
      expect(screen.getByText('Hair')).toBeInTheDocument();
    });

    it('shows distinctive features section', () => {
      render(<RightPanel />);

      expect(screen.getByText('Distinctive Features')).toBeInTheDocument();
      expect(screen.getByText('Walks with slight limp')).toBeInTheDocument();
    });
  });

  describe('Evidence Tab - Empty State', () => {
    it('shows empty state when no evidence', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Evidence'));

      expect(screen.getByText('No evidence yet')).toBeInTheDocument();
    });
  });

  describe('Evidence Tab - With Data', () => {
    const mockEvidence = [
      {
        id: 'ev1',
        object_ids: ['obj1'],
        title: 'Forced entry marks on window',
        description: 'Tool marks consistent with pry bar, 15cm scratch pattern',
        confidence: 0.85,
        sources: [{ type: 'upload', commit_id: 'c1' }],
        created_at: new Date().toISOString(),
      },
      {
        id: 'ev2',
        object_ids: ['obj2'],
        title: 'Footprints near vault',
        description: 'Size 10 boot prints, estimated 180cm tall suspect',
        confidence: 0.72,
        sources: [{ type: 'upload', commit_id: 'c1' }],
        created_at: new Date().toISOString(),
      },
    ];

    beforeEach(() => {
      vi.mocked(useStore).mockReturnValue(createMockStore({
        sceneGraph: { evidence: mockEvidence },
      }));
    });

    it('shows evidence cards', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Evidence'));

      expect(screen.getByText('Forced entry marks on window')).toBeInTheDocument();
      expect(screen.getByText('Footprints near vault')).toBeInTheDocument();
    });

    it('expands evidence card on click', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Evidence'));

      fireEvent.click(screen.getByText('Forced entry marks on window'));
      expect(screen.getByText(/Tool marks consistent with pry bar/)).toBeInTheDocument();
    });

    it('shows confidence levels', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Evidence'));

      expect(screen.getByText(/85% confidence/)).toBeInTheDocument();
      expect(screen.getByText(/72% confidence/)).toBeInTheDocument();
    });
  });

  describe('Reasoning Tab - Empty State', () => {
    it('shows empty state when no trajectories', () => {
      render(<RightPanel />);
      fireEvent.click(screen.getByText('Reasoning'));

      expect(screen.getByText('No trajectories')).toBeInTheDocument();
      expect(screen.getByText('Run the reasoning engine to generate movement hypotheses')).toBeInTheDocument();
    });
  });

  describe('Reasoning Tab - With Data', () => {
    const mockTrajectories = [
      {
        id: 'traj1',
        rank: 1,
        overall_confidence: 0.85,
        segments: [
          {
            id: 's1',
            from_position: [0, 0, 0],
            to_position: [5, 0, 3],
            evidence_refs: [],
            confidence: 0.9,
            explanation: 'Entry through north window',
          },
          {
            id: 's2',
            from_position: [5, 0, 3],
            to_position: [8, 0, 5],
            evidence_refs: [],
            confidence: 0.8,
            explanation: 'Movement to vault area',
          },
        ],
      },
    ];

    beforeEach(() => {
      vi.mocked(useStore).mockReturnValue(createMockStore({
        trajectories: mockTrajectories,
      }));
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
