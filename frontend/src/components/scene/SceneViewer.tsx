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
import type { SceneObject, SceneGraph, EvidenceCard, ObjectType } from '@/lib/types';

// Color mapping for different object types
const OBJECT_COLORS: Record<ObjectType, string> = {
  furniture: '#5c4a3d',
  door: '#6b5344',
  window: '#1a2a4a',
  wall: '#ccc7bd',
  evidence_item: '#ef4444',
  weapon: '#dc2626',
  footprint: '#f59e0b',
  bloodstain: '#991b1b',
  vehicle: '#3b82f6',
  person_marker: '#8b5cf6',
  other: '#6b7280',
};

// State colors
const STATE_COLORS: Record<string, string> = {
  visible: '#22c55e',
  occluded: '#f59e0b',
  suspicious: '#ef4444',
  removed: '#6b7280',
};

// Render a single SceneObject from the API
function DynamicSceneObject({ object, isSelected, labelOffset = 0 }: { object: SceneObject; isSelected: boolean; labelOffset?: number }) {
  const color = OBJECT_COLORS[object.type] || OBJECT_COLORS.other;
  const position: [number, number, number] = object.pose?.position || [0, 0, 0];

  // Calculate dimensions from bounding box
  const bbox = object.bbox;
  const dimensions: [number, number, number] = bbox
    ? [
        Math.abs(bbox.max[0] - bbox.min[0]) || 1,
        Math.abs(bbox.max[1] - bbox.min[1]) || 1,
        Math.abs(bbox.max[2] - bbox.min[2]) || 1,
      ]
    : [1, 1, 1];

  // Center position based on bbox
  const centerY = bbox ? (bbox.min[1] + bbox.max[1]) / 2 : position[1] + dimensions[1] / 2;
  const adjustedPosition: [number, number, number] = [position[0], centerY, position[2]];

  // Render different shapes based on type
  const renderObject = () => {
    switch (object.type) {
      case 'door':
        return (
          <group position={adjustedPosition}>
            <mesh castShadow>
              <boxGeometry args={[dimensions[0], dimensions[1], 0.1]} />
              <meshStandardMaterial color={color} roughness={0.6} />
            </mesh>
            {/* Door handle */}
            <mesh position={[dimensions[0] * 0.35, 0, 0.08]} castShadow>
              <sphereGeometry args={[0.06, 16, 16]} />
              <meshStandardMaterial color="#c9a227" roughness={0.3} metalness={0.8} />
            </mesh>
          </group>
        );

      case 'window':
        return (
          <group position={adjustedPosition}>
            <mesh>
              <planeGeometry args={[dimensions[0], dimensions[1]]} />
              <meshStandardMaterial
                color="#1a2a4a"
                roughness={0.1}
                metalness={0.3}
                emissive="#1a3050"
                emissiveIntensity={0.3}
              />
            </mesh>
            {/* Window frame */}
            <Line
              points={[
                [-dimensions[0]/2, -dimensions[1]/2, 0.01],
                [dimensions[0]/2, -dimensions[1]/2, 0.01],
                [dimensions[0]/2, dimensions[1]/2, 0.01],
                [-dimensions[0]/2, dimensions[1]/2, 0.01],
                [-dimensions[0]/2, -dimensions[1]/2, 0.01],
              ]}
              color="#e8e8e8"
              lineWidth={2}
            />
          </group>
        );

      case 'wall':
        return (
          <mesh position={adjustedPosition} receiveShadow>
            <planeGeometry args={[dimensions[0], dimensions[1]]} />
            <meshStandardMaterial color={color} roughness={0.9} />
          </mesh>
        );

      case 'evidence_item':
      case 'weapon':
      case 'footprint':
      case 'bloodstain':
        // Evidence markers - show as highlighted markers
        return (
          <group position={position}>
            {/* Marker cone */}
            <mesh rotation={[Math.PI, 0, 0]} castShadow>
              <coneGeometry args={[0.15, 0.3, 4]} />
              <meshStandardMaterial
                color={color}
                emissive={color}
                emissiveIntensity={isSelected ? 0.5 : 0.2}
              />
            </mesh>
            {/* Glow ring on floor */}
            <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0.01, 0]}>
              <ringGeometry args={[0.2, 0.35, 32]} />
              <meshStandardMaterial
                color={color}
                transparent
                opacity={0.5}
                emissive={color}
                emissiveIntensity={0.3}
              />
            </mesh>
            {/* Label - staggered height to prevent overlap */}
            <Html position={[0, 0.5 + labelOffset * 0.4, 0]} center style={{ pointerEvents: 'none' }}>
              <div
                className="px-2 py-1 rounded text-xs font-medium shadow-lg"
                style={{
                  backgroundColor: `${color}ee`,
                  color: '#fff',
                  whiteSpace: 'nowrap',
                  maxWidth: '150px',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                }}
              >
                {object.label}
              </div>
            </Html>
          </group>
        );

      case 'person_marker':
        return (
          <group position={position}>
            {/* Body */}
            <mesh position={[0, 0.5, 0]} castShadow>
              <capsuleGeometry args={[0.15, 0.6, 8, 16]} />
              <meshStandardMaterial color="#8b5cf6" transparent opacity={0.9} />
            </mesh>
            {/* Head */}
            <mesh position={[0, 1.05, 0]} castShadow>
              <sphereGeometry args={[0.12, 16, 16]} />
              <meshStandardMaterial color="#8b5cf6" transparent opacity={0.9} />
            </mesh>
          </group>
        );

      case 'vehicle':
        return (
          <group position={adjustedPosition}>
            {/* Simple car shape */}
            <mesh castShadow>
              <boxGeometry args={[dimensions[0], dimensions[1] * 0.6, dimensions[2]]} />
              <meshStandardMaterial color={color} roughness={0.4} metalness={0.6} />
            </mesh>
            <mesh position={[0, dimensions[1] * 0.35, -dimensions[2] * 0.1]} castShadow>
              <boxGeometry args={[dimensions[0] * 0.8, dimensions[1] * 0.4, dimensions[2] * 0.6]} />
              <meshStandardMaterial color={color} roughness={0.4} metalness={0.6} />
            </mesh>
          </group>
        );

      case 'furniture':
      default:
        return (
          <group position={adjustedPosition}>
            <mesh castShadow receiveShadow>
              <boxGeometry args={dimensions} />
              <meshStandardMaterial
                color={color}
                roughness={0.6}
                emissive={isSelected ? '#3b82f6' : '#000000'}
                emissiveIntensity={isSelected ? 0.2 : 0}
              />
            </mesh>
            {/* Label for furniture and other objects */}
            <Html position={[0, dimensions[1] / 2 + 0.3 + labelOffset * 0.3, 0]} center style={{ pointerEvents: 'none' }}>
              <div
                className="px-2 py-1 rounded text-xs shadow-lg"
                style={{
                  backgroundColor: '#1f1f24ee',
                  color: '#e8e8e8',
                  whiteSpace: 'nowrap',
                  maxWidth: '140px',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  border: '1px solid #333',
                }}
              >
                {object.label}
                <span className="text-[#8b8b96] ml-1 text-[10px]">
                  {Math.round((object.confidence || 0.8) * 100)}%
                </span>
              </div>
            </Html>
          </group>
        );
    }
  };

  return (
    <group>
      {renderObject()}
      {/* Selection outline for suspicious items */}
      {object.state === 'suspicious' && (
        <mesh position={adjustedPosition}>
          <boxGeometry args={[dimensions[0] + 0.1, dimensions[1] + 0.1, dimensions[2] + 0.1]} />
          <meshBasicMaterial color="#ef4444" wireframe transparent opacity={0.5} />
        </mesh>
      )}
    </group>
  );
}

