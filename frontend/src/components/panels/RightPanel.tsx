'use client';

import { useState } from 'react';
import Image from 'next/image';
import {
  User,
  FileText,
  Brain,
  AlertTriangle,
  ChevronDown,
  ChevronUp,
  MapPin,
  RefreshCw,
} from 'lucide-react';
import { cn, formatConfidence } from '@/lib/utils';
import { useStore } from '@/lib/store';
import { getAssetUrl } from '@/lib/api';
import {
  ProfileEmptyState,
  EvidenceEmptyState,
  TrajectoriesEmptyState,
} from '@/components/ui/LoadingStates';
import type { SuspectProfile, Trajectory, EvidenceCard } from '@/lib/types';

type PanelTab = 'suspect' | 'evidence' | 'reasoning';

interface RightPanelProps {
  className?: string;
}

export function RightPanel({ className }: RightPanelProps) {
  const [activeTab, setActiveTab] = useState<PanelTab>('suspect');
  const { suspectProfile, trajectories, sceneGraph } = useStore();

  const tabs = [
    { id: 'suspect' as const, label: 'Suspect', icon: User },
    { id: 'evidence' as const, label: 'Evidence', icon: FileText },
    { id: 'reasoning' as const, label: 'Reasoning', icon: Brain },
  ];

  return (
    <aside className={cn('w-72 bg-[#111114] border-l border-[#1e1e24] flex flex-col', className)}>
      {/* Tab buttons */}
      <div className="flex border-b border-[#1e1e24]">
        {tabs.map((tab) => {
          const Icon = tab.icon;
          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={cn(
                'flex-1 flex items-center justify-center gap-2 py-3 text-xs font-medium transition-colors',
                activeTab === tab.id
                  ? 'text-[#f0f0f2] border-b-2 border-[#3b82f6] bg-[#1f1f24]/50'
                  : 'text-[#606068] hover:text-[#a0a0a8]'
              )}
            >
              <Icon className="w-4 h-4" />
              {tab.label}
            </button>
          );
        })}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-y-auto p-4">
        {activeTab === 'suspect' && <SuspectTab profile={suspectProfile} />}
        {activeTab === 'evidence' && <EvidenceTab evidence={sceneGraph?.evidence || []} />}
        {activeTab === 'reasoning' && <ReasoningTab trajectories={trajectories} />}
      </div>
    </aside>
  );
}

