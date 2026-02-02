import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Sidebar } from './Sidebar';

// Mock store
const mockToggleFolder = vi.fn();
let mockStore = {
  evidenceFolders: [],
  toggleFolder: mockToggleFolder,
  sidebarWidth: 280,
  commits: [],
  viewMode: 'evidence' as const,
};

vi.mock('@/lib/store', () => ({
  useStore: () => mockStore,
}));

// Mock child components
vi.mock('@/components/evidence/DropZone', () => ({
  DropZone: ({ onFilesDropped, isUploading, disabled }: {
    onFilesDropped: () => void;
    isUploading: boolean;
    disabled: boolean
  }) => (
    <div data-testid="dropzone" data-uploading={isUploading} data-disabled={disabled}>
      <button onClick={onFilesDropped}>Upload</button>
    </div>
  ),
}));

vi.mock('@/components/evidence/WitnessForm', () => ({
  WitnessForm: ({ caseId, onSubmit }: { caseId: string; onSubmit: () => void }) => (
    <div data-testid="witness-form" data-case-id={caseId}>
      <button onClick={onSubmit}>Submit Witness</button>
    </div>
  ),
}));

vi.mock('@/components/timeline/CommitTimeline', () => ({
  CommitTimeline: ({ commits }: { commits: unknown[] }) => (
    <div data-testid="commit-timeline">
      Commits: {commits.length}
    </div>
  ),
}));

vi.mock('@/lib/utils', () => ({
  cn: (...classes: (string | boolean | undefined)[]) => classes.filter(Boolean).join(' '),
}));