// Estimate position from position description text
function estimatePositionFromDescription(description: string | undefined, index: number, total: number): [number, number, number] {
  if (!description) {
    // Default spread in a grid pattern
    const cols = Math.ceil(Math.sqrt(total));
    const row = Math.floor(index / cols);
    const col = index % cols;
    return [
      (col - cols / 2) * 2 + 1,
      0,
      (row - Math.ceil(total / cols) / 2) * 2
    ];
  }

  const desc = description.toLowerCase();
  let x = 0, z = 0, y = 0;

  // Parse horizontal position
  if (desc.includes('left')) x = -3 - Math.random() * 2;
  else if (desc.includes('right')) x = 3 + Math.random() * 2;
  else if (desc.includes('center')) x = (Math.random() - 0.5) * 2;
  else x = (Math.random() - 0.5) * 6;

  // Parse depth position
  if (desc.includes('back') || desc.includes('rear') || desc.includes('window')) z = -4 - Math.random();
  else if (desc.includes('front') || desc.includes('entrance') || desc.includes('door')) z = 4 + Math.random();
  else if (desc.includes('hallway')) z = 5;
  else z = (Math.random() - 0.5) * 6;

  // Parse height
  if (desc.includes('floor') || desc.includes('ground') || desc.includes('footprint')) y = 0.01;
  else if (desc.includes('wall') || desc.includes('hanging')) y = 1.5;
  else if (desc.includes('desk') || desc.includes('table')) y = 0.8;
  else y = 0;

  return [x, y, z];
}

// Render all objects from the SceneGraph
function DynamicSceneObjects({ sceneGraph }: { sceneGraph: SceneGraph }) {
  const { selectedObjectIds } = useStore();

  if (!sceneGraph?.objects || sceneGraph.objects.length === 0) {
    return null;
  }

  // Assign positions to objects that don't have real positions
  const objectsWithPositions = useMemo(() => {
    return sceneGraph.objects.map((obj, index) => {
      const hasRealPosition = obj.pose?.position &&
        (obj.pose.position[0] !== 0 || obj.pose.position[1] !== 0 || obj.pose.position[2] !== 0);

      if (hasRealPosition) {
        return obj;
      }

      // Estimate position from metadata description
      const positionDescription = obj.metadata?.position_description as string | undefined;
      const estimatedPosition = estimatePositionFromDescription(
        positionDescription,
        index,
        sceneGraph.objects.length
      );

      return {
        ...obj,
        pose: {
          ...obj.pose,
          position: estimatedPosition,
        },
      };
    });
  }, [sceneGraph.objects]);

  return (
    <group>
      {objectsWithPositions.map((obj, index) => (
        <DynamicSceneObject
          key={obj.id}
          object={obj}
          isSelected={selectedObjectIds.includes(obj.id)}
          labelOffset={index % 3}
        />
      ))}
    </group>
  );
}

// Render evidence cards as 3D annotations
function DynamicEvidenceAnnotations({ evidence }: { evidence: EvidenceCard[] }) {
  if (!evidence || evidence.length === 0) return null;

  return (
    <group>
      {evidence.slice(0, 10).map((ev, index) => {
        // Position evidence annotations in a spread pattern
        const angle = (index / evidence.length) * Math.PI * 2;
        const radius = 3;
        const position: [number, number, number] = [
          Math.cos(angle) * radius,
          2,
          Math.sin(angle) * radius,
        ];

        return (
          <group key={ev.id} position={position}>
            <Html center style={{ pointerEvents: 'none' }}>
              <div className="px-2 py-1 rounded bg-[#1f1f24]/90 border border-[#3b82f6]/50 text-xs text-white max-w-[150px]">
                <div className="font-medium truncate">{ev.title}</div>
                <div className="text-[#8b8b96] text-[10px]">
                  {Math.round(ev.confidence * 100)}% confidence
                </div>
              </div>
            </Html>
          </group>
        );
      })}
    </group>
  );
}

