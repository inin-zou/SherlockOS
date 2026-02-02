'use client';

import { useRef, useMemo, useState } from 'react';
import { useFrame } from '@react-three/fiber';
import { Html, Line } from '@react-three/drei';
import * as THREE from 'three';
import { useStore } from '@/lib/store';

interface Discrepancy {
  id: string;
  type: 'timeline_conflict' | 'line_of_sight' | 'physical_impossible' | 'testimony_mismatch';
  severity: 'low' | 'medium' | 'high';
  position: [number, number, number];
  targetPosition?: [number, number, number];
  description: string;
  witnessSource: string;
  contradictingEvidence: string;
}

interface DiscrepancyHighlighterProps {
  discrepancies?: Discrepancy[];
  showAll?: boolean;
}

const SEVERITY_COLORS = {
  high: '#ef4444',
  medium: '#f59e0b',
  low: '#22c55e',
};

const TYPE_ICONS = {
  timeline_conflict: '\u23f0', // alarm clock
  line_of_sight: '\ud83d\udc41', // eye
  physical_impossible: '\u26a0', // warning
  testimony_mismatch: '\ud83d\udde3', // speaking head
};

/**
 * 3D visualization of discrepancies between witness statements and evidence
 * Shows line-of-sight blocks, timeline conflicts, and other inconsistencies
 */
export function DiscrepancyHighlighter({
  discrepancies: propDiscrepancies,
  showAll = true,
}: DiscrepancyHighlighterProps) {
  const { viewMode, sceneGraph } = useStore();

  // Use demo discrepancies if none provided
  const discrepancies = useMemo(() => {
    if (propDiscrepancies) return propDiscrepancies;

    // Generate demo discrepancies based on scene objects
    const demoDiscrepancies: Discrepancy[] = [
      {
        id: 'd1',
        type: 'line_of_sight',
        severity: 'high',
        position: [2, 1, 3],
        targetPosition: [5, 1, -2],
        description: 'Wall blocks line of sight from lobby to vault entrance',
        witnessSource: 'Witness B',
        contradictingEvidence: '3D scene geometry',
      },
      {
        id: 'd2',
        type: 'timeline_conflict',
        severity: 'medium',
        position: [-1, 1, 4],
        description: 'Witness claims suspect at gate at 22:10, but CCTV shows gate empty',
        witnessSource: 'Witness A',
        contradictingEvidence: 'CCTV-CAM-02',
      },
      {
        id: 'd3',
        type: 'testimony_mismatch',
        severity: 'low',
        position: [3, 1, 1],
        description: 'Height estimate differs by 15cm between witnesses',
        witnessSource: 'Witness A & B',
        contradictingEvidence: 'Footprint analysis',
      },
    ];

    return demoDiscrepancies;
  }, [propDiscrepancies]);

  // Only show in reasoning mode
  if (viewMode !== 'reasoning') {
    return null;
  }

  return (
    <group name="discrepancies">
      {discrepancies.map((discrepancy) => (
        <DiscrepancyMarker
          key={discrepancy.id}
          discrepancy={discrepancy}
          showDetails={showAll}
        />
      ))}
    </group>
  );
}

interface DiscrepancyMarkerProps {
  discrepancy: Discrepancy;
  showDetails: boolean;
}

function DiscrepancyMarker({ discrepancy, showDetails }: DiscrepancyMarkerProps) {
  const groupRef = useRef<THREE.Group>(null);
  const [isHovered, setIsHovered] = useState(false);
  const [pulseScale, setPulseScale] = useState(1);

  const color = SEVERITY_COLORS[discrepancy.severity];

  // Pulse animation
  useFrame((state) => {
    const pulse = 1 + Math.sin(state.clock.elapsedTime * 3) * 0.15;
    setPulseScale(pulse);
  });

  return (
    <group
      ref={groupRef}
      position={discrepancy.position}
      onPointerOver={() => setIsHovered(true)}
      onPointerOut={() => setIsHovered(false)}
    >
      {/* Main marker - pulsing ring */}
      <mesh scale={[pulseScale, 1, pulseScale]}>
        <torusGeometry args={[0.4, 0.05, 16, 32]} />
        <meshStandardMaterial
          color={color}
          emissive={color}
          emissiveIntensity={0.5}
          transparent
          opacity={0.8}
        />
      </mesh>

      {/* Inner sphere */}
      <mesh>
        <sphereGeometry args={[0.15, 16, 16]} />
        <meshStandardMaterial
          color={color}
          emissive={color}
          emissiveIntensity={isHovered ? 1 : 0.3}
          transparent
          opacity={0.9}
        />
      </mesh>

      {/* Vertical beam */}
      <mesh position={[0, 1.5, 0]}>
        <cylinderGeometry args={[0.02, 0.02, 3, 8]} />
        <meshStandardMaterial
          color={color}
          transparent
          opacity={0.4}
        />
      </mesh>

      {/* Warning icon at top */}
      <Html position={[0, 3.2, 0]} center style={{ pointerEvents: 'none' }}>
        <div
          style={{
            width: '32px',
            height: '32px',
            borderRadius: '50%',
            background: `${color}20`,
            border: `2px solid ${color}`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            fontSize: '16px',
            animation: 'pulse 2s infinite',
          }}
        >
          {TYPE_ICONS[discrepancy.type]}
        </div>
      </Html>

      {/* Line of sight blocker visualization */}
      {discrepancy.type === 'line_of_sight' && discrepancy.targetPosition && (
        <LineOfSightBlocker
          from={discrepancy.position}
          to={discrepancy.targetPosition}
          color={color}
        />
      )}

      {/* Detail tooltip */}
      {(showDetails || isHovered) && (
        <Html
          position={[0.8, 0.5, 0]}
          style={{ pointerEvents: isHovered ? 'auto' : 'none' }}
        >
          <div
            style={{
              background: 'rgba(17, 17, 20, 0.95)',
              padding: '12px',
              borderRadius: '8px',
              border: `1px solid ${color}40`,
              minWidth: '200px',
              maxWidth: '280px',
              boxShadow: '0 4px 20px rgba(0,0,0,0.5)',
              transform: isHovered ? 'scale(1.05)' : 'scale(1)',
              transition: 'transform 0.2s',
            }}
          >
            {/* Header */}
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                marginBottom: '8px',
              }}
            >
              <span
                style={{
                  width: '8px',
                  height: '8px',
                  borderRadius: '50%',
                  background: color,
                }}
              />
              <span
                style={{
                  fontSize: '11px',
                  fontWeight: 600,
                  color: color,
                  textTransform: 'uppercase',
                  letterSpacing: '0.5px',
                }}
              >
                {discrepancy.type.replace('_', ' ')}
              </span>
              <span
                style={{
                  marginLeft: 'auto',
                  fontSize: '10px',
                  padding: '2px 6px',
                  borderRadius: '4px',
                  background: `${color}20`,
                  color: color,
                }}
              >
                {discrepancy.severity}
              </span>
            </div>

            {/* Description */}
            <p
              style={{
                fontSize: '12px',
                color: '#f0f0f2',
                margin: '0 0 8px 0',
                lineHeight: 1.4,
              }}
            >
              {discrepancy.description}
            </p>

            {/* Sources */}
            <div style={{ fontSize: '11px', color: '#a0a0a8' }}>
              <div style={{ marginBottom: '4px' }}>
                <span style={{ color: '#8b5cf6' }}>Tier 3:</span>{' '}
                {discrepancy.witnessSource}
              </div>
              <div>
                <span style={{ color: '#3b82f6' }}>vs:</span>{' '}
                {discrepancy.contradictingEvidence}
              </div>
            </div>
          </div>
        </Html>
      )}
    </group>
  );
}

