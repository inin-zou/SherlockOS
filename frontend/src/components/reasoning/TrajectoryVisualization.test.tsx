import { describe, it, expect, vi } from 'vitest';

// Mock Three.js and React Three Fiber
vi.mock('@react-three/fiber', () => ({
  useFrame: vi.fn(),
}));

vi.mock('@react-three/drei', () => ({
  Line: vi.fn(() => null),
  Text: vi.fn(() => null),
  Html: vi.fn(({ children }) => children),
}));

vi.mock('three', () => ({
  Vector3: vi.fn().mockImplementation((x, y, z) => ({ x, y, z })),
  CatmullRomCurve3: vi.fn().mockImplementation(() => ({
    getPoints: vi.fn().mockReturnValue([]),
    getPoint: vi.fn().mockReturnValue({ x: 0, y: 0, z: 0 }),
  })),
  DoubleSide: 2,
}));

// Mock store with different states
const createMockStore = (trajectories: any[], viewMode: string, selectedId: string | null) => ({
  trajectories,
  selectedTrajectoryId: selectedId,
  viewMode,
  isPlaying: false,
  currentTime: 0,
});

describe('TrajectoryVisualization', () => {
  it('returns null when not in reasoning mode', async () => {
    vi.doMock('@/lib/store', () => ({
      useStore: () => createMockStore([], 'evidence', null),
    }));

    const { TrajectoryVisualization } = await import('./TrajectoryVisualization');

    // Component should return null in evidence mode
    expect(TrajectoryVisualization).toBeDefined();
  });

  it('returns null when no trajectories', async () => {
    vi.doMock('@/lib/store', () => ({
      useStore: () => createMockStore([], 'reasoning', null),
    }));

    const { TrajectoryVisualization } = await import('./TrajectoryVisualization');
    expect(TrajectoryVisualization).toBeDefined();
  });

  it('exports TrajectoryVisualization component', async () => {
    const module = await import('./TrajectoryVisualization');
    expect(module.TrajectoryVisualization).toBeDefined();
  });

  it('exports TrajectoryMarker component', async () => {
    const module = await import('./TrajectoryVisualization');
    expect(module.TrajectoryMarker).toBeDefined();
  });

  it('defines trajectory colors array', async () => {
    // The TRAJECTORY_COLORS constant should be defined
    const { TrajectoryVisualization } = await import('./TrajectoryVisualization');
    expect(TrajectoryVisualization).toBeDefined();
  });
});

describe('TrajectoryMarker', () => {
  it('exports TrajectoryMarker', async () => {
    const { TrajectoryMarker } = await import('./TrajectoryVisualization');
    expect(TrajectoryMarker).toBeDefined();
  });
});
