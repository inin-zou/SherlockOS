import { describe, it, expect } from 'vitest';
import {
  classifyFile,
  classifyFiles,
  getJobTypeForTier,
  groupByTier,
  TIER_CONFIG,
  type EvidenceTier,
  type ClassifiedFile,
} from './fileClassifier';

// Mock File constructor
function createMockFile(name: string, type: string, size: number = 1024): File {
  const blob = new Blob(['test content'], { type });
  return new File([blob], name, { type });
}

describe('fileClassifier', () => {
  describe('TIER_CONFIG', () => {
    it('should have all 4 tiers defined', () => {
      expect(TIER_CONFIG[0]).toBeDefined();
      expect(TIER_CONFIG[1]).toBeDefined();
      expect(TIER_CONFIG[2]).toBeDefined();
      expect(TIER_CONFIG[3]).toBeDefined();
    });

    it('should have correct tier names', () => {
      expect(TIER_CONFIG[0].name).toBe('Environment');
      expect(TIER_CONFIG[1].name).toBe('Ground Truth');
      expect(TIER_CONFIG[2].name).toBe('Electronic Logs');
      expect(TIER_CONFIG[3].name).toBe('Testimonials');
    });
  });

  describe('classifyFile', () => {
    it('should classify PDF as Tier 0 (Environment) by MIME type', () => {
      const file = createMockFile('blueprint.pdf', 'application/pdf');
      const result = classifyFile(file);
      expect(result.tier).toBe(0);
      expect(result.category).toBe('environment');
    });

    it('should classify E57 as Tier 0 (Environment) by extension', () => {
      const file = createMockFile('pointcloud.e57', 'application/octet-stream');
      const result = classifyFile(file);
      expect(result.tier).toBe(0);
    });

    it('should classify OBJ as Tier 0 (Environment) by extension', () => {
      const file = createMockFile('model.obj', 'model/obj');
      const result = classifyFile(file);
      expect(result.tier).toBe(0);
    });

    it('should classify MP4 as Tier 1 (Ground Truth) by extension', () => {
      const file = createMockFile('cctv.mp4', 'video/mp4');
      const result = classifyFile(file);
      expect(result.tier).toBe(1);
      expect(result.category).toBe('ground_truth');
    });

    it('should classify JPG as Tier 1 (Ground Truth)', () => {
      const file = createMockFile('photo.jpg', 'image/jpeg');
      const result = classifyFile(file);
      expect(result.tier).toBe(1);
    });

    it('should classify PNG as Tier 1 (Ground Truth)', () => {
      const file = createMockFile('screenshot.png', 'image/png');
      const result = classifyFile(file);
      expect(result.tier).toBe(1);
    });

    it('should classify WAV as Tier 1 (Ground Truth)', () => {
      const file = createMockFile('audio.wav', 'audio/wav');
      const result = classifyFile(file);
      expect(result.tier).toBe(1);
    });

    it('should classify JSON as Tier 2 (Electronic Logs) by extension', () => {
      const file = createMockFile('logs.json', 'application/json');
      const result = classifyFile(file);
      expect(result.tier).toBe(2);
      expect(result.category).toBe('electronic_logs');
    });

    it('should classify CSV as Tier 2 (Electronic Logs)', () => {
      const file = createMockFile('data.csv', 'text/csv');
      const result = classifyFile(file);
      expect(result.tier).toBe(2);
    });

    it('should classify LOG as Tier 2 (Electronic Logs) by extension', () => {
      const file = createMockFile('system.log', 'application/octet-stream');
      const result = classifyFile(file);
      expect(result.tier).toBe(2);
    });

    it('should classify TXT as Tier 3 (Testimonials) by extension', () => {
      const file = createMockFile('statement.txt', 'application/octet-stream');
      const result = classifyFile(file);
      expect(result.tier).toBe(3);
      expect(result.category).toBe('testimonials');
    });

    it('should classify MD as Tier 3 (Testimonials)', () => {
      const file = createMockFile('notes.md', 'text/markdown');
      const result = classifyFile(file);
      expect(result.tier).toBe(3);
    });

    it('should classify DOCX as Tier 3 (Testimonials)', () => {
      const file = createMockFile('report.docx', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document');
      const result = classifyFile(file);
      expect(result.tier).toBe(3);
    });

    it('should default unknown files to Tier 1 (Ground Truth)', () => {
      const file = createMockFile('unknown.xyz', 'application/octet-stream');
      const result = classifyFile(file);
      // Default is Tier 1 for unrecognized files
      expect(result.tier).toBe(1);
    });
  });

  describe('classifyFiles', () => {
    it('should classify multiple files', () => {
      const files = [
        createMockFile('blueprint.e57', 'application/octet-stream'),
        createMockFile('cctv.mp4', 'video/mp4'),
        createMockFile('logs.json', 'application/json'),
        createMockFile('statement.txt', 'application/octet-stream'),
      ];

      const results = classifyFiles(files);
      expect(results).toHaveLength(4);
      expect(results[0].tier).toBe(0);
      expect(results[1].tier).toBe(1);
      expect(results[2].tier).toBe(2);
      expect(results[3].tier).toBe(3);
    });

    it('should handle empty file list', () => {
      const results = classifyFiles([]);
      expect(results).toHaveLength(0);
    });
  });

  describe('groupByTier', () => {
    it('should group files by tier', () => {
      const classified: ClassifiedFile[] = [
        { file: createMockFile('a.pdf', 'application/pdf'), tier: 0, category: 'environment', icon: 'Map' },
        { file: createMockFile('b.mp4', 'video/mp4'), tier: 1, category: 'ground_truth', icon: 'Video' },
        { file: createMockFile('c.json', 'application/json'), tier: 2, category: 'electronic_logs', icon: 'FileJson' },
        { file: createMockFile('d.pdf', 'application/pdf'), tier: 0, category: 'environment', icon: 'Map' },
      ];

      const grouped = groupByTier(classified);
      expect(grouped[0]).toHaveLength(2);
      expect(grouped[1]).toHaveLength(1);
      expect(grouped[2]).toHaveLength(1);
      expect(grouped[3]).toHaveLength(0);
    });
  });

  describe('getJobTypeForTier', () => {
    it('should return reconstruction for Tier 0', () => {
      expect(getJobTypeForTier(0)).toBe('reconstruction');
    });

    it('should return scene_analysis for Tier 1', () => {
      expect(getJobTypeForTier(1)).toBe('scene_analysis');
    });

    it('should return null for Tier 2', () => {
      expect(getJobTypeForTier(2)).toBeNull();
    });

    it('should return null for Tier 3', () => {
      expect(getJobTypeForTier(3)).toBeNull();
    });
  });
});
