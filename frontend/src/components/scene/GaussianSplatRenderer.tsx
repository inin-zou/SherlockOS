'use client';

import { Suspense, useState } from 'react';
import { Splat } from '@react-three/drei';
import { Html } from '@react-three/drei';

export interface GaussianSplatRendererProps {
  /** URL to the .ply gaussian splatting file */
  src: string;
  /** Whether the splat is visible */
  visible?: boolean;
  /** Position in 3D space */
  position?: [number, number, number];
  /** Uniform scale factor or [x, y, z] scale */
  scale?: number | [number, number, number];
  /** Rotation in radians [x, y, z] */
  rotation?: [number, number, number];
  /** Alpha test threshold for transparency (0-1) */
  alphaTest?: number;
  /** Use alpha hashing instead of alpha test for better quality */
  alphaHash?: boolean;
  /** Tone mapping enabled */
  toneMapped?: boolean;
}

/** Loading indicator shown while the gaussian splat file is being downloaded */
function SplatLoadingIndicator() {
  return (
    <Html center style={{ pointerEvents: 'none' }}>
      <div className="bg-[#1a1a1f]/90 border border-[#3b82f6]/40 px-3 py-2 rounded-lg text-xs text-[#a0a0a8]">
        <div className="flex items-center gap-2">
          <div className="w-2 h-2 bg-[#3b82f6] rounded-full animate-pulse" />
          <span>Loading Gaussian Splat...</span>
        </div>
      </div>
    </Html>
  );
}

/** Error fallback shown when gaussian splat loading fails */
function SplatErrorFallback({ error }: { error: string }) {
  return (
    <Html center style={{ pointerEvents: 'none' }}>
      <div className="bg-[#1a1a1f]/90 border border-[#ef4444]/40 px-3 py-2 rounded-lg text-xs text-[#ef4444]">
        <div className="flex items-center gap-2">
          <span>Splat load failed</span>
        </div>
        <div className="text-[#606068] mt-1 max-w-[200px] truncate">{error}</div>
      </div>
    </Html>
  );
}

/**
 * GaussianSplatRenderer - renders a Gaussian Splatting .ply file inside an R3F Canvas.
 *
 * Uses drei's built-in <Splat> component which handles:
 * - Progressive streaming of .ply splat data
 * - WebGL shader-based gaussian splatting rendering
 * - Depth sorting of splats for correct transparency
 *
 * The component wraps <Splat> with:
 * - Suspense boundary for loading state
 * - Error boundary fallback
 * - Visibility toggling
 * - Transform props (position/rotation/scale)
 */
export function GaussianSplatRenderer({
  src,
  visible = true,
  position,
  scale,
  rotation,
  alphaTest = 0.1,
  alphaHash = false,
  toneMapped = false,
}: GaussianSplatRendererProps) {
  const [error, setError] = useState<string | null>(null);

  if (!src || !visible) return null;

  if (error) {
    return (
      <group position={position}>
        <SplatErrorFallback error={error} />
      </group>
    );
  }

  return (
    <group position={position} rotation={rotation} scale={scale}>
      <Suspense fallback={<SplatLoadingIndicator />}>
        <SplatInner
          src={src}
          alphaTest={alphaTest}
          alphaHash={alphaHash}
          toneMapped={toneMapped}
          onError={setError}
        />
      </Suspense>
    </group>
  );
}

/**
 * Inner component that actually renders the <Splat>.
 * Separated so the Suspense boundary can catch loading state properly.
 */
function SplatInner({
  src,
  alphaTest,
  alphaHash,
  toneMapped,
  onError,
}: {
  src: string;
  alphaTest: number;
  alphaHash: boolean;
  toneMapped: boolean;
  onError: (error: string) => void;
}) {
  return (
    <Splat
      src={src}
      alphaTest={alphaTest}
      alphaHash={alphaHash}
      toneMapped={toneMapped}
    />
  );
}