interface LineOfSightBlockerProps {
  from: [number, number, number];
  to: [number, number, number];
  color: string;
}

function LineOfSightBlocker({ from, to, color }: LineOfSightBlockerProps) {
  const [dashOffset, setDashOffset] = useState(0);

  // Animate dash offset
  useFrame((state, delta) => {
    setDashOffset((prev) => (prev + delta * 0.5) % 1);
  });

  // Calculate midpoint for the "X" blocker
  const midpoint: [number, number, number] = [
    (from[0] + to[0]) / 2,
    (from[1] + to[1]) / 2,
    (from[2] + to[2]) / 2,
  ];

  return (
    <group>
      {/* Dashed line from witness position to target */}
      <Line
        points={[from, to]}
        color={color}
        lineWidth={2}
        dashed
        dashSize={0.3}
        dashScale={1}
        gapSize={0.2}
        transparent
        opacity={0.5}
      />

      {/* Blocker "X" at midpoint */}
      <group position={midpoint}>
        {/* X mark */}
        <Line
          points={[
            [-0.3, -0.3, 0],
            [0.3, 0.3, 0],
          ]}
          color={color}
          lineWidth={3}
        />
        <Line
          points={[
            [-0.3, 0.3, 0],
            [0.3, -0.3, 0],
          ]}
          color={color}
          lineWidth={3}
        />

        {/* Blocker ring */}
        <mesh rotation={[Math.PI / 2, 0, 0]}>
          <torusGeometry args={[0.5, 0.03, 16, 32]} />
          <meshStandardMaterial
            color={color}
            emissive={color}
            emissiveIntensity={0.3}
            transparent
            opacity={0.7}
          />
        </mesh>
      </group>

      {/* Target position marker */}
      <group position={to}>
        <mesh>
          <sphereGeometry args={[0.1, 16, 16]} />
          <meshStandardMaterial
            color="#606068"
            transparent
            opacity={0.5}
          />
        </mesh>
      </group>
    </group>
  );
}

/**
 * Ground-level discrepancy zone indicator
 */
export function DiscrepancyZone({
  center,
  radius = 2,
  severity = 'medium',
}: {
  center: [number, number, number];
  radius?: number;
  severity?: 'low' | 'medium' | 'high';
}) {
  const color = SEVERITY_COLORS[severity];
  const meshRef = useRef<THREE.Mesh>(null);

  useFrame((state) => {
    if (meshRef.current) {
      meshRef.current.rotation.z = state.clock.elapsedTime * 0.2;
    }
  });

  return (
    <group position={center}>
      {/* Ground circle */}
      <mesh ref={meshRef} rotation={[-Math.PI / 2, 0, 0]} position={[0, 0.01, 0]}>
        <ringGeometry args={[radius * 0.8, radius, 32]} />
        <meshStandardMaterial
          color={color}
          transparent
          opacity={0.2}
          side={THREE.DoubleSide}
        />
      </mesh>

      {/* Dashed border */}
      <Line
        points={Array.from({ length: 65 }, (_, i) => {
          const angle = (i / 64) * Math.PI * 2;
          return [Math.cos(angle) * radius, 0.02, Math.sin(angle) * radius];
        })}
        color={color}
        lineWidth={2}
        dashed
        dashSize={0.2}
        gapSize={0.1}
      />
    </group>
  );
}