// Room geometry component - realistic office/crime scene
function RoomGeometry() {
  return (
    <group>
      {/* Floor - wooden parquet look */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0, 0]} receiveShadow>
        <planeGeometry args={[14, 12]} />
        <meshStandardMaterial color="#5c4a3d" roughness={0.8} metalness={0.1} />
      </mesh>

      {/* Floor planks pattern */}
      {Array.from({ length: 7 }).map((_, i) => (
        <mesh key={`plank-${i}`} rotation={[-Math.PI / 2, 0, 0]} position={[-6 + i * 2, 0.005, 0]}>
          <planeGeometry args={[0.02, 12]} />
          <meshStandardMaterial color="#4a3d32" roughness={0.9} />
        </mesh>
      ))}

      {/* Back wall - cream/beige office wall */}
      <mesh position={[0, 2, -6]} receiveShadow>
        <planeGeometry args={[14, 4]} />
        <meshStandardMaterial color="#d4cfc5" roughness={0.9} metalness={0} />
      </mesh>

      {/* Back wall baseboard - dark wood */}
      <mesh position={[0, 0.1, -5.9]} castShadow>
        <boxGeometry args={[14, 0.2, 0.15]} />
        <meshStandardMaterial color="#3d2e24" roughness={0.7} />
      </mesh>

      {/* Left wall - slightly different shade */}
      <mesh position={[-7, 2, 0]} rotation={[0, Math.PI / 2, 0]} receiveShadow>
        <planeGeometry args={[12, 4]} />
        <meshStandardMaterial color="#ccc7bd" roughness={0.9} metalness={0} />
      </mesh>

      {/* Right wall */}
      <mesh position={[7, 2, 0]} rotation={[0, -Math.PI / 2, 0]} receiveShadow>
        <planeGeometry args={[12, 4]} />
        <meshStandardMaterial color="#ccc7bd" roughness={0.9} metalness={0} />
      </mesh>

      {/* Window on right wall - night sky with moonlight */}
      <mesh position={[6.85, 2.2, -1.5]} rotation={[0, -Math.PI / 2, 0]}>
        <planeGeometry args={[3, 2.2]} />
        <meshStandardMaterial color="#1a2a4a" roughness={0.1} metalness={0.3} emissive="#1a3050" emissiveIntensity={0.3} />
      </mesh>
      {/* Window frame - white */}
      <mesh position={[6.8, 2.2, -1.5]} rotation={[0, -Math.PI / 2, 0]}>
        <boxGeometry args={[3.2, 0.1, 0.1]} />
        <meshStandardMaterial color="#e8e8e8" roughness={0.5} />
      </mesh>
      <mesh position={[6.8, 2.2, -1.5]} rotation={[0, -Math.PI / 2, 0]}>
        <boxGeometry args={[0.1, 2.4, 0.1]} />
        <meshStandardMaterial color="#e8e8e8" roughness={0.5} />
      </mesh>
      {/* Window sill */}
      <mesh position={[6.7, 1.05, -1.5]} castShadow>
        <boxGeometry args={[0.2, 0.08, 3.4]} />
        <meshStandardMaterial color="#e8e8e8" roughness={0.5} />
      </mesh>

      {/* Door on back wall - wooden */}
      <mesh position={[-4, 1.2, -5.88]} castShadow>
        <boxGeometry args={[1.2, 2.4, 0.08]} />
        <meshStandardMaterial color="#6b5344" roughness={0.6} />
      </mesh>
      {/* Door handle */}
      <mesh position={[-3.5, 1.1, -5.82]} castShadow>
        <sphereGeometry args={[0.06, 16, 16]} />
        <meshStandardMaterial color="#c9a227" roughness={0.3} metalness={0.8} />
      </mesh>
      {/* Door frame */}
      <mesh position={[-4, 2.45, -5.85]}>
        <boxGeometry args={[1.5, 0.1, 0.15]} />
        <meshStandardMaterial color="#e8e4dc" roughness={0.7} />
      </mesh>

      {/* Ceiling */}
      <mesh position={[0, 4, 0]} rotation={[Math.PI / 2, 0, 0]}>
        <planeGeometry args={[14, 12]} />
        <meshStandardMaterial color="#f0ece4" roughness={0.95} />
      </mesh>

      {/* Ceiling light fixture */}
      <mesh position={[0, 3.9, -2]}>
        <cylinderGeometry args={[0.4, 0.5, 0.15, 16]} />
        <meshStandardMaterial color="#e8e4dc" roughness={0.5} emissive="#fffaf0" emissiveIntensity={0.5} />
      </mesh>

      {/* Wall art / picture frame on back wall */}
      <mesh position={[3, 2.5, -5.9]} castShadow>
        <boxGeometry args={[1.5, 1, 0.05]} />
        <meshStandardMaterial color="#2a2520" roughness={0.4} />
      </mesh>
      <mesh position={[3, 2.5, -5.87]}>
        <planeGeometry args={[1.3, 0.8]} />
        <meshStandardMaterial color="#8a9aa8" roughness={0.8} />
      </mesh>

      {/* Wall clock */}
      <mesh position={[-1, 3, -5.9]} rotation={[Math.PI / 2, 0, 0]}>
        <cylinderGeometry args={[0.3, 0.3, 0.05, 32]} />
        <meshStandardMaterial color="#1a1a1a" roughness={0.3} />
      </mesh>
      <mesh position={[-1, 3, -5.87]}>
        <circleGeometry args={[0.25, 32]} />
        <meshStandardMaterial color="#f5f5f0" roughness={0.9} />
      </mesh>

      {/* Electrical outlet on right wall */}
      <mesh position={[6.9, 0.4, 1]} rotation={[0, -Math.PI / 2, 0]}>
        <boxGeometry args={[0.12, 0.08, 0.02]} />
        <meshStandardMaterial color="#f0ece4" roughness={0.8} />
      </mesh>
    </group>
  );
}

