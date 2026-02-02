import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import CaseDetailPage from './page';

// Mock next/navigation
const mockPush = vi.fn();
vi.mock('next/navigation', () => ({
  useParams: () => ({ id: 'test-case-id' }),
  useRouter: () => ({
    push: mockPush,
  }),
}));

// Mock next/dynamic to return mocked components directly
vi.mock('next/dynamic', () => ({
  default: (fn: () => Promise<{ SceneViewer: React.FC }>) => {
    return function DynamicComponent() {
      return <div data-testid="scene-viewer">Scene Viewer Mock</div>;
    };
  },
}));

// Mock API
const mockGetCase = vi.fn();
const mockGetTimeline = vi.fn();
const mockGetSnapshot = vi.fn();

vi.mock('@/lib/api', () => ({
  getCase: () => mockGetCase(),
  getTimeline: () => mockGetTimeline(),
  getSnapshot: () => mockGetSnapshot(),
}));

// Mock store
const mockSetCurrentCase = vi.fn();
const mockSetCommits = vi.fn();
const mockSetSceneGraph = vi.fn();
const mockSetJobs = vi.fn();

let mockStore = {
  currentCase: null as { id: string; title: string; description?: string } | null,
  commits: [],
  jobs: [],
  viewMode: 'evidence' as const,
  sceneGraph: null,
  setCurrentCase: mockSetCurrentCase,
  setCommits: mockSetCommits,
  setSceneGraph: mockSetSceneGraph,
  setJobs: mockSetJobs,
};

vi.mock('@/lib/store', () => ({
  useStore: () => mockStore,
}));

// Mock hooks
vi.mock('@/hooks/useUpload', () => ({
  useUpload: () => ({
    uploadFiles: vi.fn(),
    progress: [],
    isUploading: false,
  }),
}));

vi.mock('@/hooks/useJobs', () => ({
  useJobs: () => ({
    activeJobs: [],
    pollJob: vi.fn(),
  }),
}));

// Mock child components to simplify testing
vi.mock('@/components/layout/Header', () => ({
  Header: ({ activeJobCount, onJobsClick }: { activeJobCount: number; onJobsClick: () => void }) => (
    <header data-testid="header">
      <span>Active Jobs: {activeJobCount}</span>
      <button onClick={onJobsClick}>Toggle Jobs</button>
    </header>
  ),
}));

vi.mock('@/components/layout/Sidebar', () => ({
  Sidebar: () => <aside data-testid="sidebar">Sidebar</aside>,
}));

vi.mock('@/components/layout/ModeSelector', () => ({
  ModeSelector: () => <div data-testid="mode-selector">Mode Selector</div>,
}));

vi.mock('@/components/timeline/Timeline', () => ({
  Timeline: () => <div data-testid="timeline">Timeline</div>,
}));

vi.mock('@/components/panels/ModePanel', () => ({
  ModePanel: () => <div data-testid="mode-panel">Mode Panel</div>,
}));

vi.mock('@/components/jobs/JobProgress', () => ({
  JobProgress: ({ jobs }: { jobs: unknown[] }) => (
    <div data-testid="job-progress">Jobs: {jobs.length}</div>
  ),
}));

vi.mock('@/components/evidence/DropZone', () => ({
  DropOverlay: ({ isVisible }: { isVisible: boolean }) =>
    isVisible ? <div data-testid="drop-overlay">Drop files here</div> : null,
}));

