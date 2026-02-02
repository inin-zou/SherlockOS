'use client';

import { Component, type ReactNode } from 'react';
import { AlertTriangle, RefreshCw } from 'lucide-react';
import { Button } from './Button';

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
  onReset?: () => void;
  name?: string; // For identifying which boundary caught the error
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error(`ErrorBoundary [${this.props.name || 'unknown'}] caught error:`, error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
    this.props.onReset?.();
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="flex flex-col items-center justify-center p-6 min-h-[200px] bg-[#111114] border border-[#2a2a32] rounded-lg">
          <div className="w-12 h-12 rounded-xl bg-[#ef4444]/20 flex items-center justify-center mb-4">
            <AlertTriangle className="w-6 h-6 text-[#ef4444]" />
          </div>
          <h3 className="text-sm font-medium text-[#f0f0f2] mb-2">Something went wrong</h3>
          <p className="text-xs text-[#606068] text-center max-w-sm mb-4">
            {this.state.error?.message || 'An unexpected error occurred'}
          </p>
          <Button variant="secondary" size="sm" onClick={this.handleReset}>
            <RefreshCw className="w-4 h-4" />
            Try Again
          </Button>
        </div>
      );
    }

    return this.props.children;
  }
}

/**
 * Compact error boundary for smaller components
 */
export class CompactErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error(`CompactErrorBoundary [${this.props.name || 'unknown'}] caught error:`, error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
    this.props.onReset?.();
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="flex items-center gap-2 p-3 bg-[#ef4444]/10 border border-[#ef4444]/20 rounded-lg">
          <AlertTriangle className="w-4 h-4 text-[#ef4444] shrink-0" />
          <span className="text-xs text-[#ef4444]">Error loading component</span>
          <button
            onClick={this.handleReset}
            className="ml-auto text-xs text-[#a0a0a8] hover:text-[#f0f0f2]"
          >
            Retry
          </button>
        </div>
      );
    }

    return this.props.children;
  }
}

/**
 * Scene-specific error boundary with 3D fallback
 */
export class SceneErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('SceneErrorBoundary caught error:', error, errorInfo);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
    this.props.onReset?.();
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="w-full h-full flex flex-col items-center justify-center bg-[#0a0a0c]">
          <div className="w-16 h-16 rounded-2xl bg-[#ef4444]/20 flex items-center justify-center mb-4">
            <AlertTriangle className="w-8 h-8 text-[#ef4444]" />
          </div>
          <h3 className="text-sm font-medium text-[#f0f0f2] mb-2">3D Scene Error</h3>
          <p className="text-xs text-[#606068] text-center max-w-sm mb-4">
            Failed to load the 3D scene viewer. This may be due to WebGL compatibility issues.
          </p>
          <Button variant="secondary" size="sm" onClick={this.handleReset}>
            <RefreshCw className="w-4 h-4" />
            Reload Scene
          </Button>
        </div>
      );
    }

    return this.props.children;
  }
}
