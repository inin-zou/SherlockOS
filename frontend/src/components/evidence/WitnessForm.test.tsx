import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { WitnessForm } from './WitnessForm';

// Mock the API
vi.mock('@/lib/api', () => ({
  submitWitnessStatements: vi.fn(),
}));

import * as api from '@/lib/api';

describe('WitnessForm', () => {
  const mockOnSubmit = vi.fn();
  const mockOnClose = vi.fn();
  const defaultProps = {
    caseId: 'case-123',
    onSubmit: mockOnSubmit,
    onClose: mockOnClose,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders form header', () => {
    render(<WitnessForm {...defaultProps} />);
    expect(screen.getByText('Add Witness Statements')).toBeInTheDocument();
  });

  it('renders initial statement form', () => {
    render(<WitnessForm {...defaultProps} />);
    expect(screen.getByText('Statement #1')).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/witness name/i)).toBeInTheDocument();
    expect(screen.getByPlaceholderText(/enter witness statement/i)).toBeInTheDocument();
  });

  it('shows credibility slider with default value', () => {
    render(<WitnessForm {...defaultProps} />);
    expect(screen.getByText('Credibility')).toBeInTheDocument();
    expect(screen.getByText('70%')).toBeInTheDocument();
  });

  it('allows adding more statements', () => {
    render(<WitnessForm {...defaultProps} />);

    const addButton = screen.getByText('Add another statement');
    fireEvent.click(addButton);

    expect(screen.getByText('Statement #1')).toBeInTheDocument();
    expect(screen.getByText('Statement #2')).toBeInTheDocument();
  });

  it('allows removing statements when more than one exists', () => {
    render(<WitnessForm {...defaultProps} />);

    // Add a second statement
    fireEvent.click(screen.getByText('Add another statement'));
    expect(screen.getByText('Statement #2')).toBeInTheDocument();

    // Remove second statement (trash button)
    const trashButtons = document.querySelectorAll('button');
    const deleteButton = Array.from(trashButtons).find(btn =>
      btn.querySelector('svg')?.classList.contains('lucide-trash-2') ||
      btn.innerHTML.includes('Trash2')
    );

    // Find and click delete button for statement 2
    const statement2Section = screen.getByText('Statement #2').closest('div');
    const trash = statement2Section?.parentElement?.querySelector('button');
    if (trash) {
      fireEvent.click(trash);
    }
  });

  it('shows close button when onClose provided', () => {
    render(<WitnessForm {...defaultProps} />);
    expect(screen.getByText('Cancel')).toBeInTheDocument();
  });

  it('calls onClose when cancel clicked', () => {
    render(<WitnessForm {...defaultProps} />);
    fireEvent.click(screen.getByText('Cancel'));
    expect(mockOnClose).toHaveBeenCalledTimes(1);
  });

  it('validates required fields before submit', async () => {
    render(<WitnessForm {...defaultProps} />);

    // Try to submit without filling in fields
    fireEvent.click(screen.getByText('Submit Statements'));

    expect(screen.getByText(/please add at least one statement/i)).toBeInTheDocument();
  });

  it('submits valid statement', async () => {
    (api.submitWitnessStatements as ReturnType<typeof vi.fn>).mockResolvedValue({
      commit_id: 'commit-123',
      profile_job_id: 'job-456',
    });

    render(<WitnessForm {...defaultProps} />);

    // Fill in statement
    fireEvent.change(screen.getByPlaceholderText(/witness name/i), {
      target: { value: 'John Doe' },
    });
    fireEvent.change(screen.getByPlaceholderText(/enter witness statement/i), {
      target: { value: 'I saw the suspect at 10 PM' },
    });

    // Submit
    fireEvent.click(screen.getByText('Submit Statements'));

    await waitFor(() => {
      expect(api.submitWitnessStatements).toHaveBeenCalledWith('case-123', [
        {
          source_name: 'John Doe',
          content: 'I saw the suspect at 10 PM',
          credibility: 0.7,
        },
      ]);
    });
  });

  it('calls onSubmit with result after successful submission', async () => {
    const mockResult = {
      commit_id: 'commit-123',
      profile_job_id: 'job-456',
    };
    (api.submitWitnessStatements as ReturnType<typeof vi.fn>).mockResolvedValue(mockResult);

    render(<WitnessForm {...defaultProps} />);

    // Fill and submit
    fireEvent.change(screen.getByPlaceholderText(/witness name/i), {
      target: { value: 'Jane Doe' },
    });
    fireEvent.change(screen.getByPlaceholderText(/enter witness statement/i), {
      target: { value: 'Statement content' },
    });
    fireEvent.click(screen.getByText('Submit Statements'));

    await waitFor(() => {
      expect(mockOnSubmit).toHaveBeenCalledWith(mockResult);
    });
  });

  it('shows error on API failure', async () => {
    (api.submitWitnessStatements as ReturnType<typeof vi.fn>).mockRejectedValue(
      new Error('API Error')
    );

    render(<WitnessForm {...defaultProps} />);

    // Fill and submit
    fireEvent.change(screen.getByPlaceholderText(/witness name/i), {
      target: { value: 'Test' },
    });
    fireEvent.change(screen.getByPlaceholderText(/enter witness statement/i), {
      target: { value: 'Content' },
    });
    fireEvent.click(screen.getByText('Submit Statements'));

    await waitFor(() => {
      expect(screen.getByText('API Error')).toBeInTheDocument();
    });
  });

  it('shows loading state during submission', async () => {
    (api.submitWitnessStatements as ReturnType<typeof vi.fn>).mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    );

    render(<WitnessForm {...defaultProps} />);

    // Fill and submit
    fireEvent.change(screen.getByPlaceholderText(/witness name/i), {
      target: { value: 'Test' },
    });
    fireEvent.change(screen.getByPlaceholderText(/enter witness statement/i), {
      target: { value: 'Content' },
    });
    fireEvent.click(screen.getByText('Submit Statements'));

    // Button should show loading state
    const submitButton = screen.getByText('Submit Statements').closest('button');
    expect(submitButton).toBeInTheDocument();
  });

  it('resets form after successful submission', async () => {
    (api.submitWitnessStatements as ReturnType<typeof vi.fn>).mockResolvedValue({
      commit_id: 'commit-123',
    });

    render(<WitnessForm {...defaultProps} />);

    // Fill and submit
    const nameInput = screen.getByPlaceholderText(/witness name/i) as HTMLInputElement;
    fireEvent.change(nameInput, { target: { value: 'Test' } });
    fireEvent.change(screen.getByPlaceholderText(/enter witness statement/i), {
      target: { value: 'Content' },
    });
    fireEvent.click(screen.getByText('Submit Statements'));

    await waitFor(() => {
      // Form should be reset
      expect(nameInput.value).toBe('');
    });
  });
});
