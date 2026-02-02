'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { Header } from '@/components/layout/Header';
import * as api from '@/lib/api';
import type { Case } from '@/lib/types';
import { Plus, FolderOpen, Clock, ChevronRight, Loader2 } from 'lucide-react';

export default function CasesListPage() {
  const router = useRouter();
  const [cases, setCases] = useState<Case[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreating, setIsCreating] = useState(false);

  useEffect(() => {
    const loadCases = async () => {
      try {
        const data = await api.getCases();
        setCases(data || []);
      } catch (err) {
        console.error('Failed to load cases:', err);
      } finally {
        setIsLoading(false);
      }
    };

    loadCases();
  }, []);

  const handleCreateCase = async () => {
    setIsCreating(true);
    try {
      const newCase = await api.createCase(
        'New Investigation',
        `Created on ${new Date().toLocaleDateString()}`
      );
      router.push(`/cases/${newCase.id}`);
    } catch (err) {
      console.error('Failed to create case:', err);
      setIsCreating(false);
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  return (
    <div className="h-screen flex flex-col bg-[#111114]">
      <Header activeJobCount={0} onJobsClick={() => {}} />

      <div className="flex-1 overflow-auto">
        <div className="max-w-4xl mx-auto p-8">
          {/* Header */}
          <div className="flex items-center justify-between mb-8">
            <div>
              <h1 className="text-2xl font-semibold text-white">Cases</h1>
              <p className="text-sm text-[#8b8b96] mt-1">
                Manage your investigation cases
              </p>
            </div>
            <button
              onClick={handleCreateCase}
              disabled={isCreating}
              className="flex items-center gap-2 px-4 py-2 bg-[#3b82f6] hover:bg-[#2563eb] disabled:opacity-50 disabled:cursor-not-allowed rounded-lg transition-colors text-sm font-medium"
            >
              {isCreating ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Plus className="w-4 h-4" />
              )}
              New Case
            </button>
          </div>

          {/* Cases Grid */}
          {isLoading ? (
            <div className="flex items-center justify-center py-16">
              <Loader2 className="w-8 h-8 text-[#3b82f6] animate-spin" />
            </div>
          ) : cases.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="w-16 h-16 rounded-full bg-[#1f1f24] flex items-center justify-center mb-4">
                <FolderOpen className="w-8 h-8 text-[#606068]" />
              </div>
              <h3 className="text-lg font-medium text-white mb-2">No cases yet</h3>
              <p className="text-sm text-[#8b8b96] mb-6">
                Create your first investigation case to get started
              </p>
              <button
                onClick={handleCreateCase}
                disabled={isCreating}
                className="flex items-center gap-2 px-4 py-2 bg-[#3b82f6] hover:bg-[#2563eb] disabled:opacity-50 rounded-lg transition-colors text-sm font-medium"
              >
                <Plus className="w-4 h-4" />
                Create First Case
              </button>
            </div>
          ) : (
            <div className="grid gap-4">
              {cases.map((caseItem) => (
                <button
                  key={caseItem.id}
                  onClick={() => router.push(`/cases/${caseItem.id}`)}
                  className="w-full text-left p-4 bg-[#1f1f24] hover:bg-[#2a2a32] border border-[#2a2a32] hover:border-[#3b82f6]/30 rounded-lg transition-all group"
                >
                  <div className="flex items-center gap-4">
                    <div className="w-10 h-10 rounded-lg bg-[#3b82f6]/10 flex items-center justify-center flex-shrink-0">
                      <FolderOpen className="w-5 h-5 text-[#3b82f6]" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <h3 className="font-medium text-white truncate group-hover:text-[#3b82f6] transition-colors">
                        {caseItem.title}
                      </h3>
                      {caseItem.description && (
                        <p className="text-sm text-[#8b8b96] truncate mt-0.5">
                          {caseItem.description}
                        </p>
                      )}
                    </div>
                    <div className="flex items-center gap-4 text-xs text-[#606068]">
                      <div className="flex items-center gap-1.5">
                        <Clock className="w-3.5 h-3.5" />
                        {formatDate(caseItem.created_at)}
                      </div>
                      <ChevronRight className="w-4 h-4 text-[#606068] group-hover:text-[#3b82f6] transition-colors" />
                    </div>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