describe('CaseDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockStore = {
      currentCase: null,
      commits: [],
      jobs: [],
      viewMode: 'evidence' as const,
      sceneGraph: null,
      setCurrentCase: mockSetCurrentCase,
      setCommits: mockSetCommits,
      setSceneGraph: mockSetSceneGraph,
      setJobs: mockSetJobs,
    };
    mockGetCase.mockResolvedValue({ id: 'test-case-id', title: 'Test Case', description: 'A test case' });
    mockGetTimeline.mockResolvedValue({ commits: [] });
    mockGetSnapshot.mockResolvedValue({ scenegraph: { objects: [], relationships: [] } });
  });

  it('shows loading state initially', () => {
    mockGetCase.mockImplementation(() => new Promise(() => {})); // Never resolves
    render(<CaseDetailPage />);

    expect(screen.getByText('Loading case...')).toBeInTheDocument();
  });

  it('loads and displays case data', async () => {
    mockStore.currentCase = { id: 'test-case-id', title: 'Loaded Test Case', description: 'Description' };
    mockStore.commits = [{ id: '1' }, { id: '2' }] as any;

    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(mockSetCurrentCase).toHaveBeenCalled();
    });

    await waitFor(() => {
      expect(mockSetCommits).toHaveBeenCalled();
    });
  });

  it('displays error state when case loading fails', async () => {
    mockGetCase.mockRejectedValue(new Error('Case not found'));
    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(screen.getByText('Case Not Found')).toBeInTheDocument();
    });

    expect(screen.getByText('Case not found')).toBeInTheDocument();
    expect(screen.getByText('Back to Dashboard')).toBeInTheDocument();
  });

  it('navigates back to dashboard on error button click', async () => {
    mockGetCase.mockRejectedValue(new Error('Not found'));
    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(screen.getByText('Back to Dashboard')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Back to Dashboard'));
    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('renders all main layout components when loaded', async () => {
    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(screen.getByTestId('header')).toBeInTheDocument();
    });

    expect(screen.getByTestId('sidebar')).toBeInTheDocument();
    expect(screen.getByTestId('mode-selector')).toBeInTheDocument();
    expect(screen.getByTestId('timeline')).toBeInTheDocument();
    expect(screen.getByTestId('mode-panel')).toBeInTheDocument();
    expect(screen.getByTestId('scene-viewer')).toBeInTheDocument();
  });

  it('displays case title when loaded', async () => {
    mockStore.currentCase = { id: 'test-case-id', title: 'My Investigation' };

    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(screen.getByText('My Investigation')).toBeInTheDocument();
    });
  });

  it('displays commit count', async () => {
    mockStore.currentCase = { id: 'test-case-id', title: 'Test' };
    mockStore.commits = [{ id: '1' }, { id: '2' }, { id: '3' }] as any;

    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(screen.getByText('3 commits')).toBeInTheDocument();
    });
  });

  it('displays truncated case ID', async () => {
    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(screen.getByText(/ID: test-cas\.\.\./)).toBeInTheDocument();
    });
  });

  it('handles snapshot loading error gracefully', async () => {
    mockGetSnapshot.mockRejectedValue(new Error('No snapshot'));

    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(mockSetSceneGraph).toHaveBeenCalledWith(null);
    });
  });

  it('shows job panel when toggle is clicked', async () => {
    mockStore.jobs = [{ id: 'job-1', type: 'reconstruction', status: 'running' }] as any;

    render(<CaseDetailPage />);

    // Wait for loading to finish
    await waitFor(() => {
      expect(screen.queryByText('Loading case...')).not.toBeInTheDocument();
    });

    // Find and click the toggle button
    const toggleButton = screen.getByText('Toggle Jobs');
    fireEvent.click(toggleButton);

    // Job panel should now be visible
    await waitFor(() => {
      expect(screen.getByTestId('job-progress')).toBeInTheDocument();
    });
  });

  it('navigates back when back button is clicked', async () => {
    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(screen.getByTitle('Back to Dashboard')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByTitle('Back to Dashboard'));
    expect(mockPush).toHaveBeenCalledWith('/');
  });

  it('shows default title when case has no title', async () => {
    mockStore.currentCase = { id: 'test', title: '' };

    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(screen.getByText('Untitled Case')).toBeInTheDocument();
    });
  });

  it('loads timeline data on mount', async () => {
    const commits = [
      { id: 'c1', type: 'upload_scan', summary: 'Uploaded files' },
      { id: 'c2', type: 'witness_statement', summary: 'Added witness' },
    ];
    mockGetTimeline.mockResolvedValue({ commits });

    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(mockSetCommits).toHaveBeenCalledWith(commits);
    });
  });

  it('loads scene snapshot on mount', async () => {
    const scenegraph = {
      objects: [{ id: 'obj1', type: 'evidence' }],
      relationships: [],
    };
    mockGetSnapshot.mockResolvedValue({ scenegraph });

    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(mockSetSceneGraph).toHaveBeenCalledWith(scenegraph);
    });
  });
});

describe('CaseDetailPage drag and drop', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetCase.mockResolvedValue({ id: 'test', title: 'Test' });
    mockGetTimeline.mockResolvedValue({ commits: [] });
    mockGetSnapshot.mockResolvedValue({ scenegraph: null });
  });

  it('does not show drop overlay by default', async () => {
    render(<CaseDetailPage />);

    await waitFor(() => {
      expect(screen.queryByTestId('drop-overlay')).not.toBeInTheDocument();
    });
  });
});
