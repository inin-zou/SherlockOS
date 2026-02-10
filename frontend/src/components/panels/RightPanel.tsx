'use client';

import { useState } from 'react';
import Image from 'next/image';
import {
  User,
  FileText,
  Brain,
  ChevronDown,
  ChevronUp,
  MapPin,
  RefreshCw,
  Clock,
  Sparkles,
} from 'lucide-react';
import Link from 'next/link';
import { cn, formatConfidence } from '@/lib/utils';
import { useStore } from '@/lib/store';
import { getAssetUrl } from '@/lib/api';
import {
  ProfileEmptyState,
  EvidenceEmptyState,
  TrajectoriesEmptyState,
} from '@/components/ui/LoadingStates';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import type { SuspectProfile, Trajectory, EvidenceCard } from '@/lib/types';

interface RightPanelProps {
  className?: string;
}

export function RightPanel({ className }: RightPanelProps) {
  const { suspectProfile, trajectories, sceneGraph } = useStore();

  return (
    <aside className={cn('w-80 bg-card border-l border-border flex flex-col', className)}>
      <Tabs defaultValue="suspect" className="flex flex-col h-full gap-0">
        <TabsList variant="line" className="w-full justify-center border-b border-border rounded-none px-2 shrink-0">
          <TabsTrigger value="suspect" className="gap-1.5 text-[13px]">
            <User className="w-3.5 h-3.5" />
            Suspect
          </TabsTrigger>
          <TabsTrigger value="evidence" className="gap-1.5 text-[13px]">
            <FileText className="w-3.5 h-3.5" />
            Evidence
          </TabsTrigger>
          <TabsTrigger value="reasoning" className="gap-1.5 text-[13px]">
            <Brain className="w-3.5 h-3.5" />
            Reasoning
          </TabsTrigger>
        </TabsList>

        <div className="flex-1 overflow-y-auto">
          <TabsContent value="suspect" className="px-5 py-6 m-0">
            <SuspectTab profile={suspectProfile} />
          </TabsContent>
          <TabsContent value="evidence" className="px-5 py-6 m-0">
            <EvidenceTab evidence={sceneGraph?.evidence || []} />
          </TabsContent>
          <TabsContent value="reasoning" className="px-5 py-6 m-0">
            <ReasoningTab trajectories={trajectories} />
          </TabsContent>
        </div>
      </Tabs>
    </aside>
  );
}

// Demo stakeholder data for default state
const demoStakeholder = {
  name: 'Nova Welsh',
  role: 'Victim',
  initials: 'NW',
  overview: 'Nova Welsh, 34, was the last known person in the North Wing Gallery before the incident. She is a registered art appraiser who was conducting an after-hours evaluation of the Meridian Collection.',
  details: [
    { label: 'Relationship', value: 'Art Appraiser (Contracted)' },
    { label: 'Age', value: '34' },
    { label: 'Role', value: 'Victim / Key Witness' },
  ],
  chainOfAction: [
    { time: '8:00 PM', action: 'Arrived at gallery via main entrance' },
    { time: '8:15 PM', action: 'Signed in at security desk — Badge #NW-0412' },
    { time: '8:30 PM', action: 'Entered North Wing Gallery alone' },
    { time: '9:15 PM', action: 'Last seen on CCTV-CAM-04 near vault corridor' },
    { time: '9:22 PM', action: 'Glass break acoustic trigger activated' },
    { time: '9:45 PM', action: 'Found by security patrol — vault door ajar' },
  ],
  comments: [
    { author: 'Det. Reyes', initials: 'DR', time: '10:30 PM', text: 'Witness appeared shaken but cooperative. Consistent timeline with CCTV.' },
    { author: 'Forensics', initials: 'FO', time: '11:15 PM', text: 'No fingerprints found on vault lock — gloves suspected.' },
  ],
};

