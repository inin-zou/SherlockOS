'use client';

import { Suspense, useMemo, useState } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import {
  OrbitControls,
  PerspectiveCamera,
  Html,
  Line,
  Points,
  PointMaterial,
} from '@react-three/drei';
import { useStore, type ViewMode } from '@/lib/store';
import { TrajectoryVisualization } from '@/components/reasoning/TrajectoryVisualization';
import { DiscrepancyHighlighter } from '@/components/reasoning/DiscrepancyHighlighter';

// Generate random point cloud for demo
function generatePointCloud(count: number, bounds: { min: number[]; max: number[] }) {
  const positions = new Float32Array(count * 3);
  const colors = new Float32Array(count * 3);

  for (let i = 0; i < count; i++) {
    const i3 = i * 3;

    // Position within bounds
    positions[i3] = bounds.min[0] + Math.random() * (bounds.max[0] - bounds.min[0]);
    positions[i3 + 1] = bounds.min[1] + Math.random() * (bounds.max[1] - bounds.min[1]);
    positions[i3 + 2] = bounds.min[2] + Math.random() * (bounds.max[2] - bounds.min[2]);

    // Color based on height (y) for visual depth
    const heightFactor = (positions[i3 + 1] - bounds.min[1]) / (bounds.max[1] - bounds.min[1]);

    // Mix between teal and purple based on height
    colors[i3] = 0.1 + heightFactor * 0.5; // R
    colors[i3 + 1] = 0.4 + Math.random() * 0.3; // G
    colors[i3 + 2] = 0.5 + heightFactor * 0.4; // B
  }

  return { positions, colors };
}

function PointCloud() {
  const { sceneGraph } = useStore();

  const bounds = sceneGraph?.bounds || { min: [-5, 0, -5], max: [5, 4, 5] };

  const { positions, colors } = useMemo(
    () => generatePointCloud(50000, { min: bounds.min as number[], max: bounds.max as number[] }),
    [bounds]
  );

  return (
    <Points positions={positions} colors={colors}>
      <PointMaterial
        vertexColors
        size={0.02}
        sizeAttenuation
        transparent
        opacity={0.8}
        depthWrite={false}
      />
    </Points>
  );
}

// Trajectory path component - using opacity state instead of ref
function TrajectoryPath({
  points,
  color = '#ffffff',
  isSelected = false
}: {
  points: [number, number, number][];
  color?: string;
  isSelected?: boolean;
}) {
  const [opacity, setOpacity] = useState(isSelected ? 1 : 0.6);

  useFrame((state) => {
    if (isSelected) {
      setOpacity(0.6 + Math.sin(state.clock.elapsedTime * 3) * 0.4);
    }
  });

  return (
    <Line
      points={points}
      color={color}
      lineWidth={isSelected ? 3 : 2}
      transparent
      opacity={opacity}
    />
  );
}

// Annotation marker with label
function Annotation({
  position,
  label,
  color = '#3b82f6',
}: {
  position: [number, number, number];
  label: string;
  color?: string;
}) {
  return (
    <group position={position}>
      {/* Marker point */}
      <mesh>
        <sphereGeometry args={[0.08, 16, 16]} />
        <meshBasicMaterial color={color} />
      </mesh>

      {/* Connecting line */}
      <Line
        points={[[0, 0, 0], [0, 0.5, 0]]}
        color={color}
        lineWidth={1}
        transparent
        opacity={0.5}
      />

      {/* Label */}
      <Html
        position={[0, 0.7, 0]}
        center
        distanceFactor={10}
        style={{ pointerEvents: 'none' }}
      >
        <div className="annotation-label">
          {label}
        </div>
      </Html>
    </group>
  );
}

// Person marker (silhouette)
function PersonMarker({
  position,
  rotation = 0
}: {
  position: [number, number, number];
  rotation?: number;
}) {
  return (
    <group position={position} rotation={[0, rotation, 0]}>
      {/* Body */}
      <mesh position={[0, 0.5, 0]}>
        <capsuleGeometry args={[0.15, 0.6, 8, 16]} />
        <meshBasicMaterial color="#ffffff" transparent opacity={0.9} />
      </mesh>
      {/* Head */}
      <mesh position={[0, 1.05, 0]}>
        <sphereGeometry args={[0.12, 16, 16]} />
        <meshBasicMaterial color="#ffffff" transparent opacity={0.9} />
      </mesh>
    </group>
  );
}

