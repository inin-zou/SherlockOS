import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import type {
  Case,
  Commit,
  Job,
  SceneGraph,
  Trajectory,
  EvidenceFolder,
  TimelineTrack,
  SceneAnnotation,
  SuspectProfile,
} from './types';

// View modes
export type ViewMode = 'evidence' | 'simulation' | 'reasoning';
export type SidebarTab = 'home' | 'evidence' | 'witness' | 'suspects' | 'reasoning' | 'settings';

interface AppState {
  // Current case
  currentCase: Case | null;
  setCurrentCase: (caseData: Case | null) => void;

  // Cases list
  cases: Case[];
  setCases: (cases: Case[]) => void;
  addCase: (caseData: Case) => void;
  removeCase: (caseId: string) => void;

  // Scene graph
  sceneGraph: SceneGraph | null;
  setSceneGraph: (sg: SceneGraph | null) => void;

  // Timeline commits
  commits: Commit[];
  setCommits: (commits: Commit[]) => void;
  addCommit: (commit: Commit) => void;

  // Jobs
  jobs: Job[];
  setJobs: (jobs: Job[]) => void;
  updateJob: (jobId: string, updates: Partial<Job>) => void;

  // Trajectories (from reasoning)
  trajectories: Trajectory[];
  setTrajectories: (trajectories: Trajectory[]) => void;
  selectedTrajectoryId: string | null;
  setSelectedTrajectoryId: (id: string | null) => void;

  // Suspect profile
  suspectProfile: SuspectProfile | null;
  setSuspectProfile: (profile: SuspectProfile | null) => void;

  // View mode
  viewMode: ViewMode;
  setViewMode: (mode: ViewMode) => void;

  // Selected objects
  selectedObjectIds: string[];
  setSelectedObjectIds: (ids: string[]) => void;
  toggleObjectSelection: (id: string) => void;

  // Evidence sidebar
  activeSidebarTab: SidebarTab;
  setActiveSidebarTab: (tab: SidebarTab) => void;
  evidenceFolders: EvidenceFolder[];
  setEvidenceFolders: (folders: EvidenceFolder[]) => void;
  toggleFolder: (folderId: string) => void;

  // Timeline
  timelineTracks: TimelineTrack[];
  setTimelineTracks: (tracks: TimelineTrack[]) => void;
  currentTime: number;
  setCurrentTime: (time: number) => void;
  isPlaying: boolean;
  setIsPlaying: (playing: boolean) => void;

  // 3D Scene
  annotations: SceneAnnotation[];
  setAnnotations: (annotations: SceneAnnotation[]) => void;
  cameraPosition: [number, number, number];
  setCameraPosition: (pos: [number, number, number]) => void;

  // UI state
  sidebarWidth: number;
  setSidebarWidth: (width: number) => void;
  timelineHeight: number;
  setTimelineHeight: (height: number) => void;
  isLoading: boolean;
  setIsLoading: (loading: boolean) => void;

  // Reset
  reset: () => void;
}

const initialState = {
  currentCase: null,
  cases: [],
  sceneGraph: null,
  commits: [],
  jobs: [],
  trajectories: [],
  selectedTrajectoryId: null,
  suspectProfile: null,
  viewMode: 'evidence' as ViewMode,
  selectedObjectIds: [],
  activeSidebarTab: 'evidence' as SidebarTab,
  evidenceFolders: [],
  timelineTracks: [],
  currentTime: 0,
  isPlaying: false,
  annotations: [],
  cameraPosition: [5, 5, 5] as [number, number, number],
  sidebarWidth: 280,
  timelineHeight: 200,
  isLoading: false,
};

export const useStore = create<AppState>()(
  devtools(
    (set) => ({
      ...initialState,

      setCurrentCase: (caseData) => set({ currentCase: caseData }),

      setCases: (cases) => set({ cases }),
      addCase: (caseData) =>
        set((state) => ({ cases: [...state.cases, caseData] })),
      removeCase: (caseId) =>
        set((state) => {
          const newCases = state.cases.filter((c) => c.id !== caseId);
          // If we're removing the current case, switch to another one
          const newCurrentCase =
            state.currentCase?.id === caseId
              ? newCases[0] || null
              : state.currentCase;
          return { cases: newCases, currentCase: newCurrentCase };
        }),

      setSceneGraph: (sg) => set({ sceneGraph: sg }),

      setCommits: (commits) => set({ commits }),
      addCommit: (commit) =>
        set((state) => ({ commits: [commit, ...state.commits] })),

      setJobs: (jobs) => set({ jobs }),
      updateJob: (jobId, updates) =>
        set((state) => ({
          jobs: state.jobs.map((job) =>
            job.id === jobId ? { ...job, ...updates } : job
          ),
        })),

      setTrajectories: (trajectories) => set({ trajectories }),
      setSelectedTrajectoryId: (id) => set({ selectedTrajectoryId: id }),

      setSuspectProfile: (profile) => set({ suspectProfile: profile }),

      setViewMode: (mode) => set({ viewMode: mode }),

      setSelectedObjectIds: (ids) => set({ selectedObjectIds: ids }),
      toggleObjectSelection: (id) =>
        set((state) => ({
          selectedObjectIds: state.selectedObjectIds.includes(id)
            ? state.selectedObjectIds.filter((oid) => oid !== id)
            : [...state.selectedObjectIds, id],
        })),

      setActiveSidebarTab: (tab) => set({ activeSidebarTab: tab }),

      setEvidenceFolders: (folders) => set({ evidenceFolders: folders }),
      toggleFolder: (folderId) =>
        set((state) => ({
          evidenceFolders: state.evidenceFolders.map((folder) =>
            folder.id === folderId
              ? { ...folder, isOpen: !folder.isOpen }
              : folder
          ),
        })),

      setTimelineTracks: (tracks) => set({ timelineTracks: tracks }),
      setCurrentTime: (time) => set({ currentTime: time }),
      setIsPlaying: (playing) => set({ isPlaying: playing }),

      setAnnotations: (annotations) => set({ annotations }),
      setCameraPosition: (pos) => set({ cameraPosition: pos }),

      setSidebarWidth: (width) => set({ sidebarWidth: width }),
      setTimelineHeight: (height) => set({ timelineHeight: height }),
      setIsLoading: (loading) => set({ isLoading: loading }),

      reset: () => set(initialState),
    }),
    { name: 'SherlockOS' }
  )
);