// Office furniture - realistic with proper materials
function OfficeFurniture() {
  return (
    <group>
      {/* Executive desk - dark wood */}
      <group position={[3, 0, -4]}>
        {/* Desk top - polished wood */}
        <mesh position={[0, 0.75, 0]} castShadow receiveShadow>
          <boxGeometry args={[2.8, 0.08, 1.3]} />
          <meshStandardMaterial color="#5c3d2e" roughness={0.3} metalness={0.1} />
        </mesh>
        {/* Desk drawer section */}
        <mesh position={[0.8, 0.4, 0]} castShadow>
          <boxGeometry args={[0.9, 0.6, 1.2]} />
          <meshStandardMaterial color="#4a3228" roughness={0.5} />
        </mesh>
        {/* Drawer handles - brass */}
        <mesh position={[0.8, 0.55, 0.61]} castShadow>
          <boxGeometry args={[0.15, 0.03, 0.03]} />
          <meshStandardMaterial color="#c9a227" roughness={0.3} metalness={0.8} />
        </mesh>
        <mesh position={[0.8, 0.3, 0.61]} castShadow>
          <boxGeometry args={[0.15, 0.03, 0.03]} />
          <meshStandardMaterial color="#c9a227" roughness={0.3} metalness={0.8} />
        </mesh>
        {/* Desk legs */}
        <mesh position={[-1.3, 0.35, -0.55]} castShadow>
          <boxGeometry args={[0.08, 0.7, 0.08]} />
          <meshStandardMaterial color="#3d2820" roughness={0.5} />
        </mesh>
        <mesh position={[-1.3, 0.35, 0.55]} castShadow>
          <boxGeometry args={[0.08, 0.7, 0.08]} />
          <meshStandardMaterial color="#3d2820" roughness={0.5} />
        </mesh>

        {/* Computer monitor */}
        <mesh position={[-0.3, 1.15, -0.4]} castShadow>
          <boxGeometry args={[1.1, 0.65, 0.05]} />
          <meshStandardMaterial color="#1a1a1a" roughness={0.2} metalness={0.5} />
        </mesh>
        {/* Monitor screen - glowing */}
        <mesh position={[-0.3, 1.15, -0.37]}>
          <planeGeometry args={[1.0, 0.58]} />
          <meshStandardMaterial color="#0a1520" emissive="#1a4a80" emissiveIntensity={0.4} roughness={0.1} />
        </mesh>
        {/* Monitor stand */}
        <mesh position={[-0.3, 0.85, -0.4]} castShadow>
          <boxGeometry args={[0.25, 0.12, 0.15]} />
          <meshStandardMaterial color="#2a2a2a" roughness={0.3} metalness={0.6} />
        </mesh>

        {/* Keyboard */}
        <mesh position={[-0.3, 0.8, 0.1]} castShadow>
          <boxGeometry args={[0.5, 0.02, 0.18]} />
          <meshStandardMaterial color="#2a2a2a" roughness={0.6} />
        </mesh>
        {/* Mouse */}
        <mesh position={[0.2, 0.8, 0.15]} castShadow>
          <boxGeometry args={[0.08, 0.025, 0.12]} />
          <meshStandardMaterial color="#1a1a1a" roughness={0.4} />
        </mesh>

        {/* Desk lamp */}
        <mesh position={[-1.1, 0.8, -0.3]} castShadow>
          <cylinderGeometry args={[0.08, 0.1, 0.03, 16]} />
          <meshStandardMaterial color="#2a2a2a" roughness={0.3} metalness={0.7} />
        </mesh>
        <mesh position={[-1.1, 1.1, -0.3]} castShadow>
          <coneGeometry args={[0.12, 0.15, 16]} />
          <meshStandardMaterial color="#c9a227" roughness={0.4} metalness={0.6} />
        </mesh>

        {/* Papers scattered */}
        <mesh position={[0.5, 0.8, 0.3]} rotation={[0, 0.2, 0]} castShadow>
          <boxGeometry args={[0.3, 0.01, 0.4]} />
          <meshStandardMaterial color="#f5f5e8" roughness={0.9} />
        </mesh>

        {/* Coffee mug */}
        <mesh position={[-0.9, 0.85, 0.4]} castShadow>
          <cylinderGeometry args={[0.04, 0.035, 0.1, 16]} />
          <meshStandardMaterial color="#d4d4c8" roughness={0.7} />
        </mesh>
      </group>

      {/* Leather office chair */}
      <group position={[3, 0, -2.2]}>
        <mesh position={[0, 0.5, 0]} castShadow>
          <boxGeometry args={[0.55, 0.12, 0.55]} />
          <meshStandardMaterial color="#2a1810" roughness={0.6} />
        </mesh>
        <mesh position={[0, 0.95, -0.22]} castShadow>
          <boxGeometry args={[0.5, 0.7, 0.1]} />
          <meshStandardMaterial color="#2a1810" roughness={0.6} />
        </mesh>
        {/* Armrests */}
        <mesh position={[-0.3, 0.65, 0]} castShadow>
          <boxGeometry args={[0.05, 0.05, 0.35]} />
          <meshStandardMaterial color="#1a1a1a" roughness={0.3} metalness={0.5} />
        </mesh>
        <mesh position={[0.3, 0.65, 0]} castShadow>
          <boxGeometry args={[0.05, 0.05, 0.35]} />
          <meshStandardMaterial color="#1a1a1a" roughness={0.3} metalness={0.5} />
        </mesh>
        {/* Chair stem */}
        <mesh position={[0, 0.3, 0]} castShadow>
          <cylinderGeometry args={[0.03, 0.03, 0.35, 8]} />
          <meshStandardMaterial color="#3a3a3a" roughness={0.3} metalness={0.7} />
        </mesh>
        {/* Chair base */}
        <mesh position={[0, 0.1, 0]} castShadow>
          <cylinderGeometry args={[0.25, 0.25, 0.05, 5]} />
          <meshStandardMaterial color="#2a2a2a" roughness={0.3} metalness={0.6} />
        </mesh>
      </group>

      {/* Metal filing cabinet */}
      <group position={[-4.5, 0, -4.8]}>
        <mesh position={[0, 0.65, 0]} castShadow receiveShadow>
          <boxGeometry args={[0.7, 1.3, 0.55]} />
          <meshStandardMaterial color="#5a6270" roughness={0.4} metalness={0.6} />
        </mesh>
        {/* Drawer fronts */}
        <mesh position={[0, 1.0, 0.28]} castShadow>
          <boxGeometry args={[0.62, 0.35, 0.02]} />
          <meshStandardMaterial color="#4a5260" roughness={0.35} metalness={0.65} />
        </mesh>
        <mesh position={[0, 0.55, 0.28]} castShadow>
          <boxGeometry args={[0.62, 0.35, 0.02]} />
          <meshStandardMaterial color="#4a5260" roughness={0.35} metalness={0.65} />
        </mesh>
        {/* Drawer handles - chrome */}
        <mesh position={[0, 1.0, 0.3]} castShadow>
          <boxGeometry args={[0.2, 0.025, 0.025]} />
          <meshStandardMaterial color="#e8e8e8" roughness={0.1} metalness={0.9} />
        </mesh>
        <mesh position={[0, 0.55, 0.3]} castShadow>
          <boxGeometry args={[0.2, 0.025, 0.025]} />
          <meshStandardMaterial color="#e8e8e8" roughness={0.1} metalness={0.9} />
        </mesh>
      </group>

      {/* Leather sofa */}
      <group position={[-5.5, 0, 1.5]}>
        {/* Seat base */}
        <mesh position={[0, 0.25, 0]} castShadow receiveShadow>
          <boxGeometry args={[0.9, 0.3, 2.0]} />
          <meshStandardMaterial color="#3d2820" roughness={0.7} />
        </mesh>
        {/* Seat cushions */}
        <mesh position={[0.1, 0.45, -0.45]} castShadow>
          <boxGeometry args={[0.7, 0.15, 0.85]} />
          <meshStandardMaterial color="#4a3028" roughness={0.6} />
        </mesh>
        <mesh position={[0.1, 0.45, 0.45]} castShadow>
          <boxGeometry args={[0.7, 0.15, 0.85]} />
          <meshStandardMaterial color="#4a3028" roughness={0.6} />
        </mesh>
        {/* Backrest */}
        <mesh position={[-0.35, 0.65, 0]} castShadow>
          <boxGeometry args={[0.2, 0.65, 2.0]} />
          <meshStandardMaterial color="#3d2820" roughness={0.7} />
        </mesh>
        {/* Armrests */}
        <mesh position={[0, 0.5, 0.95]} castShadow>
          <boxGeometry args={[0.7, 0.35, 0.15]} />
          <meshStandardMaterial color="#3d2820" roughness={0.7} />
        </mesh>
        <mesh position={[0, 0.5, -0.95]} castShadow>
          <boxGeometry args={[0.7, 0.35, 0.15]} />
          <meshStandardMaterial color="#3d2820" roughness={0.7} />
        </mesh>
        {/* Throw pillow */}
        <mesh position={[0.15, 0.6, 0.6]} rotation={[0.1, 0.3, 0.1]} castShadow>
          <boxGeometry args={[0.15, 0.35, 0.35]} />
          <meshStandardMaterial color="#8b4513" roughness={0.8} />
        </mesh>
      </group>

      {/* Glass coffee table */}
      <group position={[-3.5, 0, 1.5]}>
        {/* Glass top */}
        <mesh position={[0, 0.4, 0]} castShadow>
          <boxGeometry args={[0.9, 0.02, 1.2]} />
          <meshStandardMaterial color="#a8c8d8" roughness={0.05} metalness={0.1} transparent opacity={0.6} />
        </mesh>
        {/* Metal legs */}
        <mesh position={[-0.4, 0.2, -0.55]} castShadow>
          <boxGeometry args={[0.03, 0.4, 0.03]} />
          <meshStandardMaterial color="#c0c0c0" roughness={0.2} metalness={0.8} />
        </mesh>
        <mesh position={[0.4, 0.2, -0.55]} castShadow>
          <boxGeometry args={[0.03, 0.4, 0.03]} />
          <meshStandardMaterial color="#c0c0c0" roughness={0.2} metalness={0.8} />
        </mesh>
        <mesh position={[-0.4, 0.2, 0.55]} castShadow>
          <boxGeometry args={[0.03, 0.4, 0.03]} />
          <meshStandardMaterial color="#c0c0c0" roughness={0.2} metalness={0.8} />
        </mesh>
        <mesh position={[0.4, 0.2, 0.55]} castShadow>
          <boxGeometry args={[0.03, 0.4, 0.03]} />
          <meshStandardMaterial color="#c0c0c0" roughness={0.2} metalness={0.8} />
        </mesh>
        {/* Magazine */}
        <mesh position={[0.1, 0.42, 0.2]} rotation={[0, 0.4, 0]} castShadow>
          <boxGeometry args={[0.25, 0.01, 0.35]} />
          <meshStandardMaterial color="#d35400" roughness={0.8} />
        </mesh>
      </group>

      {/* Tall bookshelf */}
      <group position={[-6.6, 0, 3.5]}>
        <mesh position={[0, 1.3, 0]} castShadow receiveShadow>
          <boxGeometry args={[0.35, 2.6, 1.4]} />
          <meshStandardMaterial color="#3d2820" roughness={0.6} />
        </mesh>
        {/* Shelves */}
        {[0.5, 1.1, 1.7, 2.3].map((y, i) => (
          <mesh key={`shelf-${i}`} position={[0.18, y, 0]} castShadow>
            <boxGeometry args={[0.02, 0.03, 1.3]} />
            <meshStandardMaterial color="#4a3228" roughness={0.5} />
          </mesh>
        ))}
        {/* Books - various colors */}
        <mesh position={[0.1, 0.75, -0.3]} castShadow>
          <boxGeometry args={[0.12, 0.45, 0.08]} />
          <meshStandardMaterial color="#8b0000" roughness={0.8} />
        </mesh>
        <mesh position={[0.1, 0.75, -0.18]} castShadow>
          <boxGeometry args={[0.12, 0.42, 0.07]} />
          <meshStandardMaterial color="#1a4a20" roughness={0.8} />
        </mesh>
        <mesh position={[0.1, 0.75, -0.08]} castShadow>
          <boxGeometry args={[0.12, 0.48, 0.08]} />
          <meshStandardMaterial color="#2a2a6a" roughness={0.8} />
        </mesh>
        <mesh position={[0.1, 1.35, -0.25]} castShadow>
          <boxGeometry args={[0.12, 0.44, 0.07]} />
          <meshStandardMaterial color="#4a1a4a" roughness={0.8} />
        </mesh>
        <mesh position={[0.1, 1.35, -0.12]} castShadow>
          <boxGeometry args={[0.12, 0.38, 0.09]} />
          <meshStandardMaterial color="#c9a227" roughness={0.8} />
        </mesh>
      </group>

      {/* Potted plant */}
      <group position={[5.5, 0, 3]}>
        <mesh position={[0, 0.2, 0]} castShadow>
          <cylinderGeometry args={[0.18, 0.14, 0.4, 16]} />
          <meshStandardMaterial color="#6b4423" roughness={0.8} />
        </mesh>
        <mesh position={[0, 0.38, 0]}>
          <cylinderGeometry args={[0.16, 0.16, 0.05, 16]} />
          <meshStandardMaterial color="#3d2a1a" roughness={0.95} />
        </mesh>
        {/* Plant leaves */}
        {[0, 1, 2, 3, 4, 5].map((i) => (
          <mesh key={`leaf-${i}`} position={[Math.cos(i * 1.05) * 0.12, 0.55 + (i % 2) * 0.15, Math.sin(i * 1.05) * 0.12]} rotation={[0.3, i * 1.05, 0.2]} castShadow>
            <planeGeometry args={[0.15, 0.35]} />
            <meshStandardMaterial color="#2d5a30" roughness={0.8} side={2} />
          </mesh>
        ))}
      </group>

      {/* Trash bin */}
      <group position={[1.5, 0, -2.8]}>
        <mesh position={[0, 0.2, 0]} castShadow>
          <cylinderGeometry args={[0.12, 0.1, 0.4, 16]} />
          <meshStandardMaterial color="#3a3a3a" roughness={0.5} metalness={0.4} />
        </mesh>
      </group>

      {/* Area rug under coffee table */}
      <mesh position={[-4.2, 0.01, 1.5]} rotation={[-Math.PI / 2, 0, 0]} receiveShadow>
        <planeGeometry args={[2.5, 3]} />
        <meshStandardMaterial color="#6b4535" roughness={0.95} />
      </mesh>
    </group>
  );
}