// Ground grid
function Grid() {
  return (
    <gridHelper
      args={[20, 40, '#1e1e24', '#1e1e24']}
      position={[0, 0, 0]}
    />
  );
}

// Scene content
function SceneContent() {
  const { viewMode } = useStore();

  // Demo trajectory (shown in evidence mode)
  const demoTrajectory: [number, number, number][] = [
    [-2, 0.1, 3],
    [-1, 0.1, 2],
    [0, 0.1, 1.5],
    [1, 0.1, 1],
    [2, 0.1, 0.5],
    [2.5, 0.1, 0],
    [2.5, 0.1, -0.5],
    [2, 0.5, -1],
    [1.5, 1, -1.5],
    [1, 1.5, -2],
  ];

  // Demo annotations
  const demoAnnotations = [
    { position: [-2, 0.5, 3] as [number, number, number], label: 'CCTV: Silver Sedan Passing Gate', color: '#3b82f6' },
    { position: [2, 1.8, -2] as [number, number, number], label: 'Entry Point: Broken Window', color: '#f59e0b' },
  ];

  return (
    <>
      {/* Lighting */}
      <ambientLight intensity={0.4} />
      <directionalLight position={[10, 10, 5]} intensity={0.5} />

      {/* Point cloud */}
      <PointCloud />

      {/* Grid */}
      <Grid />

      {/* Evidence mode: Show basic trajectory and annotations */}
      {viewMode === 'evidence' && (
        <>
          <TrajectoryPath
            points={demoTrajectory}
            color="#ffffff"
            isSelected={true}
          />
          <PersonMarker position={[1, 0, -2]} rotation={Math.PI / 4} />
          {demoAnnotations.map((ann, i) => (
            <Annotation
              key={i}
              position={ann.position}
              label={ann.label}
              color={ann.color}
            />
          ))}
        </>
      )}

      {/* Simulation mode: Show generated trajectories */}
      {viewMode === 'simulation' && (
        <>
          <TrajectoryVisualization />
          <PersonMarker position={[1, 0, -2]} rotation={Math.PI / 4} />
          {demoAnnotations.map((ann, i) => (
            <Annotation
              key={i}
              position={ann.position}
              label={ann.label}
              color={ann.color}
            />
          ))}
        </>
      )}

      {/* Reasoning mode: Show trajectories and discrepancies */}
      {viewMode === 'reasoning' && (
        <>
          <TrajectoryVisualization />
          <DiscrepancyHighlighter />
          <PersonMarker position={[1, 0, -2]} rotation={Math.PI / 4} />
        </>
      )}
    </>
  );
}

// Loading fallback
function LoadingScene() {
  return (
    <div className="absolute inset-0 flex items-center justify-center">
      <div className="flex flex-col items-center gap-3">
        <div className="w-8 h-8 border-2 border-[#3b82f6] border-t-transparent rounded-full animate-spin" />
        <span className="text-sm text-[#606068]">Loading scene...</span>
      </div>
    </div>
  );
}

export function SceneViewer() {
  return (
    <div className="relative w-full h-full scene-canvas">
      <Suspense fallback={<LoadingScene />}>
        <Canvas
          gl={{ antialias: true, alpha: true }}
          dpr={[1, 2]}
        >
          <PerspectiveCamera
            makeDefault
            position={[8, 6, 8]}
            fov={50}
          />
          <OrbitControls
            enablePan
            enableZoom
            enableRotate
            minDistance={2}
            maxDistance={50}
            maxPolarAngle={Math.PI / 2.1}
          />
          <SceneContent />
        </Canvas>
      </Suspense>

      {/* View mode indicator - removed, using ModeSelector in header */}

      {/* Scene info overlay */}
      <div className="absolute bottom-4 left-4 glass glass-border rounded-lg px-3 py-2 text-xs text-[#a0a0a8]">
        <div className="flex items-center gap-4">
          <span>Points: 50,000</span>
          <span>Objects: 12</span>
          <span>Trajectories: 3</span>
        </div>
      </div>
    </div>
  );
}
