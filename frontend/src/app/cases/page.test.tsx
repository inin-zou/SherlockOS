import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import CasesListPage from './page';

// Mock next/navigation
const mockPush = vi.fn();
vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: mockPush,
  }),
}));

// Mock API
const mockGetCases = vi.fn();
const mockCreateCase = vi.fn();

vi.mock('@/lib/api', () => ({
  getCases: () => mockGetCases(),
  createCase: (title: string, description: string) => mockCreateCase(title, description),
}));

// Mock Header component
vi.mock('@/components/layout/Header', () => ({
  Header: ({ activeJobCount, onJobsClick }: { activeJobCount: number; onJobsClick: () => void }) => (
    <header data-testid="header">
      <span>Jobs: {activeJobCount}</span>
      <button onClick={onJobsClick}>Jobs Button</button>
    </header>
  ),
}));

describe('CasesListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetCases.mockResolvedValue([]);
    mockCreateCase.mockResolvedValue({ id: 'new-case-id', title: 'New Investigation' });
  });

  it('renders the page title', async () => {
    mockGetCases.mockResolvedValue([]);
    render(<CasesListPage />);

    await waitFor(() => {
      expect(screen.getByText('Cases')).toBeInTheDocument();
    });
  });

  it('shows loading spinner initially', () => {
    mockGetCases.mockImplementation(() => new Promise(() => {})); // Never resolves
    render(<CasesListPage />);

    // Loading spinner should be visible - check for the animate-spin class
    const spinner = document.querySelector('.animate-spin');
    expect(spinner).toBeTruthy();
  });

  it('displays empty state when no cases exist', async () => {
    mockGetCases.mockResolvedValue([]);
    render(<CasesListPage />);

    await waitFor(() => {
      expect(screen.getByText('No cases yet')).toBeInTheDocument();
    });

    expect(screen.getByText('Create your first investigation case to get started')).toBeInTheDocument();
  });

  it('displays list of cases when cases exist', async () => {
    mockGetCases.mockResolvedValue([
      { id: 'case-1', title: 'Test Case 1', description: 'Description 1', created_at: '2024-01-15T10:00:00Z' },
      { id: 'case-2', title: 'Test Case 2', description: 'Description 2', created_at: '2024-01-16T10:00:00Z' },
    ]);

    render(<CasesListPage />);

    await waitFor(() => {
      expect(screen.getByText('Test Case 1')).toBeInTheDocument();
    });

    expect(screen.getByText('Test Case 2')).toBeInTheDocument();
    expect(screen.getByText('Description 1')).toBeInTheDocument();
    expect(screen.getByText('Description 2')).toBeInTheDocument();
  });

  it('navigates to case detail when case is clicked', async () => {
    mockGetCases.mockResolvedValue([
      { id: 'case-123', title: 'Click Me Case', description: 'Test', created_at: '2024-01-15T10:00:00Z' },
    ]);

    render(<CasesListPage />);

    await waitFor(() => {
      expect(screen.getByText('Click Me Case')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Click Me Case'));
    expect(mockPush).toHaveBeenCalledWith('/cases/case-123');
  });

  it('creates new case when New Case button is clicked', async () => {
    mockGetCases.mockResolvedValue([]);
    render(<CasesListPage />);

    await waitFor(() => {
      expect(screen.getByText('New Case')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('New Case'));

    await waitFor(() => {
      expect(mockCreateCase).toHaveBeenCalledWith(
        'New Investigation',
        expect.stringContaining('Created on')
      );
    });

    expect(mockPush).toHaveBeenCalledWith('/cases/new-case-id');
  });

  it('creates new case from empty state button', async () => {
    mockGetCases.mockResolvedValue([]);
    render(<CasesListPage />);

    await waitFor(() => {
      expect(screen.getByText('Create First Case')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('Create First Case'));

    await waitFor(() => {
      expect(mockCreateCase).toHaveBeenCalled();
    });
  });

  it('handles API error gracefully', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    mockGetCases.mockRejectedValue(new Error('API Error'));

    render(<CasesListPage />);

    await waitFor(() => {
      expect(consoleSpy).toHaveBeenCalled();
    });

    // Should show empty state on error
    await waitFor(() => {
      expect(screen.getByText('No cases yet')).toBeInTheDocument();
    });

    consoleSpy.mockRestore();
  });

  it('formats dates correctly', async () => {
    mockGetCases.mockResolvedValue([
      { id: 'case-1', title: 'Date Test', description: '', created_at: '2024-01-15T10:00:00Z' },
    ]);

    render(<CasesListPage />);

    await waitFor(() => {
      expect(screen.getByText('Jan 15, 2024')).toBeInTheDocument();
    });
  });

  it('disables buttons while creating case', async () => {
    mockGetCases.mockResolvedValue([]);
    mockCreateCase.mockImplementation(() => new Promise(() => {})); // Never resolves

    render(<CasesListPage />);

    await waitFor(() => {
      expect(screen.getByText('New Case')).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText('New Case'));

    await waitFor(() => {
      const buttons = screen.getAllByRole('button');
      const newCaseButtons = buttons.filter(b => b.textContent?.includes('Case'));
      newCaseButtons.forEach(btn => {
        expect(btn).toBeDisabled();
      });
    });
  });
});