// Evidence markers in the room
function EvidenceMarkers() {
  return (
    <group>
      {/* Footprint markers near window - yellow crime scene markers */}
      <EvidenceNumberMarker position={[5.5, 0.18, -2]} number={1} color="#f59e0b" />
      <EvidenceNumberMarker position={[4.5, 0.18, -2.5]} number={2} color="#f59e0b" />

      {/* Evidence near desk */}
      <EvidenceNumberMarker position={[2.5, 0.9, -4.2]} number={3} color="#ef4444" />

      {/* Evidence at filing cabinet */}
      <EvidenceNumberMarker position={[-4.2, 1.5, -4.5]} number={4} color="#ef4444" />

      {/* Footprint outlines on floor - path from window */}
      <group position={[5.8, 0.02, -1.8]}>
        <mesh rotation={[-Math.PI / 2, 0, 0.5]}>
          <planeGeometry args={[0.12, 0.32]} />
          <meshBasicMaterial color="#ef4444" transparent opacity={0.8} />
        </mesh>
      </group>
      <group position={[5.2, 0.02, -2.2]}>
        <mesh rotation={[-Math.PI / 2, 0, 0.4]}>
          <planeGeometry args={[0.12, 0.32]} />
          <meshBasicMaterial color="#ef4444" transparent opacity={0.8} />
        </mesh>
      </group>
      <group position={[4.6, 0.02, -2.6]}>
        <mesh rotation={[-Math.PI / 2, 0, 0.3]}>
          <planeGeometry args={[0.12, 0.32]} />
          <meshBasicMaterial color="#ef4444" transparent opacity={0.8} />
        </mesh>
      </group>
    </group>
  );
}

