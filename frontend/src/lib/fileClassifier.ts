/**
 * Evidence Tier Classification System
 *
 * Tier 0: Environment - Blueprints, LiDAR, CAD files
 * Tier 1: Ground Truth - CCTV, videos, photos, audio
 * Tier 2: Electronic Logs - Sensor data, access logs
 * Tier 3: Testimonials - Witness statements, reports
 */

export type EvidenceTier = 0 | 1 | 2 | 3;

export interface ClassifiedFile {
  file: File;
  tier: EvidenceTier;
  category: string;
  icon: string;
}

export interface TierInfo {
  tier: EvidenceTier;
  name: string;
  category: string;
  description: string;
  color: string;
  extensions: string[];
  mimeTypes: string[];
}

export const TIER_CONFIG: Record<EvidenceTier, TierInfo> = {
  0: {
    tier: 0,
    name: 'Environment',
    category: 'environment',
    description: 'Blueprints, LiDAR, CAD, 3D scans',
    color: '#3b82f6', // blue
    extensions: ['.pdf', '.e57', '.las', '.laz', '.dwg', '.dxf', '.obj', '.fbx', '.glb', '.gltf', '.ply', '.pcd'],
    mimeTypes: ['application/pdf', 'model/gltf-binary', 'model/gltf+json'],
  },
  1: {
    tier: 1,
    name: 'Ground Truth',
    category: 'ground_truth',
    description: 'CCTV footage, photos, audio recordings',
    color: '#10b981', // emerald
    extensions: ['.mp4', '.mov', '.avi', '.webm', '.jpg', '.jpeg', '.png', '.webp', '.heic', '.wav', '.mp3', '.m4a'],
    mimeTypes: ['video/', 'image/', 'audio/'],
  },
  2: {
    tier: 2,
    name: 'Electronic Logs',
    category: 'electronic_logs',
    description: 'Sensor data, access logs, system records',
    color: '#f59e0b', // amber
    extensions: ['.json', '.csv', '.log', '.xml', '.xlsx', '.xls'],
    mimeTypes: ['application/json', 'text/csv', 'application/xml'],
  },
  3: {
    tier: 3,
    name: 'Testimonials',
    category: 'testimonials',
    description: 'Witness statements, interview transcripts',
    color: '#8b5cf6', // purple
    extensions: ['.txt', '.md', '.docx', '.doc', '.rtf'],
    mimeTypes: ['text/plain', 'text/markdown', 'application/msword'],
  },
};

/**
 * Classify a file into an evidence tier based on extension and MIME type
 */
export function classifyFile(file: File): ClassifiedFile {
  const extension = '.' + file.name.split('.').pop()?.toLowerCase();
  const mimeType = file.type.toLowerCase();

  // Check each tier
  for (const tierNum of [0, 1, 2, 3] as EvidenceTier[]) {
    const config = TIER_CONFIG[tierNum];

    // Check extension match
    if (config.extensions.includes(extension)) {
      return {
        file,
        tier: tierNum,
        category: config.category,
        icon: getTierIcon(tierNum),
      };
    }

    // Check MIME type prefix match
    for (const mime of config.mimeTypes) {
      if (mime.endsWith('/') ? mimeType.startsWith(mime) : mimeType === mime) {
        return {
          file,
          tier: tierNum,
          category: config.category,
          icon: getTierIcon(tierNum),
        };
      }
    }
  }

  // Default to Tier 1 (Ground Truth) for unrecognized files
  return {
    file,
    tier: 1,
    category: 'ground_truth',
    icon: 'File',
  };
}

/**
 * Classify multiple files
 */
export function classifyFiles(files: FileList | File[]): ClassifiedFile[] {
  return Array.from(files).map(classifyFile);
}

/**
 * Group classified files by tier
 */
export function groupByTier(files: ClassifiedFile[]): Record<EvidenceTier, ClassifiedFile[]> {
  const grouped: Record<EvidenceTier, ClassifiedFile[]> = {
    0: [],
    1: [],
    2: [],
    3: [],
  };

  for (const file of files) {
    grouped[file.tier].push(file);
  }

  return grouped;
}

/**
 * Get icon name for tier
 */
export function getTierIcon(tier: EvidenceTier): string {
  const icons: Record<EvidenceTier, string> = {
    0: 'Map',
    1: 'Video',
    2: 'FileJson',
    3: 'MessageSquare',
  };
  return icons[tier];
}

/**
 * Get the job type to trigger for a tier
 */
export function getJobTypeForTier(tier: EvidenceTier): string | null {
  const jobTypes: Record<EvidenceTier, string | null> = {
    0: 'reconstruction',
    1: 'scene_analysis',
    2: null, // Parsed client-side
    3: null, // Uses witness-statements API
  };
  return jobTypes[tier];
}

/**
 * Check if a tier requires server-side processing
 */
export function requiresServerProcessing(tier: EvidenceTier): boolean {
  return tier === 0 || tier === 1;
}

/**
 * Parse Tier 2 (Electronic Logs) file content
 */
export async function parseTier2File(file: File): Promise<{
  events: Array<{
    timestamp: string;
    type: string;
    description: string;
    metadata?: Record<string, unknown>;
  }>;
}> {
  const text = await file.text();
  const extension = file.name.split('.').pop()?.toLowerCase();

  if (extension === 'json') {
    try {
      const data = JSON.parse(text);
      // Handle various JSON formats
      if (Array.isArray(data)) {
        return {
          events: data.map((item, i) => ({
            timestamp: item.timestamp || item.time || item.date || new Date().toISOString(),
            type: item.type || item.event_type || 'log_entry',
            description: item.description || item.message || item.data || JSON.stringify(item),
            metadata: item,
          })),
        };
      }
      return { events: [] };
    } catch {
      return { events: [] };
    }
  }

  if (extension === 'csv') {
    const lines = text.split('\n').filter(Boolean);
    const headers = lines[0]?.split(',').map((h) => h.trim().toLowerCase());

    return {
      events: lines.slice(1).map((line) => {
        const values = line.split(',');
        const row: Record<string, string> = {};
        headers?.forEach((h, i) => {
          row[h] = values[i]?.trim() || '';
        });

        return {
          timestamp: row.timestamp || row.time || row.date || new Date().toISOString(),
          type: row.type || row.event || 'log_entry',
          description: row.description || row.message || line,
          metadata: row,
        };
      }),
    };
  }

  // Default: treat as log file
  const lines = text.split('\n').filter(Boolean);
  return {
    events: lines.map((line, i) => ({
      timestamp: new Date().toISOString(),
      type: 'log_entry',
      description: line,
    })),
  };
}

/**
 * Parse Tier 3 (Testimonials) file content
 */
export async function parseTier3File(file: File): Promise<{
  source_name: string;
  content: string;
  credibility: number;
}> {
  const text = await file.text();
  const filename = file.name.replace(/\.[^.]+$/, '').replace(/_/g, ' ');

  return {
    source_name: filename,
    content: text,
    credibility: 0.7, // Default credibility
  };
}
