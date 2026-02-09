'use client';

import { useState } from 'react';
import {
  ChevronDown,
  ChevronRight,
  Folder,
  FolderOpen,
  FileText,
  Image,
  Video,
  Music,
  FileJson,
  Box,
  File,
  Upload,
  MessageSquare,
} from 'lucide-react';
import { useStore } from '@/lib/store';
import { cn } from '@/lib/utils';
import { DropZone } from '@/components/evidence/DropZone';
import { WitnessForm } from '@/components/evidence/WitnessForm';
import { CommitTimeline } from '@/components/timeline/CommitTimeline';
import {
  EvidenceEmptyState,
  ProfileEmptyState,
  ReasoningEmptyState,
} from '@/components/ui/LoadingStates';
import type { UploadProgress } from '@/hooks/useUpload';
import type { EvidenceFolder, EvidenceItem } from '@/lib/types';

const iconMap: Record<string, React.ComponentType<{ className?: string }>> = {
  FileText,
  Image,
  Video,
  Music,
  FileJson,
  Box,
  File,
};

function getFileIcon(type: string) {
  const map: Record<string, React.ComponentType<{ className?: string }>> = {
    pdf: FileText,
    image: Image,
    video: Video,
    audio: Music,
    json: FileJson,
    '3d': Box,
    text: FileText,
  };
  return map[type] || File;
}

interface FolderItemProps {
  folder: EvidenceFolder;
  onToggle: () => void;
}

function FolderItem({ folder, onToggle }: FolderItemProps) {
  const Icon = folder.isOpen ? FolderOpen : Folder;

  return (
    <div className="select-none">
      <button
        onClick={onToggle}
        className={cn(
          'w-full flex items-center gap-2 px-3 py-2 text-sm',
          'hover:bg-[#1f1f24] rounded-lg transition-colors',
          'text-[#a0a0a8] hover:text-[#f0f0f2]'
        )}
      >
        {folder.isOpen ? (
          <ChevronDown className="w-4 h-4 shrink-0" />
        ) : (
          <ChevronRight className="w-4 h-4 shrink-0" />
        )}
        <Icon className="w-4 h-4 shrink-0 text-[#f59e0b]" />
        <span className="truncate">{folder.name}</span>
        <span className="ml-auto text-xs text-[#606068]">
          {folder.items.length}
        </span>
      </button>

      {folder.isOpen && folder.items.length > 0 && (
        <div className="ml-4 mt-1 space-y-0.5">
          {folder.items.map((item) => (
            <EvidenceItemRow key={item.id} item={item} />
          ))}
        </div>
      )}
    </div>
  );
}

function EvidenceItemRow({ item }: { item: EvidenceItem }) {
  const FileIcon = getFileIcon(item.type);

  return (
    <button
      className={cn(
        'w-full flex items-center gap-2 px-3 py-1.5 text-sm',
        'hover:bg-[#1f1f24] rounded-lg transition-colors',
        'text-[#a0a0a8] hover:text-[#f0f0f2]',
        'evidence-item'
      )}
      title={item.name}
    >
      <FileIcon className="w-4 h-4 shrink-0 text-[#606068]" />
      <span className="truncate text-left">{item.name}</span>
    </button>
  );
}


interface SidebarProps {
  caseId?: string;
  onUpload?: (files: FileList | File[]) => Promise<void>;
  uploadProgress?: UploadProgress[];
  isUploading?: boolean;
}

