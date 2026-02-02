'use client';

import { useState } from 'react';
import { Search, Plus, X, Infinity, Loader2, Download, Check } from 'lucide-react';
import { Input } from '@/components/ui/Input';
import { Button } from '@/components/ui/Button';
import { useStore } from '@/lib/store';
import { cn } from '@/lib/utils';
import * as api from '@/lib/api';

interface HeaderProps {
  activeJobCount?: number;
  onJobsClick?: () => void;
}

export function Header({ activeJobCount = 0, onJobsClick }: HeaderProps) {
  const { cases, currentCase, setCurrentCase } = useStore();
  const [searchQuery, setSearchQuery] = useState('');
  const [isExporting, setIsExporting] = useState(false);
  const [exportSuccess, setExportSuccess] = useState(false);

  const handleExport = async () => {
    if (!currentCase?.id || isExporting) return;

    setIsExporting(true);
    setExportSuccess(false);

    try {
      // Trigger export job
      const job = await api.triggerExport(currentCase.id, 'html');

      // Poll for completion
      let attempts = 0;
      const maxAttempts = 30; // 30 seconds max

      const pollInterval = setInterval(async () => {
        attempts++;
        try {
          const jobStatus = await api.getJob(job.id);

          if (jobStatus.status === 'done') {
            clearInterval(pollInterval);
            setIsExporting(false);
            setExportSuccess(true);

            // Get the report URL and open it
            if (jobStatus.output) {
              const output = jobStatus.output as { report_asset_key?: string };
              if (output.report_asset_key) {
                const reportUrl = api.getAssetUrl(output.report_asset_key);
                window.open(reportUrl, '_blank');
              }
            }

            // Reset success indicator after 3 seconds
            setTimeout(() => setExportSuccess(false), 3000);
          } else if (jobStatus.status === 'failed') {
            clearInterval(pollInterval);
            setIsExporting(false);
            console.error('Export failed:', jobStatus.error);
          } else if (attempts >= maxAttempts) {
            clearInterval(pollInterval);
            setIsExporting(false);
            console.error('Export timed out');
          }
        } catch (err) {
          console.error('Error polling export job:', err);
        }
      }, 1000);
    } catch (err) {
      console.error('Failed to start export:', err);
      setIsExporting(false);
    }
  };

  const openCases = cases.slice(0, 5); // Show max 5 tabs

  return (
    <header className="h-14 bg-[#111114] border-b border-[#1e1e24] flex items-center px-4 gap-4">
      {/* Logo */}
      <div className="flex items-center gap-2 pr-4 border-r border-[#2a2a32]">
        <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-[#6366f1] to-[#8b5cf6] flex items-center justify-center">
          <Infinity className="w-5 h-5 text-white" />
        </div>
        <span className="font-semibold text-[#f0f0f2] hidden sm:inline">SherlockOS</span>
      </div>

      {/* Search */}
      <div className="w-64">
        <Input
          icon={<Search className="w-4 h-4" />}
          placeholder="Search..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
        />
      </div>

      {/* Case Tabs */}
      <div className="flex-1 flex items-center gap-1 overflow-x-auto px-2">
        {openCases.map((caseItem) => (
          <button
            key={caseItem.id}
            onClick={() => setCurrentCase(caseItem)}
            className={cn(
              'flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm transition-all',
              'hover:bg-[#1f1f24] group',
              currentCase?.id === caseItem.id
                ? 'bg-[#1f1f24] text-[#f0f0f2]'
                : 'text-[#a0a0a8]'
            )}
          >
            <div
              className={cn(
                'w-4 h-4 rounded flex items-center justify-center text-xs',
                currentCase?.id === caseItem.id
                  ? 'bg-[#3b82f6]'
                  : 'bg-[#2a2a32]'
              )}
            >
              {caseItem.title.charAt(0).toUpperCase()}
            </div>
            <span className="max-w-32 truncate">{caseItem.title}</span>
            <X
              className={cn(
                'w-3.5 h-3.5 opacity-0 group-hover:opacity-100 transition-opacity',
                'hover:text-[#ef4444]'
              )}
              onClick={(e) => {
                e.stopPropagation();
                // Handle close case
              }}
            />
          </button>
        ))}

        <Button
          variant="ghost"
          size="icon"
          className="shrink-0 h-8 w-8"
          onClick={() => {
            // Handle new case
          }}
        >
          <Plus className="w-4 h-4" />
        </Button>
      </div>

      {/* Right actions */}
      <div className="flex items-center gap-2">
        {activeJobCount > 0 && (
          <button
            onClick={onJobsClick}
            className={cn(
              'flex items-center gap-2 px-3 py-1.5 rounded-lg',
              'bg-[#3b82f6]/10 hover:bg-[#3b82f6]/20 transition-colors'
            )}
          >
            <Loader2 className="w-3.5 h-3.5 text-[#3b82f6] animate-spin" />
            <span className="text-xs font-medium text-[#3b82f6]">
              {activeJobCount} job{activeJobCount > 1 ? 's' : ''}
            </span>
          </button>
        )}
        <Button
          variant="secondary"
          size="sm"
          onClick={handleExport}
          disabled={isExporting || !currentCase}
          className={cn(
            exportSuccess && 'bg-green-600/20 border-green-600/30 text-green-500'
          )}
        >
          {isExporting ? (
            <>
              <Loader2 className="w-3.5 h-3.5 mr-1.5 animate-spin" />
              Exporting...
            </>
          ) : exportSuccess ? (
            <>
              <Check className="w-3.5 h-3.5 mr-1.5" />
              Exported
            </>
          ) : (
            <>
              <Download className="w-3.5 h-3.5 mr-1.5" />
              Export
            </>
          )}
        </Button>
      </div>
    </header>
  );
}
