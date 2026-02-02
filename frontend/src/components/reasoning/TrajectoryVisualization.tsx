'use client';

import { useRef, useMemo, useState, useEffect } from 'react';
import { useFrame } from '@react-three/fiber';
import { Line, Text, Html } from '@react-three/drei';
import * as THREE from 'three';
import { useStore } from '@/lib/store';
import type { Trajectory, TrajectorySegment } from '@/lib/types';

interface TrajectoryVisualizationProps {
  animated?: boolean;
  showLabels?: boolean;
  opacity?: number;
}

/**
 * 3D visualization of trajectory hypotheses
 * Renders animated paths with waypoints and confidence indicators
 */
export function TrajectoryVisualization({
  animated = true,
  showLabels = true,
  opacity = 0.8,
}: TrajectoryVisualizationProps) {
  const { trajectories, selectedTrajectoryId, viewMode } = useStore();

  // Only show in reasoning mode
  if (viewMode !== 'reasoning' || trajectories.length === 0) {
    return null;
  }

  return (
    <group name="trajectories">
      {trajectories.map((trajectory, index) => (
        <TrajectoryPath
          key={trajectory.id}
          trajectory={trajectory}
          isSelected={trajectory.id === selectedTrajectoryId}
          colorIndex={index}
          animated={animated}
          showLabels={showLabels}
          opacity={opacity}
        />
      ))}
    </group>
  );
}

interface TrajectoryPathProps {
  trajectory: Trajectory;
  isSelected: boolean;
  colorIndex: number;
  animated: boolean;
  showLabels: boolean;
  opacity: number;
}

const TRAJECTORY_COLORS = [
  '#6366f1', // indigo
  '#8b5cf6', // purple
  '#ec4899', // pink
  '#f59e0b', // amber
  '#10b981', // emerald
];

function TrajectoryPath({
  trajectory,
  isSelected,
  colorIndex,
  animated,
  showLabels,
  opacity,
}: TrajectoryPathProps) {
  const groupRef = useRef<THREE.Group>(null);
  const [animationProgress, setAnimationProgress] = useState(0);

  const color = TRAJECTORY_COLORS[colorIndex % TRAJECTORY_COLORS.length];
  const lineWidth = isSelected ? 3 : 1.5;
  const finalOpacity = isSelected ? opacity : opacity * 0.5;

  // Build path points from segments
  const pathPoints = useMemo(() => {
    const points: THREE.Vector3[] = [];

    trajectory.segments.forEach((segment, i) => {
      if (i === 0) {
        points.push(new THREE.Vector3(...segment.from_position));
      }

      // Add waypoints if available
      if (segment.waypoints) {
        segment.waypoints.forEach((wp) => {
          points.push(new THREE.Vector3(...wp));
        });
      }

      points.push(new THREE.Vector3(...segment.to_position));
    });

    return points;
  }, [trajectory.segments]);

  // Animation
  useFrame((state, delta) => {
    if (animated && isSelected) {
      setAnimationProgress((prev) => {
        const next = prev + delta * 0.3;
        return next > 1 ? 0 : next;
      });
    }
  });

  // Get animated subset of points
  const visiblePoints = useMemo(() => {
    if (!animated || !isSelected) return pathPoints;

    const numPoints = Math.max(2, Math.floor(pathPoints.length * animationProgress));
    return pathPoints.slice(0, numPoints);
  }, [pathPoints, animationProgress, animated, isSelected]);

  // Create curve for smooth path
  const curve = useMemo(() => {
    if (pathPoints.length < 2) return null;
    return new THREE.CatmullRomCurve3(pathPoints, false, 'centripetal', 0.5);
  }, [pathPoints]);

  const smoothPoints = useMemo(() => {
    if (!curve) return [];
    return curve.getPoints(50);
  }, [curve]);

  if (pathPoints.length < 2) return null;

  return (
    <group ref={groupRef} name={`trajectory-${trajectory.id}`}>
      {/* Main path line */}
      <Line
        points={animated && isSelected ? visiblePoints : smoothPoints}
        color={color}
        lineWidth={lineWidth}
        transparent
        opacity={finalOpacity}
        dashed={!isSelected}
        dashSize={0.3}
        dashScale={1}
        gapSize={0.1}
      />

      {/* Waypoint markers */}
      {pathPoints.map((point, i) => (
        <group key={i} position={point}>
          {/* Waypoint sphere */}
          <mesh>
            <sphereGeometry args={[isSelected ? 0.15 : 0.1, 16, 16]} />
            <meshStandardMaterial
              color={color}
              transparent
              opacity={finalOpacity}
              emissive={color}
              emissiveIntensity={isSelected ? 0.5 : 0.2}
            />
          </mesh>

          {/* Index label */}
          {showLabels && isSelected && (
            <Html
              position={[0, 0.3, 0]}
              center
              style={{
                pointerEvents: 'none',
              }}
            >
              <div
                style={{
                  background: 'rgba(17, 17, 20, 0.9)',
                  padding: '2px 6px',
                  borderRadius: '4px',
                  fontSize: '10px',
                  color: color,
                  fontFamily: 'monospace',
                  whiteSpace: 'nowrap',
                }}
              >
                {i === 0 ? 'START' : i === pathPoints.length - 1 ? 'END' : `P${i}`}
              </div>
            </Html>
          )}
        </group>
      ))}

      {/* Direction arrows along path */}
      {isSelected && smoothPoints.length > 10 && (
        <>
          {[0.25, 0.5, 0.75].map((t) => {
            const index = Math.floor(t * (smoothPoints.length - 1));
            const point = smoothPoints[index];
            const nextPoint = smoothPoints[Math.min(index + 1, smoothPoints.length - 1)];
            const direction = new THREE.Vector3()
              .subVectors(nextPoint, point)
              .normalize();

            return (
              <group key={t} position={point}>
                <mesh rotation={[0, Math.atan2(direction.x, direction.z), 0]}>
                  <coneGeometry args={[0.08, 0.2, 8]} />
                  <meshStandardMaterial
                    color={color}
                    transparent
                    opacity={finalOpacity}
                  />
                </mesh>
              </group>
            );
          })}
        </>
      )}

      {/* Confidence indicator at start */}
      {showLabels && isSelected && pathPoints.length > 0 && (
        <Html
          position={[pathPoints[0].x, pathPoints[0].y + 0.6, pathPoints[0].z]}
          center
          style={{ pointerEvents: 'none' }}
        >
          <div
            style={{
              background: 'rgba(17, 17, 20, 0.95)',
              padding: '4px 8px',
              borderRadius: '6px',
              border: `1px solid ${color}40`,
              fontSize: '11px',
              color: '#f0f0f2',
              fontFamily: 'system-ui',
            }}
          >
            <div style={{ color, fontWeight: 600, marginBottom: '2px' }}>
              Hypothesis #{trajectory.rank}
            </div>
            <div style={{ color: '#a0a0a8', fontSize: '10px' }}>
              {Math.round(trajectory.overall_confidence * 100)}% confidence
            </div>
          </div>
        </Html>
      )}
    </group>
  );
}

