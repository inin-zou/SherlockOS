import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { Header } from './Header';

// Mock the store
const mockSetCurrentCase = vi.fn();
let mockCurrentCase: { id: string; title: string } | null = { id: 'case-123', title: 'Test Case' };
let mockCases: Array<{ id: string; title: string }> = [];

vi.mock('@/lib/store', () => ({
  useStore: () => ({
    cases: mockCases,
    currentCase: mockCurrentCase,
    setCurrentCase: mockSetCurrentCase,
  }),
}));

// Mock the API
vi.mock('@/lib/api', () => ({
  triggerExport: vi.fn().mockResolvedValue({ id: 'job-123' }),
  getJob: vi.fn().mockResolvedValue({
    id: 'job-123',
    status: 'done',
    output: { report_asset_key: 'reports/test.html' },
  }),
  getAssetUrl: vi.fn().mockReturnValue('https://example.com/reports/test.html'),
}));

describe('Header', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCurrentCase = { id: 'case-123', title: 'Test Case' };
    mockCases = [];
  });

  it('renders logo and title', () => {
    render(<Header />);
    expect(screen.getByText('SherlockOS')).toBeInTheDocument();
  });

  it('renders search input', () => {
    render(<Header />);
    expect(screen.getByPlaceholderText('Search...')).toBeInTheDocument();
  });

  it('renders Export button', () => {
    render(<Header />);
    expect(screen.getByText('Export')).toBeInTheDocument();
  });

  it('shows job count when jobs are active', () => {
    render(<Header activeJobCount={3} />);
    expect(screen.getByText('3 jobs')).toBeInTheDocument();
  });

  it('shows singular job text for 1 job', () => {
    render(<Header activeJobCount={1} />);
    expect(screen.getByText('1 job')).toBeInTheDocument();
  });

  it('hides job indicator when no active jobs', () => {
    render(<Header activeJobCount={0} />);
    expect(screen.queryByText(/job/)).not.toBeInTheDocument();
  });

  it('calls onJobsClick when job indicator clicked', () => {
    const onJobsClick = vi.fn();
    render(<Header activeJobCount={2} onJobsClick={onJobsClick} />);

    fireEvent.click(screen.getByText('2 jobs'));
    expect(onJobsClick).toHaveBeenCalled();
  });

  it('disables Export button when no current case', () => {
    mockCurrentCase = null;
    render(<Header />);

    const exportButton = screen.getByText('Export').closest('button');
    expect(exportButton).toBeDisabled();
  });

  it('triggers export when button clicked', async () => {
    const { triggerExport } = await import('@/lib/api');

    render(<Header />);

    const exportButton = screen.getByText('Export');
    fireEvent.click(exportButton);

    await waitFor(() => {
      expect(triggerExport).toHaveBeenCalledWith('case-123', 'html');
    });
  });

  it('shows Exporting state during export', async () => {
    render(<Header />);

    const exportButton = screen.getByText('Export');
    fireEvent.click(exportButton);

    await waitFor(() => {
      expect(screen.getByText('Exporting...')).toBeInTheDocument();
    });
  });

  it('renders case tabs when cases exist', () => {
    mockCases = [
      { id: 'case-1', title: 'Case Alpha' },
      { id: 'case-2', title: 'Case Beta' },
    ];

    render(<Header />);

    expect(screen.getByText('Case Alpha')).toBeInTheDocument();
    expect(screen.getByText('Case Beta')).toBeInTheDocument();
  });

  it('highlights current case tab', () => {
    mockCases = [
      { id: 'case-123', title: 'Current Case' },
      { id: 'case-456', title: 'Other Case' },
    ];
    mockCurrentCase = { id: 'case-123', title: 'Current Case' };

    render(<Header />);

    const currentTab = screen.getByText('Current Case').closest('button');
    expect(currentTab).toHaveClass('bg-[#1f1f24]');
  });

  it('switches case when tab clicked', () => {
    mockCases = [
      { id: 'case-123', title: 'Case A' },
      { id: 'case-456', title: 'Case B' },
    ];

    render(<Header />);

    fireEvent.click(screen.getByText('Case B'));
    expect(mockSetCurrentCase).toHaveBeenCalledWith({ id: 'case-456', title: 'Case B' });
  });
});
