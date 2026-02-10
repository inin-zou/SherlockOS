'use client';

import { Suspense } from 'react';
import { Canvas } from '@react-three/fiber';
import { OrbitControls, PerspectiveCamera } from '@react-three/drei';
import { GaussianSplatRenderer } from '@/components/scene/GaussianSplatRenderer';

function LoadingFallback() {
  return (
    <div className="absolute inset-0 flex items-center justify-center text-white">
      <div className="flex flex-col items-center gap-3">
        <div className="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
        <span className="text-sm text-gray-400">Loading Gaussian Splat test scene...</span>
      </div>
    </div>
  );
}

export default function TestSplatPage() {
  return (
    <div className="w-screen h-screen bg-black relative">
      {/* Info overlay */}
      <div className="absolute top-4 left-4 z-10 bg-black/80 border border-gray-700 rounded-lg px-4 py-3 text-white text-sm">
        <h1 className="font-bold text-lg mb-1">3D Scene Reconstruction</h1>
        <p className="text-gray-400">
          Loading <code className="text-blue-400">/scene.splat</code> (500K gaussians, ~15MB)
        </p>
        <p className="text-gray-500 text-xs mt-1">
          Drag to rotate | Scroll to zoom | Right-click to pan
        </p>
      </div>

      <Suspense fallback={<LoadingFallback />}>
        <Canvas
          gl={{ antialias: true, alpha: true }}
          dpr={[1, 2]}
        >
          <PerspectiveCamera makeDefault position={[0, 2, 5]} fov={50} />
          <OrbitControls
            enablePan
            enableZoom
            enableRotate
            minDistance={1}
            maxDistance={20}
          />
          <ambientLight intensity={0.5} />
          <directionalLight position={[5, 5, 5]} intensity={0.5} />
          <GaussianSplatRenderer src="/scene.splat" />
        </Canvas>
      </Suspense>
    </div>
  );
}