function EvidenceNumberMarker({ position, number, color = '#f59e0b' }: { position: [number, number, number]; number: number; color?: string }) {
  return (
    <group position={position}>
      {/* Cone marker */}
      <mesh rotation={[Math.PI, 0, 0]}>
        <coneGeometry args={[0.12, 0.25, 4]} />
        <meshBasicMaterial color={color} />
      </mesh>
      {/* Label */}
      <Html
        position={[0, 0.2, 0]}
        center
        style={{ pointerEvents: 'none' }}
      >
        <div
          className="text-black text-xs font-bold w-5 h-5 rounded-full flex items-center justify-center shadow-lg"
          style={{ backgroundColor: color }}
        >
          {number}
        </div>
      </Html>
    </group>
  );
}

// Generate random point cloud for ambient effect (reduced)
function generatePointCloud(count: number, bounds: { min: number[]; max: number[] }) {
  const positions = new Float32Array(count * 3);
  const colors = new Float32Array(count * 3);

  for (let i = 0; i < count; i++) {
    const i3 = i * 3;

    // Position within bounds
    positions[i3] = bounds.min[0] + Math.random() * (bounds.max[0] - bounds.min[0]);
    positions[i3 + 1] = bounds.min[1] + Math.random() * (bounds.max[1] - bounds.min[1]);
    positions[i3 + 2] = bounds.min[2] + Math.random() * (bounds.max[2] - bounds.min[2]);

    // Subtle color
    colors[i3] = 0.15;
    colors[i3 + 1] = 0.2;
    colors[i3 + 2] = 0.25;
  }

  return { positions, colors };
}

