import { describe, it, expect, beforeEach } from 'vitest';
import { useStore } from './store';
import { act } from '@testing-library/react';
import type { Case, Commit, Job, Trajectory, EvidenceFolder, SuspectProfile } from './types';

describe('Store', () => {
  beforeEach(() => {
    // Reset store to initial state before each test
    act(() => {
      useStore.getState().reset();
    });
  });

  describe('Initial State', () => {
    it('has correct initial values', () => {
      const state = useStore.getState();

      expect(state.currentCase).toBeNull();
      expect(state.cases).toEqual([]);
      expect(state.sceneGraph).toBeNull();
      expect(state.commits).toEqual([]);
      expect(state.jobs).toEqual([]);
      expect(state.trajectories).toEqual([]);
      expect(state.selectedTrajectoryId).toBeNull();
      expect(state.suspectProfile).toBeNull();
      expect(state.viewMode).toBe('evidence');
      expect(state.selectedObjectIds).toEqual([]);
      expect(state.evidenceFolders).toEqual([]);
      expect(state.currentTime).toBe(0);
      expect(state.isPlaying).toBe(false);
      expect(state.sidebarWidth).toBe(280);
      expect(state.timelineHeight).toBe(200);
      expect(state.isLoading).toBe(false);
      expect(state.cameraPosition).toEqual([5, 5, 5]);
    });
  });

  describe('Case Management', () => {
    it('sets current case', () => {
      const caseData: Case = {
        id: 'case-1',
        title: 'Test Case',
        description: 'A test',
        created_at: '2024-01-01',
      };

      act(() => {
        useStore.getState().setCurrentCase(caseData);
      });

      expect(useStore.getState().currentCase).toEqual(caseData);
    });

    it('clears current case', () => {
      act(() => {
        useStore.getState().setCurrentCase({ id: '1', title: 'Test', created_at: '2024-01-01' });
        useStore.getState().setCurrentCase(null);
      });

      expect(useStore.getState().currentCase).toBeNull();
    });

    it('sets cases list', () => {
      const cases: Case[] = [
        { id: '1', title: 'Case 1', created_at: '2024-01-01' },
        { id: '2', title: 'Case 2', created_at: '2024-01-02' },
      ];

      act(() => {
        useStore.getState().setCases(cases);
      });

      expect(useStore.getState().cases).toEqual(cases);
    });

    it('adds a case', () => {
      const existingCase: Case = { id: '1', title: 'Existing', created_at: '2024-01-01' };
      const newCase: Case = { id: '2', title: 'New', created_at: '2024-01-02' };

      act(() => {
        useStore.getState().setCases([existingCase]);
        useStore.getState().addCase(newCase);
      });

      expect(useStore.getState().cases).toHaveLength(2);
      expect(useStore.getState().cases[1]).toEqual(newCase);
    });
  });

  describe('SceneGraph', () => {
    it('sets scene graph', () => {
      const sceneGraph = {
        objects: [{ id: 'obj-1', type: 'evidence' }],
        relationships: [],
      };

      act(() => {
        useStore.getState().setSceneGraph(sceneGraph as any);
      });

      expect(useStore.getState().sceneGraph).toEqual(sceneGraph);
    });

    it('clears scene graph', () => {
      act(() => {
        useStore.getState().setSceneGraph({ objects: [], relationships: [] } as any);
        useStore.getState().setSceneGraph(null);
      });

      expect(useStore.getState().sceneGraph).toBeNull();
    });
  });

  describe('Commits', () => {
    it('sets commits', () => {
      const commits: Commit[] = [
        { id: 'c1', case_id: 'case-1', type: 'upload_scan', summary: 'Upload', payload: {}, created_at: '2024-01-01' },
        { id: 'c2', case_id: 'case-1', type: 'witness_statement', summary: 'Witness', payload: {}, created_at: '2024-01-02' },
      ];

      act(() => {
        useStore.getState().setCommits(commits);
      });

      expect(useStore.getState().commits).toEqual(commits);
    });

    it('adds commit to beginning of list', () => {
      const existingCommit: Commit = {
        id: 'c1',
        case_id: 'case-1',
        type: 'upload_scan',
        summary: 'Old',
        payload: {},
        created_at: '2024-01-01',
      };
      const newCommit: Commit = {
        id: 'c2',
        case_id: 'case-1',
        type: 'reasoning_result',
        summary: 'New',
        payload: {},
        created_at: '2024-01-02',
      };

      act(() => {
        useStore.getState().setCommits([existingCommit]);
        useStore.getState().addCommit(newCommit);
      });

      const commits = useStore.getState().commits;
      expect(commits).toHaveLength(2);
      expect(commits[0]).toEqual(newCommit); // New commit is first
      expect(commits[1]).toEqual(existingCommit);
    });
  });

  describe('Jobs', () => {
    it('sets jobs', () => {
      const jobs: Job[] = [
        { id: 'job-1', type: 'reconstruction', status: 'queued', progress: 0, created_at: '2024-01-01' },
      ];

      act(() => {
        useStore.getState().setJobs(jobs);
      });

      expect(useStore.getState().jobs).toEqual(jobs);
    });

    it('updates job by ID', () => {
      const jobs: Job[] = [
        { id: 'job-1', type: 'reconstruction', status: 'queued', progress: 0, created_at: '2024-01-01' },
        { id: 'job-2', type: 'reasoning', status: 'queued', progress: 0, created_at: '2024-01-01' },
      ];

      act(() => {
        useStore.getState().setJobs(jobs);
        useStore.getState().updateJob('job-1', { status: 'running', progress: 50 });
      });

      const updatedJobs = useStore.getState().jobs;
      expect(updatedJobs[0].status).toBe('running');
      expect(updatedJobs[0].progress).toBe(50);
      expect(updatedJobs[1].status).toBe('queued'); // Unchanged
    });

    it('does not modify jobs if ID not found', () => {
      const jobs: Job[] = [
        { id: 'job-1', type: 'reconstruction', status: 'queued', progress: 0, created_at: '2024-01-01' },
      ];

      act(() => {
        useStore.getState().setJobs(jobs);
        useStore.getState().updateJob('non-existent', { status: 'running' });
      });

      expect(useStore.getState().jobs[0].status).toBe('queued');
    });
  });

  describe('Trajectories', () => {
    it('sets trajectories', () => {
      const trajectories: Trajectory[] = [
        { id: 't1', confidence: 0.9, segments: [], explanation: 'Test' },
      ];

      act(() => {
        useStore.getState().setTrajectories(trajectories);
      });

      expect(useStore.getState().trajectories).toEqual(trajectories);
    });

    it('sets selected trajectory ID', () => {
      act(() => {
        useStore.getState().setSelectedTrajectoryId('trajectory-1');
      });

      expect(useStore.getState().selectedTrajectoryId).toBe('trajectory-1');
    });

    it('clears selected trajectory', () => {
      act(() => {
        useStore.getState().setSelectedTrajectoryId('t1');
        useStore.getState().setSelectedTrajectoryId(null);
      });

      expect(useStore.getState().selectedTrajectoryId).toBeNull();
    });
  });

  describe('Suspect Profile', () => {
    it('sets suspect profile', () => {
      const profile: SuspectProfile = {
        case_id: 'case-1',
        commit_id: 'commit-1',
        attributes: { age: '25-35', height: '170-180cm' },
        updated_at: '2024-01-01',
      };

      act(() => {
        useStore.getState().setSuspectProfile(profile);
      });

      expect(useStore.getState().suspectProfile).toEqual(profile);
    });

    it('clears suspect profile', () => {
      act(() => {
        useStore.getState().setSuspectProfile({ case_id: '1', commit_id: '1', attributes: {}, updated_at: '' });
        useStore.getState().setSuspectProfile(null);
      });

      expect(useStore.getState().suspectProfile).toBeNull();
    });
  });

  describe('View Mode', () => {
    it('sets view mode to evidence', () => {
      act(() => {
        useStore.getState().setViewMode('evidence');
      });

      expect(useStore.getState().viewMode).toBe('evidence');
    });

    it('sets view mode to simulation', () => {
      act(() => {
        useStore.getState().setViewMode('simulation');
      });

      expect(useStore.getState().viewMode).toBe('simulation');
    });

    it('sets view mode to reasoning', () => {
      act(() => {
        useStore.getState().setViewMode('reasoning');
      });

      expect(useStore.getState().viewMode).toBe('reasoning');
    });
  });

  describe('Object Selection', () => {
    it('sets selected object IDs', () => {
      act(() => {
        useStore.getState().setSelectedObjectIds(['obj-1', 'obj-2']);
      });

      expect(useStore.getState().selectedObjectIds).toEqual(['obj-1', 'obj-2']);
    });

    it('toggles object selection on', () => {
      act(() => {
        useStore.getState().toggleObjectSelection('obj-1');
      });

      expect(useStore.getState().selectedObjectIds).toContain('obj-1');
    });

    it('toggles object selection off', () => {
      act(() => {
        useStore.getState().setSelectedObjectIds(['obj-1', 'obj-2']);
        useStore.getState().toggleObjectSelection('obj-1');
      });

      expect(useStore.getState().selectedObjectIds).not.toContain('obj-1');
      expect(useStore.getState().selectedObjectIds).toContain('obj-2');
    });
  });

  describe('Evidence Folders', () => {
    it('sets evidence folders', () => {
      const folders: EvidenceFolder[] = [
        { id: 'f1', name: 'Folder 1', icon: 'Folder', isOpen: true, items: [] },
      ];

      act(() => {
        useStore.getState().setEvidenceFolders(folders);
      });

      expect(useStore.getState().evidenceFolders).toEqual(folders);
    });

    it('toggles folder open state', () => {
      const folders: EvidenceFolder[] = [
        { id: 'f1', name: 'Folder 1', icon: 'Folder', isOpen: false, items: [] },
        { id: 'f2', name: 'Folder 2', icon: 'Folder', isOpen: true, items: [] },
      ];

      act(() => {
        useStore.getState().setEvidenceFolders(folders);
        useStore.getState().toggleFolder('f1');
      });

      expect(useStore.getState().evidenceFolders[0].isOpen).toBe(true);
      expect(useStore.getState().evidenceFolders[1].isOpen).toBe(true); // Unchanged
    });

    it('toggles folder closed', () => {
      const folders: EvidenceFolder[] = [
        { id: 'f1', name: 'Folder 1', icon: 'Folder', isOpen: true, items: [] },
      ];

      act(() => {
        useStore.getState().setEvidenceFolders(folders);
        useStore.getState().toggleFolder('f1');
      });

      expect(useStore.getState().evidenceFolders[0].isOpen).toBe(false);
    });
  });

  describe('Timeline', () => {
    it('sets current time', () => {
      act(() => {
        useStore.getState().setCurrentTime(1234567890);
      });

      expect(useStore.getState().currentTime).toBe(1234567890);
    });

    it('sets playing state', () => {
      act(() => {
        useStore.getState().setIsPlaying(true);
      });

      expect(useStore.getState().isPlaying).toBe(true);

      act(() => {
        useStore.getState().setIsPlaying(false);
      });

      expect(useStore.getState().isPlaying).toBe(false);
    });

    it('sets timeline tracks', () => {
      const tracks = [
        { id: 't1', name: 'Track 1', events: [] },
      ];

      act(() => {
        useStore.getState().setTimelineTracks(tracks as any);
      });

      expect(useStore.getState().timelineTracks).toEqual(tracks);
    });
  });

  describe('3D Scene', () => {
    it('sets annotations', () => {
      const annotations = [
        { id: 'a1', position: [0, 0, 0], label: 'Evidence 1' },
      ];

      act(() => {
        useStore.getState().setAnnotations(annotations as any);
      });

      expect(useStore.getState().annotations).toEqual(annotations);
    });

    it('sets camera position', () => {
      act(() => {
        useStore.getState().setCameraPosition([10, 20, 30]);
      });

      expect(useStore.getState().cameraPosition).toEqual([10, 20, 30]);
    });
  });

  describe('UI State', () => {
    it('sets sidebar width', () => {
      act(() => {
        useStore.getState().setSidebarWidth(320);
      });

      expect(useStore.getState().sidebarWidth).toBe(320);
    });

    it('sets timeline height', () => {
      act(() => {
        useStore.getState().setTimelineHeight(250);
      });

      expect(useStore.getState().timelineHeight).toBe(250);
    });

    it('sets loading state', () => {
      act(() => {
        useStore.getState().setIsLoading(true);
      });

      expect(useStore.getState().isLoading).toBe(true);

      act(() => {
        useStore.getState().setIsLoading(false);
      });

      expect(useStore.getState().isLoading).toBe(false);
    });
  });

  describe('Reset', () => {
    it('resets all state to initial values', () => {
      // Set various state
      act(() => {
        useStore.getState().setCurrentCase({ id: '1', title: 'Test', created_at: '2024-01-01' });
        useStore.getState().setCases([{ id: '1', title: 'Test', created_at: '2024-01-01' }]);
        useStore.getState().setViewMode('reasoning');
        useStore.getState().setIsPlaying(true);
        useStore.getState().setIsLoading(true);
        useStore.getState().setSidebarWidth(400);
      });

      // Reset
      act(() => {
        useStore.getState().reset();
      });

      const state = useStore.getState();
      expect(state.currentCase).toBeNull();
      expect(state.cases).toEqual([]);
      expect(state.viewMode).toBe('evidence');
      expect(state.isPlaying).toBe(false);
      expect(state.isLoading).toBe(false);
      expect(state.sidebarWidth).toBe(280);
    });
  });
});