/**
 * Animated moving marker along selected trajectory
 */
export function TrajectoryMarker() {
  const { trajectories, selectedTrajectoryId, isPlaying, currentTime } = useStore();
  const meshRef = useRef<THREE.Mesh>(null);

  const selectedTrajectory = trajectories.find((t) => t.id === selectedTrajectoryId);

  // Build path
  const pathPoints = useMemo(() => {
    if (!selectedTrajectory) return [];

    const points: THREE.Vector3[] = [];
    selectedTrajectory.segments.forEach((segment, i) => {
      if (i === 0) {
        points.push(new THREE.Vector3(...segment.from_position));
      }
      if (segment.waypoints) {
        segment.waypoints.forEach((wp) => {
          points.push(new THREE.Vector3(...wp));
        });
      }
      points.push(new THREE.Vector3(...segment.to_position));
    });
    return points;
  }, [selectedTrajectory]);

  const curve = useMemo(() => {
    if (pathPoints.length < 2) return null;
    return new THREE.CatmullRomCurve3(pathPoints, false, 'centripetal', 0.5);
  }, [pathPoints]);

  useFrame((state) => {
    if (!meshRef.current || !curve || !isPlaying) return;

    // Use sine wave for smooth back-and-forth motion
    const t = (Math.sin(state.clock.elapsedTime * 0.5) + 1) / 2;
    const point = curve.getPoint(t);

    meshRef.current.position.copy(point);
    meshRef.current.position.y += 0.3; // Float above path
  });

  if (!selectedTrajectory || pathPoints.length < 2) return null;

  return (
    <mesh ref={meshRef}>
      <sphereGeometry args={[0.2, 32, 32]} />
      <meshStandardMaterial
        color="#f59e0b"
        emissive="#f59e0b"
        emissiveIntensity={0.8}
        transparent
        opacity={0.9}
      />
    </mesh>
  );
}