function PointCloud() {
  const { sceneGraph } = useStore();

  const bounds = sceneGraph?.bounds || { min: [-6, 0, -5], max: [6, 4, 5] };

  const { positions, colors } = useMemo(
    () => generatePointCloud(5000, { min: bounds.min as number[], max: bounds.max as number[] }),
    [bounds]
  );

  return (
    <Points positions={positions} colors={colors}>
      <PointMaterial
        vertexColors
        size={0.015}
        sizeAttenuation
        transparent
        opacity={0.3}
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


// Reconstructed room geometry - generated from scene bounds
function ReconstructedRoom({ bounds }: { bounds: { min: [number, number, number]; max: [number, number, number] } }) {
  const [minX, minY, minZ] = bounds.min;
  const [maxX, maxY, maxZ] = bounds.max;

  const width = maxX - minX;
  const height = maxY - minY;
  const depth = maxZ - minZ;
  const centerX = (minX + maxX) / 2;
  const centerY = (minY + maxY) / 2;
  const centerZ = (minZ + maxZ) / 2;

  return (
    <group>
      {/* Floor */}
      <mesh rotation={[-Math.PI / 2, 0, 0]} position={[centerX, minY, centerZ]} receiveShadow>
        <planeGeometry args={[width, depth]} />
        <meshStandardMaterial color="#5c4a3d" roughness={0.8} metalness={0.1} />
      </mesh>

      {/* Grid on floor */}
      <gridHelper
        args={[Math.max(width, depth), Math.round(Math.max(width, depth) * 2), '#4a4a50', '#3a3a40']}
        position={[centerX, minY + 0.01, centerZ]}
      />

      {/* Back wall */}
      <mesh position={[centerX, centerY, minZ]} receiveShadow>
        <planeGeometry args={[width, height]} />
        <meshStandardMaterial color="#d4cfc5" roughness={0.9} metalness={0} />
      </mesh>

      {/* Left wall */}
      <mesh position={[minX, centerY, centerZ]} rotation={[0, Math.PI / 2, 0]} receiveShadow>
        <planeGeometry args={[depth, height]} />
        <meshStandardMaterial color="#ccc7bd" roughness={0.9} metalness={0} />
      </mesh>

      {/* Right wall */}
      <mesh position={[maxX, centerY, centerZ]} rotation={[0, -Math.PI / 2, 0]} receiveShadow>
        <planeGeometry args={[depth, height]} />
        <meshStandardMaterial color="#ccc7bd" roughness={0.9} metalness={0} />
      </mesh>

      {/* Ceiling */}
      <mesh position={[centerX, maxY, centerZ]} rotation={[Math.PI / 2, 0, 0]}>
        <planeGeometry args={[width, depth]} />
        <meshStandardMaterial color="#f0ece4" roughness={0.95} transparent opacity={0.3} />
      </mesh>

      {/* Baseboard trim */}
      <mesh position={[centerX, minY + 0.1, minZ + 0.08]} castShadow>
        <boxGeometry args={[width, 0.2, 0.15]} />
        <meshStandardMaterial color="#3d2e24" roughness={0.7} />
      </mesh>
      <mesh position={[minX + 0.08, minY + 0.1, centerZ]} castShadow>
        <boxGeometry args={[0.15, 0.2, depth]} />
        <meshStandardMaterial color="#3d2e24" roughness={0.7} />
      </mesh>
      <mesh position={[maxX - 0.08, minY + 0.1, centerZ]} castShadow>
        <boxGeometry args={[0.15, 0.2, depth]} />
        <meshStandardMaterial color="#3d2e24" roughness={0.7} />
      </mesh>
    </group>
  );
}

// Scene content
function SceneContent() {
  const { viewMode, sceneGraph, trajectories } = useStore();

  // Check if we have real data from the API
  const hasRealData = sceneGraph && sceneGraph.objects && sceneGraph.objects.length > 0;

  // Check if we have actual bounds from reconstruction (not default)
  const hasReconstructedBounds = sceneGraph?.bounds &&
    (sceneGraph.bounds.min[0] !== -10 || sceneGraph.bounds.max[0] !== 10);

  // Demo trajectory - suspect path from window to desk to filing cabinet
  const demoTrajectory: [number, number, number][] = [
    [6.5, 0.15, -1.5],    // Entry point (window)
    [5, 0.15, -2],
    [4, 0.15, -3],
    [3, 0.15, -3.5],      // Approach desk
    [3, 0.15, -4],        // At desk
    [1, 0.15, -4],        // Moving along
    [-1, 0.15, -4],
    [-3, 0.15, -4.5],     // Filing cabinet area
    [-4, 0.15, -4.5],     // At filing cabinet
    [-4, 0.15, -3],
    [-4, 0.15, -5.5],     // Exit via door
  ];

  // Key location annotations
  const demoAnnotations = [
    { position: [6.5, 2, -1.5] as [number, number, number], label: 'Entry: Broken Window', color: '#ef4444' },
    { position: [3, 1.8, -4.5] as [number, number, number], label: 'Desk - Files Accessed', color: '#3b82f6' },
    { position: [-4, 2, -4.5] as [number, number, number], label: 'Filing Cabinet - Opened', color: '#f59e0b' },
    { position: [-4, 1.5, -5.5] as [number, number, number], label: 'Exit: Back Door', color: '#22c55e' },
  ];

  // Convert API trajectories to path points
  const apiTrajectoryPaths = useMemo(() => {
    if (!trajectories || trajectories.length === 0) return [];

    return trajectories.map((traj) => {
      const points: [number, number, number][] = [];
      traj.segments.forEach((seg) => {
        if (points.length === 0) {
          points.push(seg.from_position);
        }
        if (seg.waypoints) {
          seg.waypoints.forEach((wp) => points.push(wp));
        }
        points.push(seg.to_position);
      });
      return {
        id: traj.id,
        points,
        confidence: traj.overall_confidence,
        rank: traj.rank,
      };
    });
  }, [trajectories]);

  return (
    <>
      {/* Lighting */}
      <ambientLight intensity={0.5} />
      <directionalLight position={[10, 10, 5]} intensity={0.6} castShadow />
      <directionalLight position={[-5, 8, -5]} intensity={0.3} />
      <pointLight position={[2, 3, -3]} intensity={0.2} color="#3b82f6" />

      {/* Scene Content - Real data vs Demo */}
      {hasRealData ? (
        <>
          {/* Real API Data */}
          {hasReconstructedBounds ? (
            /* Reconstructed room from actual 3D reconstruction */
            <ReconstructedRoom bounds={sceneGraph.bounds} />
          ) : (
            /* Placeholder floor when only scene analysis is complete (waiting for reconstruction) */
            <>
              <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, 0, 0]} receiveShadow>
                <planeGeometry args={[20, 20]} />
                <meshStandardMaterial color="#3a3a3d" roughness={0.9} />
              </mesh>
              <gridHelper args={[20, 40, '#4a4a50', '#3a3a40']} position={[0, 0.01, 0]} />
              {/* Processing indicator */}
              <Html position={[0, 2, 0]} center style={{ pointerEvents: 'none' }}>
                <div className="bg-[#1a1a1f]/90 border border-[#3b82f6]/30 px-4 py-2 rounded-lg text-sm text-[#8b8b96]">
                  <div className="flex items-center gap-2">
                    <div className="w-2 h-2 bg-[#3b82f6] rounded-full animate-pulse" />
                    <span>Room reconstruction in progress...</span>
                  </div>
                  <div className="text-xs text-[#606068] mt-1">
                    Objects positioned from scene analysis
                  </div>
                </div>
              </Html>
            </>
          )}

          {/* Detected objects from scene analysis */}
          <DynamicSceneObjects sceneGraph={sceneGraph} />
          {sceneGraph.evidence && <DynamicEvidenceAnnotations evidence={sceneGraph.evidence} />}

          {/* API Trajectories */}
          {apiTrajectoryPaths.map((traj, i) => (
            <TrajectoryPath
              key={traj.id}
              points={traj.points}
              color={traj.rank === 1 ? '#8b5cf6' : i % 2 === 0 ? '#3b82f6' : '#22c55e'}
              isSelected={traj.rank === 1}
            />
          ))}
        </>
      ) : (
        <>
          {/* Demo Scene (only shown when no API data) */}
          <RoomGeometry />
          <OfficeFurniture />
          <EvidenceMarkers />
          <PointCloud />
        </>
      )}

      {/* Mode-specific visualizations */}
      {viewMode === 'evidence' && !hasRealData && (
        <>
          <TrajectoryPath
            points={demoTrajectory}
            color="#8b5cf6"
            isSelected={true}
          />
          <PersonMarker position={[3, 0, -3]} rotation={-Math.PI / 4} />
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

      {viewMode === 'simulation' && (
        <>
          <TrajectoryVisualization />
          {!hasRealData && <PersonMarker position={[3, 0, -3]} rotation={-Math.PI / 4} />}
          {!hasRealData && demoAnnotations.map((ann, i) => (
            <Annotation
              key={i}
              position={ann.position}
              label={ann.label}
              color={ann.color}
            />
          ))}
        </>
      )}

      {viewMode === 'reasoning' && (
        <>
          {!hasRealData && (
            <TrajectoryPath
              points={demoTrajectory}
              color="#8b5cf6"
              isSelected={true}
            />
          )}
          <TrajectoryVisualization />
          <DiscrepancyHighlighter />
          {!hasRealData && <PersonMarker position={[3, 0, -3]} rotation={-Math.PI / 4} />}
          {!hasRealData && demoAnnotations.map((ann, i) => (
            <Annotation
              key={i}
              position={ann.position}
              label={ann.label}
              color={ann.color}
            />
          ))}
        </>
      )}

      {/* Data source indicator */}
      <Html position={[0, 0.1, 0]} center style={{ pointerEvents: 'none' }}>
        <div className="text-[10px] text-[#606068] bg-[#0a0a0c]/80 px-2 py-0.5 rounded">
          {hasRealData
            ? hasReconstructedBounds
              ? `${sceneGraph.objects.length} objects • Reconstructed room`
              : `${sceneGraph.objects.length} objects • Awaiting reconstruction`
            : 'Demo Scene'}
        </div>
      </Html>
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
          shadows
        >
          <PerspectiveCamera
            makeDefault
            position={[10, 8, 10]}
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
          <span>Office Scene</span>
          <span>Objects: 8</span>
          <span>Evidence: 3</span>
        </div>
      </div>
    </div>
  );
}