export function Sidebar({ caseId, onUpload, uploadProgress = [], isUploading = false }: SidebarProps) {
  const { evidenceFolders, toggleFolder, sidebarWidth, commits, viewMode, activeSidebarTab } = useStore();
  const [showWitnessForm, setShowWitnessForm] = useState(false);

  // Demo folders if none provided
  const folders: EvidenceFolder[] =
    evidenceFolders.length > 0
      ? evidenceFolders
      : [
        {
          id: '1',
          name: 'Environment',
          icon: 'Folder',
          isOpen: true,
          items: [
            { id: 'e1', name: 'Blueprint_North_Wing_Gallery.pdf', type: 'pdf' },
            { id: 'e2', name: '32-North-Lidar-Pointcloud.e57', type: '3d' },
            { id: 'e3', name: 'CAD_Layout_Static_Display_Cases.pdf', type: 'pdf' },
            { id: 'e4', name: 'Vault_Construction_Specifications.pdf', type: 'pdf' },
          ],
        },
        {
          id: '2',
          name: 'Ground Truth',
          icon: 'Folder',
          isOpen: true,
          items: [
            { id: 'g1', name: 'CCTV-CAM-04_Vault_Entry_2215.mp4', type: 'video' },
            { id: 'g2', name: 'Security_Hallway_Sync_2210.mp4', type: 'video' },
            { id: 'g3', name: 'Acoustic_Trigger_Glass_Break.wav', type: 'audio' },
            { id: 'g4', name: 'Squad-Car_Dashcam_Exterior.mp4', type: 'video' },
          ],
        },
        {
          id: '3',
          name: 'Electronic Logs',
          icon: 'Folder',
          isOpen: false,
          items: [
            { id: 'l1', name: 'Vault_SmartLock_Audit_Feb01.csv', type: 'json' },
            { id: 'l2', name: 'Motion_Sensor_Grid_Activity.json', type: 'json' },
            { id: 'l3', name: 'Guest_WiFi_Access_Pings_2201.json', type: 'json' },
            { id: 'l4', name: 'RFID_Badge_Swipe_Security.json', type: 'json' },
          ],
        },
        {
          id: '4',
          name: 'Testimonials',
          icon: 'Folder',
          isOpen: false,
          items: [
            { id: 't1', name: 'Witness_A_Security_Guard.txt', type: 'text' },
            { id: 't2', name: 'Witness_B_Late_Night_Visitor.txt', type: 'text' },
            { id: 't3', name: 'Suspect_Alibi_Statement_Kane.txt', type: 'text' },
            { id: 't4', name: 'Initial_Patrol_Observation.txt', type: 'text' },
          ],
        },
      ];

  return (
    <aside
      className="h-full bg-[#09090B] border-r border-[#1e1e24] flex flex-col"
      style={{ width: sidebarWidth }}
    >
      {/* Navigation Icons */}
      {/* Content based on active nav */}
      <div className="flex-1 overflow-y-auto p-3">
        {activeSidebarTab === 'evidence' && (
          <>
            <div className="flex items-center justify-between mb-3">
              <h2 className="text-xs font-medium text-[#606068] uppercase tracking-wider">
                Evidence Archive
              </h2>
              <button
                className="p-1 hover:bg-[#1f1f24] rounded transition-colors text-[#606068] hover:text-[#a0a0a8]"
                title="Upload Evidence"
              >
                <Upload className="w-4 h-4" />
              </button>
            </div>

            <div className="space-y-1">
              {folders.map((folder) => (
                <FolderItem
                  key={folder.id}
                  folder={folder}
                  onToggle={() => toggleFolder(folder.id)}
                />
              ))}
            </div>
          </>
        )}

        {activeSidebarTab === 'witness' && caseId && (
          <WitnessForm
            caseId={caseId}
            onSubmit={(result) => {
              console.log('Witness statement submitted:', result);
            }}
          />
        )}

        {activeSidebarTab === 'home' && (
          <CommitTimeline
            commits={commits}
            onCommitSelect={(commit) => console.log('Selected commit:', commit.id)}
          />
        )}

        {activeSidebarTab === 'suspects' && (
          <ProfileEmptyState />
        )}

        {activeSidebarTab === 'reasoning' && (
          <ReasoningEmptyState />
        )}

        {activeSidebarTab === 'settings' && (
          <div className="text-sm text-[#606068]">
            <p>Case settings and configuration.</p>
          </div>
        )}
      </div>

      {/* Upload zone */}
      <div className="p-3 border-t border-[#1e1e24]">
        {onUpload ? (
          <DropZone
            onFilesDropped={onUpload}
            progress={uploadProgress}
            isUploading={isUploading}
            disabled={!caseId}
          />
        ) : (
          <div
            className={cn(
              'border-2 border-dashed border-[#2a2a32] rounded-lg p-4',
              'flex flex-col items-center justify-center gap-2',
              'text-[#606068] hover:text-[#a0a0a8] hover:border-[#3b82f6]/50',
              'transition-all cursor-pointer'
            )}
          >
            <Upload className="w-5 h-5" />
            <span className="text-xs">Drop files here</span>
          </div>
        )}
      </div>
    </aside>
  );
}
