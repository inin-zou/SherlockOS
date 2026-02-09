'use client';

import { useState } from 'react';
import { Search, Plus, X, Loader2, Download, Check, FileText } from 'lucide-react';
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
  const { cases, currentCase, setCurrentCase, removeCase } = useStore();
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
    <header className="h-14 bg-[#0a0a0c] border-b border-[#1e1e24] flex items-center px-4 gap-4">
      {/* Search - Moved to Far Left */}
      <div className="w-64">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-[#606068]" />
          <input
            type="text"
            placeholder="Search..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full h-9 bg-[#121214] border border-[#27272a] rounded-full pl-10 pr-3 text-sm text-[#f0f0f2] placeholder:text-[#606068] focus:outline-none focus:border-[#3b82f6]/50 transition-colors"
          />
        </div>
      </div>

      {/* Case Tabs */}
      <div className="flex-1 flex items-center gap-2 overflow-x-auto px-2 no-scrollbar">
        {openCases.map((caseItem) => (
          <button
            key={caseItem.id}
            onClick={() => setCurrentCase(caseItem)}
            title={`${caseItem.title} (${caseItem.id})`}
            className={cn(
              'flex items-center gap-2 px-3 h-9 rounded-full text-sm transition-all border shrink-0 max-w-[240px]',
              'group relative pr-8', // Added padding for close button
              currentCase?.id === caseItem.id
                ? 'bg-[#121214] text-[#f0f0f2] border-[#27272a]'
                : 'bg-transparent text-[#606068] border-transparent hover:bg-[#121214] hover:text-[#f0f0f2]'
            )}
          >
            <FileText className="w-4 h-4 shrink-0 opacity-50" />
            <span className="truncate block flex-1 text-left">
              {caseItem.title}
            </span>

            {/* Close Button */}
            <span
              className={cn(
                'absolute right-1 top-1/2 -translate-y-1/2 p-1 rounded-full',
                'opacity-0 group-hover:opacity-100 transition-opacity',
                'hover:bg-[#2a2a32] text-[#606068] hover:text-white'
              )}
              onClick={(e) => {
                e.stopPropagation();
                removeCase(caseItem.id);
              }}
            >
              <X className="w-3 h-3" />
            </span>
          </button>
        ))}

        <button
          className="w-8 h-8 rounded-full flex items-center justify-center text-[#606068] hover:text-[#f0f0f2] hover:bg-[#111114] transition-colors shrink-0"
          onClick={() => {
            // Handle new case (future feature)
          }}
          title="New Case"
        >
          <Plus className="w-5 h-5" />
        </button>
      </div>

      {/* Right actions */}
      <div className="flex items-center gap-3 shrink-0 ml-auto">
        {activeJobCount > 0 && (
          <button
            onClick={onJobsClick}
            className={cn(
              'flex items-center gap-2 px-3 py-1.5 rounded-lg',
              'bg-[#3b82f6]/10 hover:bg-[#3b82f6]/20 transition-colors'
            )}
            title="Active Jobs"
          >
            <Loader2 className="w-4 h-4 text-[#3b82f6] animate-spin" />
            <span className="text-xs font-medium text-[#3b82f6]">
              {activeJobCount}
            </span>
          </button>
        )}

        <Button
          variant="ghost"
          size="sm"
          onClick={handleExport}
          disabled={isExporting || !currentCase}
          className={cn(
            'h-8 px-3 text-xs font-medium gap-2',
            exportSuccess ? 'text-green-500' : 'text-[#606068] hover:text-[#f0f0f2]'
          )}
        >
          {isExporting ? (
            <Loader2 className="w-4 h-4 animate-spin" />
          ) : exportSuccess ? (
            <Check className="w-4 h-4" />
          ) : (
            <Download className="w-4 h-4" />
          )}
          <span className="hidden sm:inline">
            {isExporting ? 'Exporting...' : exportSuccess ? 'Exported' : 'Export'}
          </span>
        </Button>
      </div>
    </header>
  );
}