function SuspectTab({ profile }: { profile: SuspectProfile | null }) {
  const [imageError, setImageError] = useState(false);
  const [imageLoading, setImageLoading] = useState(true);

  const portraitUrl = profile?.portrait_asset_key
    ? getAssetUrl(profile.portrait_asset_key)
    : null;

  const stakeholder = demoStakeholder;

  return (
    <div className="space-y-8">
      {/* Profile Header */}
      <div className="flex flex-col items-center gap-4">
        <Avatar className="w-[72px] h-[72px] ring-2 ring-border ring-offset-2 ring-offset-card">
          {portraitUrl && !imageError ? (
            <AvatarImage
              src={portraitUrl}
              alt={stakeholder.name}
              onLoadingStatusChange={(status) => {
                if (status === 'loaded') setImageLoading(false);
                if (status === 'error') { setImageError(true); setImageLoading(false); }
              }}
            />
          ) : null}
          <AvatarFallback className="text-lg bg-secondary text-secondary-foreground">
            {stakeholder.initials}
          </AvatarFallback>
        </Avatar>

        <div className="text-center space-y-1.5">
          <h3 className="text-[15px] font-semibold text-foreground leading-tight">{stakeholder.name}</h3>
          <Badge variant="secondary" className="text-[11px]">{stakeholder.role}</Badge>
        </div>

        <Link
          href="/portrait"
          className="flex items-center gap-2 px-4 py-2.5 rounded-lg bg-primary/10 border border-primary/20 text-primary text-[13px] font-medium hover:bg-primary/20 transition-colors"
        >
          <Sparkles className="w-3.5 h-3.5" />
          Generate Suspect Portrait
        </Link>
      </div>

      <Separator />

      {/* Overview */}
      <section>
        <h4 className="text-[11px] font-semibold text-muted-foreground uppercase tracking-[0.08em] mb-3">
          Overview
        </h4>
        <p className="text-[13px] text-secondary-foreground leading-[1.7]">
          {stakeholder.overview}
        </p>
      </section>

      {/* Key Details */}
      <section>
        <h4 className="text-[11px] font-semibold text-muted-foreground uppercase tracking-[0.08em] mb-3">
          Details
        </h4>
        <Card className="py-0 gap-0 shadow-none">
          {stakeholder.details.map((detail, i) => (
            <div key={detail.label}>
              <div className="flex items-center justify-between px-4 py-3">
                <span className="text-[13px] text-muted-foreground">{detail.label}</span>
                <span className="text-[13px] text-foreground font-medium text-right">{detail.value}</span>
              </div>
              {i < stakeholder.details.length - 1 && <Separator />}
            </div>
          ))}
        </Card>
      </section>

      {/* Chain of Action */}
      <section>
        <h4 className="text-[11px] font-semibold text-muted-foreground uppercase tracking-[0.08em] mb-5">
          Chain of Action
        </h4>
        <div className="relative ml-3 pl-6 border-l border-border">
          <div className="space-y-6">
            {stakeholder.chainOfAction.map((item, i) => (
              <div key={i} className="relative">
                {/* Timeline dot */}
                <div className="absolute -left-[27px] top-[3px] w-2.5 h-2.5 rounded-full bg-primary ring-[3px] ring-card" />
                <div className="space-y-1">
                  <span className="text-[11px] text-muted-foreground font-mono tracking-wide flex items-center gap-1.5">
                    <Clock className="w-3 h-3" />
                    {item.time}
                  </span>
                  <p className="text-[13px] text-secondary-foreground leading-relaxed">{item.action}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Follow-up Comments */}
      <section>
        <h4 className="text-[11px] font-semibold text-muted-foreground uppercase tracking-[0.08em] mb-4">
          Follow-up Comments
        </h4>
        <div className="space-y-3">
          {stakeholder.comments.map((comment, i) => (
            <Card key={i} className="py-0 gap-0 shadow-none">
              <CardContent className="p-4 space-y-2.5">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2.5">
                    <Avatar className="w-6 h-6">
                      <AvatarFallback className="text-[10px] bg-secondary text-secondary-foreground">
                        {comment.initials}
                      </AvatarFallback>
                    </Avatar>
                    <span className="text-[13px] font-semibold text-foreground">{comment.author}</span>
                  </div>
                  <span className="text-[11px] text-muted-foreground font-mono">{comment.time}</span>
                </div>
                <p className="text-[13px] text-secondary-foreground leading-[1.65]">{comment.text}</p>
              </CardContent>
            </Card>
          ))}
        </div>
      </section>
    </div>
  );
}

function EvidenceTab({ evidence }: { evidence: EvidenceCard[] }) {
  if (evidence.length === 0) {
    return <EvidenceEmptyState />;
  }

  return (
    <div className="space-y-4">
      <h4 className="text-[11px] font-semibold text-muted-foreground uppercase tracking-[0.08em]">
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
      ? 'bg-emerald-500'
      : evidence.confidence >= 0.5
      ? 'bg-amber-500'
      : 'bg-red-500';

  return (
    <Card className="py-0 gap-0 shadow-none overflow-hidden">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full flex items-start gap-3 p-4 text-left hover:bg-accent/50 transition-colors"
      >
        <div className={cn('w-1 self-stretch rounded-full shrink-0', confidenceColor)} />
        <div className="flex-1 min-w-0 space-y-1">
          <p className="text-[13px] text-foreground font-medium">{evidence.title}</p>
          <p className="text-[11px] text-muted-foreground">
            {formatConfidence(evidence.confidence)} confidence
          </p>
        </div>
        {isExpanded ? (
          <ChevronUp className="w-4 h-4 text-muted-foreground shrink-0" />
        ) : (
          <ChevronDown className="w-4 h-4 text-muted-foreground shrink-0" />
        )}
      </button>
      {isExpanded && (
        <>
          <Separator />
          <CardContent className="p-4 pt-3">
            <p className="text-[13px] text-secondary-foreground leading-relaxed break-words">{evidence.description}</p>
          </CardContent>
        </>
      )}
    </Card>
  );
}

function ReasoningTab({ trajectories }: { trajectories: Trajectory[] }) {
  if (trajectories.length === 0) {
    return <TrajectoriesEmptyState />;
  }

  return (
    <div className="space-y-4">
      <h4 className="text-[11px] font-semibold text-muted-foreground uppercase tracking-[0.08em]">
        Trajectory Hypotheses ({trajectories.length})
      </h4>

      {trajectories.map((traj) => (
        <Card key={traj.id} className="py-0 gap-0 shadow-none">
          <CardContent className="p-4 space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-[13px] font-semibold text-foreground">Hypothesis #{traj.rank}</span>
              <Badge variant="secondary" className="text-[11px] bg-emerald-500/10 text-emerald-400 border-emerald-500/20">
                {formatConfidence(traj.overall_confidence)}
              </Badge>
            </div>
            <div className="space-y-2">
              {traj.segments.map((seg) => (
                <div key={seg.id} className="flex items-start gap-2.5 text-[13px] text-secondary-foreground">
                  <MapPin className="w-3.5 h-3.5 text-primary shrink-0 mt-0.5" />
                  <span className="leading-snug">{seg.explanation}</span>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
