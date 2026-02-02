'use client';

import { useState } from 'react';
import { MessageSquare, Send, X, Plus, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/Button';
import { cn } from '@/lib/utils';
import * as api from '@/lib/api';

interface WitnessFormProps {
  caseId: string;
  onSubmit?: (result: { commit_id: string; profile_job_id?: string }) => void;
  onClose?: () => void;
  className?: string;
}

interface Statement {
  id: string;
  source_name: string;
  content: string;
  credibility: number;
}

export function WitnessForm({ caseId, onSubmit, onClose, className }: WitnessFormProps) {
  const [statements, setStatements] = useState<Statement[]>([
    { id: '1', source_name: '', content: '', credibility: 0.7 },
  ]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const addStatement = () => {
    setStatements([
      ...statements,
      { id: Date.now().toString(), source_name: '', content: '', credibility: 0.7 },
    ]);
  };

  const removeStatement = (id: string) => {
    if (statements.length > 1) {
      setStatements(statements.filter((s) => s.id !== id));
    }
  };

  const updateStatement = (id: string, field: keyof Statement, value: string | number) => {
    setStatements(
      statements.map((s) => (s.id === id ? { ...s, [field]: value } : s))
    );
  };

  const handleSubmit = async () => {
    // Validate
    const validStatements = statements.filter(
      (s) => s.source_name.trim() && s.content.trim()
    );

    if (validStatements.length === 0) {
      setError('Please add at least one statement with source name and content');
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      const result = await api.submitWitnessStatements(
        caseId,
        validStatements.map((s) => ({
          source_name: s.source_name,
          content: s.content,
          credibility: s.credibility,
        }))
      );

      onSubmit?.(result);

      // Reset form
      setStatements([{ id: '1', source_name: '', content: '', credibility: 0.7 }]);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit statements');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className={cn('bg-[#111114] border border-[#2a2a32] rounded-lg', className)}>
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 border-b border-[#2a2a32]">
        <div className="flex items-center gap-2">
          <MessageSquare className="w-4 h-4 text-[#8b5cf6]" />
          <h3 className="text-sm font-medium text-[#f0f0f2]">Add Witness Statements</h3>
        </div>
        {onClose && (
          <button
            onClick={onClose}
            className="p-1 hover:bg-[#1f1f24] rounded transition-colors text-[#606068] hover:text-[#a0a0a8]"
          >
            <X className="w-4 h-4" />
          </button>
        )}
      </div>

      {/* Statements */}
      <div className="p-4 space-y-4 max-h-96 overflow-y-auto">
        {statements.map((statement, index) => (
          <div
            key={statement.id}
            className="p-3 bg-[#1f1f24] rounded-lg space-y-3"
          >
            <div className="flex items-center justify-between">
              <span className="text-xs text-[#606068]">Statement #{index + 1}</span>
              {statements.length > 1 && (
                <button
                  onClick={() => removeStatement(statement.id)}
                  className="p-1 hover:bg-[#2a2a32] rounded transition-colors text-[#606068] hover:text-[#ef4444]"
                >
                  <Trash2 className="w-3 h-3" />
                </button>
              )}
            </div>

            {/* Source name */}
            <input
              type="text"
              value={statement.source_name}
              onChange={(e) => updateStatement(statement.id, 'source_name', e.target.value)}
              placeholder="Witness name (e.g., Security Guard A)"
              className={cn(
                'w-full px-3 py-2 bg-[#111114] border border-[#2a2a32] rounded-lg',
                'text-sm text-[#f0f0f2] placeholder:text-[#606068]',
                'focus:outline-none focus:ring-1 focus:ring-[#3b82f6]'
              )}
            />

            {/* Content */}
            <textarea
              value={statement.content}
              onChange={(e) => updateStatement(statement.id, 'content', e.target.value)}
              placeholder="Enter witness statement..."
              rows={3}
              className={cn(
                'w-full px-3 py-2 bg-[#111114] border border-[#2a2a32] rounded-lg',
                'text-sm text-[#f0f0f2] placeholder:text-[#606068]',
                'focus:outline-none focus:ring-1 focus:ring-[#3b82f6]',
                'resize-none'
              )}
            />

            {/* Credibility slider */}
            <div className="space-y-1">
              <div className="flex items-center justify-between text-xs">
                <span className="text-[#606068]">Credibility</span>
                <span className="text-[#a0a0a8]">{Math.round(statement.credibility * 100)}%</span>
              </div>
              <input
                type="range"
                min="0"
                max="100"
                value={statement.credibility * 100}
                onChange={(e) =>
                  updateStatement(statement.id, 'credibility', parseInt(e.target.value) / 100)
                }
                className="w-full h-1.5 bg-[#2a2a32] rounded-full appearance-none cursor-pointer
                  [&::-webkit-slider-thumb]:appearance-none
                  [&::-webkit-slider-thumb]:w-3
                  [&::-webkit-slider-thumb]:h-3
                  [&::-webkit-slider-thumb]:rounded-full
                  [&::-webkit-slider-thumb]:bg-[#8b5cf6]
                  [&::-webkit-slider-thumb]:cursor-pointer"
              />
            </div>
          </div>
        ))}

        {/* Add button */}
        <button
          onClick={addStatement}
          className={cn(
            'w-full flex items-center justify-center gap-2 py-2',
            'border border-dashed border-[#2a2a32] rounded-lg',
            'text-sm text-[#606068] hover:text-[#a0a0a8] hover:border-[#3b82f6]/50',
            'transition-colors'
          )}
        >
          <Plus className="w-4 h-4" />
          Add another statement
        </button>
      </div>

      {/* Error */}
      {error && (
        <div className="px-4 pb-2">
          <p className="text-xs text-[#ef4444]">{error}</p>
        </div>
      )}

      {/* Footer */}
      <div className="flex items-center justify-end gap-2 px-4 py-3 border-t border-[#2a2a32]">
        {onClose && (
          <Button variant="ghost" size="sm" onClick={onClose}>
            Cancel
          </Button>
        )}
        <Button
          variant="primary"
          size="sm"
          onClick={handleSubmit}
          isLoading={isSubmitting}
        >
          <Send className="w-4 h-4" />
          Submit Statements
        </Button>
      </div>
    </div>
  );
}