function SuspectTab({ profile }: { profile: SuspectProfile | null }) {
  const [imageError, setImageError] = useState(false);
  const [imageLoading, setImageLoading] = useState(true);

  // Show empty state if no profile
  if (!profile) {
    return <ProfileEmptyState />;
  }

  // Use actual profile attributes
  const attrs = profile.attributes || {
    age_range: { min: 25, max: 35, confidence: 0.7 },
    height_range_cm: { min: 170, max: 180, confidence: 0.8 },
    build: { value: 'average', confidence: 0.6 },
    hair: { style: 'short', color: 'dark', confidence: 0.75 },
    distinctive_features: [
      { description: 'Walks with slight limp', confidence: 0.65 },
      { description: 'Wears glasses', confidence: 0.55 },
    ],
  };

  // Get portrait URL if available
  const portraitUrl = profile?.portrait_asset_key
    ? getAssetUrl(profile.portrait_asset_key)
    : null;

  return (
    <div className="space-y-4">
      {/* Portrait */}
      <div className="aspect-square bg-[#1f1f24] rounded-lg flex items-center justify-center overflow-hidden relative">
        {portraitUrl && !imageError ? (
          <>
            {imageLoading && (
              <div className="absolute inset-0 flex items-center justify-center bg-[#1f1f24]">
                <RefreshCw className="w-8 h-8 text-[#3b82f6] animate-spin" />
              </div>
            )}
            <Image
              src={portraitUrl}
              alt="Suspect Portrait"
              fill
              className={cn(
                'object-cover rounded-lg transition-opacity duration-300',
                imageLoading ? 'opacity-0' : 'opacity-100'
              )}
              onLoad={() => setImageLoading(false)}
              onError={() => {
                setImageError(true);
                setImageLoading(false);
              }}
              unoptimized // For external URLs
            />
            {/* Generated badge */}
            <div className="absolute bottom-2 right-2 px-2 py-1 rounded bg-[#8b5cf6]/80 text-xs text-white">
              AI Generated
            </div>
          </>
        ) : (
          <div className="flex flex-col items-center gap-2">
            <User className="w-16 h-16 text-[#2a2a32]" />
            <span className="text-xs text-[#606068]">
              {profile ? 'Portrait generating...' : 'No portrait yet'}
            </span>
          </div>
        )}
      </div>

      {/* Attributes */}
      <div className="space-y-3">
        <h4 className="text-xs font-medium text-[#606068] uppercase tracking-wider">
          Attributes
        </h4>

        <AttributeRow
          label="Age"
          value={`${attrs.age_range?.min}-${attrs.age_range?.max} years`}
          confidence={attrs.age_range?.confidence || 0}
        />
        <AttributeRow
          label="Height"
          value={`${attrs.height_range_cm?.min}-${attrs.height_range_cm?.max} cm`}
          confidence={attrs.height_range_cm?.confidence || 0}
        />
        <AttributeRow
          label="Build"
          value={attrs.build?.value || 'Unknown'}
          confidence={attrs.build?.confidence || 0}
        />
        <AttributeRow
          label="Hair"
          value={`${attrs.hair?.color || ''} ${attrs.hair?.style || ''}`.trim() || 'Unknown'}
          confidence={attrs.hair?.confidence || 0}
        />

        {/* Distinctive features */}
        {attrs.distinctive_features && attrs.distinctive_features.length > 0 && (
          <div className="pt-2">
            <h4 className="text-xs font-medium text-[#606068] uppercase tracking-wider mb-2">
              Distinctive Features
            </h4>
            <div className="space-y-2">
              {attrs.distinctive_features.map((feature, i) => (
                <div
                  key={i}
                  className="flex items-start gap-2 text-sm text-[#a0a0a8] bg-[#1f1f24] rounded p-2"
                >
                  <AlertTriangle className="w-4 h-4 text-[#f59e0b] shrink-0 mt-0.5" />
                  <div>
                    <p>{feature.description}</p>
                    <p className="text-xs text-[#606068] mt-0.5">
                      {formatConfidence(feature.confidence)} confidence
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

function AttributeRow({
  label,
  value,
  confidence,
}: {
  label: string;
  value: string;
  confidence: number;
}) {
  const confidenceColor =
    confidence >= 0.7 ? '#22c55e' : confidence >= 0.5 ? '#f59e0b' : '#ef4444';

  return (
    <div className="flex items-center justify-between">
      <span className="text-sm text-[#606068]">{label}</span>
      <div className="flex items-center gap-2">
        <span className="text-sm text-[#f0f0f2]">{value}</span>
        <div
          className="w-2 h-2 rounded-full"
          style={{ backgroundColor: confidenceColor }}
          title={`${Math.round(confidence * 100)}% confidence`}
        />
      </div>
    </div>
  );
}

function EvidenceTab({ evidence }: { evidence: EvidenceCard[] }) {
  // Show empty state if no evidence
  if (evidence.length === 0) {
    return <EvidenceEmptyState />;
  }

  return (
    <div className="space-y-3">
      <h4 className="text-xs font-medium text-[#606068] uppercase tracking-wider">
        Evidence Cards ({evidence.length})
      </h4>

      {evidence.map((item) => (
        <EvidenceCardItem key={item.id} evidence={item} />
      ))}
    </div>
  );
}

function EvidenceCardItem({ evidence }: { evidence: EvidenceCard }) {
  const [isExpanded, setIsExpanded] = useState(false);
  const confidenceColor =
    evidence.confidence >= 0.8
      ? '#22c55e'
      : evidence.confidence >= 0.5
      ? '#f59e0b'
      : '#ef4444';

  return (
    <div className="bg-[#1f1f24] rounded-lg overflow-hidden">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-start gap-3 p-3 text-left hover:bg-[#2a2a32]/50 transition-colors"
      >
        <div
          className="w-1 h-full min-h-[40px] rounded-full"
          style={{ backgroundColor: confidenceColor }}
        />
        <div className="flex-1 min-w-0">
          <p className="text-sm text-[#f0f0f2] font-medium">{evidence.title}</p>
          <p className="text-xs text-[#606068] mt-0.5">
            {formatConfidence(evidence.confidence)} confidence
          </p>
        </div>
        {isExpanded ? (
          <ChevronUp className="w-4 h-4 text-[#606068]" />
        ) : (
          <ChevronDown className="w-4 h-4 text-[#606068]" />
        )}
      </button>

      {isExpanded && (
        <div className="px-3 pb-3 pt-0 text-sm text-[#a0a0a8] break-words">{evidence.description}</div>
      )}
    </div>
  );
}

function ReasoningTab({ trajectories }: { trajectories: Trajectory[] }) {
  // Show empty state if no trajectories
  if (trajectories.length === 0) {
    return <TrajectoriesEmptyState />;
  }

  return (
    <div className="space-y-4">
      {/* Trajectories section */}
      <div>
        <h4 className="text-xs font-medium text-[#606068] uppercase tracking-wider mb-2">
          Trajectory Hypotheses ({trajectories.length})
        </h4>

        {trajectories.map((traj) => (
          <div key={traj.id} className="bg-[#1f1f24] rounded-lg p-3 space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-[#f0f0f2]">Hypothesis #{traj.rank}</span>
              <span className="text-xs px-2 py-0.5 rounded bg-[#22c55e]/20 text-[#22c55e]">
                {formatConfidence(traj.overall_confidence)}
              </span>
            </div>

            <div className="space-y-1">
              {traj.segments.map((seg) => (
                <div key={seg.id} className="flex items-center gap-2 text-xs text-[#a0a0a8]">
                  <MapPin className="w-3 h-3 text-[#3b82f6]" />
                  <span>{seg.explanation}</span>
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
