import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ErrorBoundary, CompactErrorBoundary, SceneErrorBoundary } from './ErrorBoundary';

// Component that throws an error
function ThrowError({ shouldThrow }: { shouldThrow: boolean }) {
  if (shouldThrow) {
    throw new Error('Test error message');
  }
  return <div>Normal content</div>;
}

// Suppress console.error for cleaner test output
beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

describe('ErrorBoundary', () => {
  it('renders children when no error', () => {
    render(
      <ErrorBoundary>
        <div>Test content</div>
      </ErrorBoundary>
    );

    expect(screen.getByText('Test content')).toBeInTheDocument();
  });

  it('renders fallback UI when error occurs', () => {
    render(
      <ErrorBoundary>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.getByText('Test error message')).toBeInTheDocument();
    expect(screen.getByText('Try Again')).toBeInTheDocument();
  });

  it('renders custom fallback when provided', () => {
    render(
      <ErrorBoundary fallback={<div>Custom fallback</div>}>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('Custom fallback')).toBeInTheDocument();
  });

  it('resets error state when Try Again is clicked', () => {
    const onReset = vi.fn();
    const { rerender } = render(
      <ErrorBoundary onReset={onReset}>
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(screen.getByText('Something went wrong')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Try Again'));

    expect(onReset).toHaveBeenCalled();
  });

  it('logs error with component name', () => {
    const consoleSpy = vi.spyOn(console, 'error');

    render(
      <ErrorBoundary name="TestComponent">
        <ThrowError shouldThrow={true} />
      </ErrorBoundary>
    );

    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('ErrorBoundary [TestComponent]'),
      expect.any(Error),
      expect.any(Object)
    );
  });
});

describe('CompactErrorBoundary', () => {
  it('renders children when no error', () => {
    render(
      <CompactErrorBoundary>
        <div>Compact content</div>
      </CompactErrorBoundary>
    );

    expect(screen.getByText('Compact content')).toBeInTheDocument();
  });

  it('renders compact error UI when error occurs', () => {
    render(
      <CompactErrorBoundary>
        <ThrowError shouldThrow={true} />
      </CompactErrorBoundary>
    );

    expect(screen.getByText('Error loading component')).toBeInTheDocument();
    expect(screen.getByText('Retry')).toBeInTheDocument();
  });

  it('resets on retry click', () => {
    const onReset = vi.fn();

    render(
      <CompactErrorBoundary onReset={onReset}>
        <ThrowError shouldThrow={true} />
      </CompactErrorBoundary>
    );

    fireEvent.click(screen.getByText('Retry'));
    expect(onReset).toHaveBeenCalled();
  });
});

describe('SceneErrorBoundary', () => {
  it('renders children when no error', () => {
    render(
      <SceneErrorBoundary>
        <div>3D Scene content</div>
      </SceneErrorBoundary>
    );

    expect(screen.getByText('3D Scene content')).toBeInTheDocument();
  });

  it('renders scene-specific error UI when error occurs', () => {
    render(
      <SceneErrorBoundary>
        <ThrowError shouldThrow={true} />
      </SceneErrorBoundary>
    );

    expect(screen.getByText('3D Scene Error')).toBeInTheDocument();
    expect(screen.getByText(/WebGL compatibility/)).toBeInTheDocument();
    expect(screen.getByText('Reload Scene')).toBeInTheDocument();
  });

  it('resets on reload click', () => {
    const onReset = vi.fn();

    render(
      <SceneErrorBoundary onReset={onReset}>
        <ThrowError shouldThrow={true} />
      </SceneErrorBoundary>
    );

    fireEvent.click(screen.getByText('Reload Scene'));
    expect(onReset).toHaveBeenCalled();
  });
});