describe('Sidebar', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockStore = {
      evidenceFolders: [],
      toggleFolder: mockToggleFolder,
      sidebarWidth: 280,
      commits: [],
      viewMode: 'evidence' as const,
    };
  });

  it('renders navigation items', () => {
    render(<Sidebar />);

    expect(screen.getByTitle('Overview')).toBeInTheDocument();
    expect(screen.getByTitle('Evidence')).toBeInTheDocument();
    expect(screen.getByTitle('Witness')).toBeInTheDocument();
    expect(screen.getByTitle('Suspects')).toBeInTheDocument();
    expect(screen.getByTitle('Reasoning')).toBeInTheDocument();
    expect(screen.getByTitle('Settings')).toBeInTheDocument();
  });

  it('shows Evidence Archive by default', () => {
    render(<Sidebar />);
    expect(screen.getByText('Evidence Archive')).toBeInTheDocument();
  });

  it('shows default demo folders when no evidence folders provided', () => {
    render(<Sidebar />);

    expect(screen.getByText('Environment')).toBeInTheDocument();
    expect(screen.getByText('Ground Truth')).toBeInTheDocument();
    expect(screen.getByText('Electronic Logs')).toBeInTheDocument();
    expect(screen.getByText('Testimonials')).toBeInTheDocument();
  });

  it('shows custom evidence folders when provided', () => {
    mockStore.evidenceFolders = [
      { id: '1', name: 'Custom Folder', icon: 'Folder', isOpen: false, items: [] },
      { id: '2', name: 'Another Folder', icon: 'Folder', isOpen: true, items: [] },
    ];

    render(<Sidebar />);

    expect(screen.getByText('Custom Folder')).toBeInTheDocument();
    expect(screen.getByText('Another Folder')).toBeInTheDocument();
    expect(screen.queryByText('Environment')).not.toBeInTheDocument();
  });

  it('toggles folder when clicked', () => {
    mockStore.evidenceFolders = [
      { id: 'folder-1', name: 'Test Folder', icon: 'Folder', isOpen: false, items: [] },
    ];

    render(<Sidebar />);

    fireEvent.click(screen.getByText('Test Folder'));
    expect(mockToggleFolder).toHaveBeenCalledWith('folder-1');
  });

  it('shows folder items when folder is open', () => {
    mockStore.evidenceFolders = [
      {
        id: '1',
        name: 'Open Folder',
        icon: 'Folder',
        isOpen: true,
        items: [
          { id: 'item-1', name: 'Document.pdf', type: 'pdf' },
          { id: 'item-2', name: 'Image.jpg', type: 'image' },
        ],
      },
    ];

    render(<Sidebar />);

    expect(screen.getByText('Document.pdf')).toBeInTheDocument();
    expect(screen.getByText('Image.jpg')).toBeInTheDocument();
  });

  it('hides folder items when folder is closed', () => {
    mockStore.evidenceFolders = [
      {
        id: '1',
        name: 'Closed Folder',
        icon: 'Folder',
        isOpen: false,
        items: [
          { id: 'item-1', name: 'Hidden.pdf', type: 'pdf' },
        ],
      },
    ];

    render(<Sidebar />);

    expect(screen.queryByText('Hidden.pdf')).not.toBeInTheDocument();
  });

  it('displays item count for each folder', () => {
    mockStore.evidenceFolders = [
      {
        id: '1',
        name: 'Folder With Items',
        icon: 'Folder',
        isOpen: false,
        items: [
          { id: '1', name: 'a', type: 'pdf' },
          { id: '2', name: 'b', type: 'pdf' },
          { id: '3', name: 'c', type: 'pdf' },
        ],
      },
    ];

    render(<Sidebar />);
    expect(screen.getByText('3')).toBeInTheDocument();
  });

  it('switches to Witness tab when clicked', () => {
    render(<Sidebar caseId="case-123" />);

    fireEvent.click(screen.getByTitle('Witness'));
    expect(screen.getByTestId('witness-form')).toBeInTheDocument();
  });

  it('switches to Overview tab showing commit timeline', () => {
    mockStore.commits = [{ id: '1' }, { id: '2' }] as any;
    render(<Sidebar />);

    fireEvent.click(screen.getByTitle('Overview'));
    expect(screen.getByTestId('commit-timeline')).toBeInTheDocument();
    expect(screen.getByText('Commits: 2')).toBeInTheDocument();
  });

  it('shows suspects empty state', () => {
    render(<Sidebar />);

    fireEvent.click(screen.getByTitle('Suspects'));
    expect(screen.getByText('No suspect profile')).toBeInTheDocument();
  });

  it('shows reasoning empty state', () => {
    render(<Sidebar />);

    fireEvent.click(screen.getByTitle('Reasoning'));
    expect(screen.getByText('No analysis yet')).toBeInTheDocument();
  });

  it('shows settings placeholder message', () => {
    render(<Sidebar />);

    fireEvent.click(screen.getByTitle('Settings'));
    expect(screen.getByText(/Case settings and configuration/)).toBeInTheDocument();
  });

  it('renders DropZone when onUpload is provided', () => {
    const mockUpload = vi.fn();
    render(<Sidebar onUpload={mockUpload} />);

    expect(screen.getByTestId('dropzone')).toBeInTheDocument();
  });

  it('disables DropZone when no caseId', () => {
    const mockUpload = vi.fn();
    render(<Sidebar onUpload={mockUpload} />);

    expect(screen.getByTestId('dropzone')).toHaveAttribute('data-disabled', 'true');
  });

  it('enables DropZone when caseId is provided', () => {
    const mockUpload = vi.fn();
    render(<Sidebar caseId="case-123" onUpload={mockUpload} />);

    expect(screen.getByTestId('dropzone')).toHaveAttribute('data-disabled', 'false');
  });

  it('shows placeholder drop zone when no onUpload provided', () => {
    render(<Sidebar />);

    expect(screen.getByText('Drop files here')).toBeInTheDocument();
  });

  it('passes upload progress to DropZone', () => {
    const mockUpload = vi.fn();
    const progress = [{ filename: 'test.jpg', progress: 50 }];

    render(<Sidebar onUpload={mockUpload} uploadProgress={progress as any} />);
    expect(screen.getByTestId('dropzone')).toBeInTheDocument();
  });

  it('indicates uploading state in DropZone', () => {
    const mockUpload = vi.fn();
    render(<Sidebar onUpload={mockUpload} isUploading={true} />);

    expect(screen.getByTestId('dropzone')).toHaveAttribute('data-uploading', 'true');
  });

  it('uses custom sidebar width from store', () => {
    mockStore.sidebarWidth = 320;
    render(<Sidebar />);

    const sidebar = document.querySelector('aside');
    expect(sidebar).toHaveStyle({ width: '320px' });
  });

  it('highlights active navigation item', () => {
    render(<Sidebar />);

    // Evidence is active by default
    const evidenceBtn = screen.getByTitle('Evidence');
    expect(evidenceBtn).toHaveClass('bg-[#1f1f24]');

    // Click on Witness
    const witnessBtn = screen.getByTitle('Witness');
    fireEvent.click(witnessBtn);
    expect(witnessBtn).toHaveClass('bg-[#1f1f24]');
    expect(evidenceBtn).not.toHaveClass('bg-[#1f1f24]');
  });
});
