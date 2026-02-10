'use client';

import {
  Home,
  Folder,
  MessageSquare,
  Users,
  Brain,
  Settings,
  Infinity,
} from 'lucide-react';
import { cn } from '@/lib/utils';
import { useStore, type SidebarTab } from '@/lib/store';

// Navigation items matching the design
const navItems: { id: SidebarTab; label: string; icon: React.ComponentType<{ className?: string }> }[] = [
  { id: 'home', label: 'Overview', icon: Home },
  { id: 'evidence', label: 'Evidence', icon: Folder },
  { id: 'witness', label: 'Witnesses', icon: MessageSquare },
  { id: 'suspects', label: 'Suspects', icon: Users },
  { id: 'reasoning', label: 'Reasoning', icon: Brain },
  { id: 'settings', label: 'Settings', icon: Settings },
];

export function NavigationRail() {
  const { activeSidebarTab, setActiveSidebarTab } = useStore();

  return (
    <div className="h-screen w-16 bg-[#27272A] flex flex-col items-center py-4 gap-6 shrink-0 z-50 rounded-r-2xl relative">
      {/* Logo */}
      <div className="w-10 h-10 flex items-center justify-center text-[#f0f0f2] shrink-0">
        <Infinity className="w-8 h-8" />
      </div>

      {/* Nav Items */}
      <div className="flex flex-col items-center gap-2">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = activeSidebarTab === item.id;
          return (
            <button
              key={item.id}
              onClick={() => setActiveSidebarTab(item.id)}
              className={cn(
                'w-10 h-10 rounded-xl flex items-center justify-center transition-all duration-200 group relative',
                isActive
                  ? 'bg-white/10 text-[#f0f0f2]'
                  : 'text-[#a0a0a8] hover:text-[#f0f0f2] hover:bg-white/5'
              )}
              title={item.label}
            >
              <Icon className="w-5 h-5" />

              {/* Tooltip */}
              <div className="absolute left-14 px-2 py-1 bg-[#1f1f24] text-white text-xs rounded opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none whitespace-nowrap z-50 border border-[#2a2a32]">
                {item.label}
              </div>
            </button>
          );
        })}
      </div>

      {/* Bottom Avatar */}
      <div className="mt-auto pb-4">
        <div className="w-10 h-10 rounded-full bg-gradient-to-br from-purple-500 to-blue-500 flex items-center justify-center text-white font-bold text-sm shadow-lg border-2 border-[#1c1c1f] overflow-hidden">
          <img
            src="https://avatars.githubusercontent.com/u/124599?v=4"
            alt="User"
            className="w-full h-full object-cover"
          />
        </div>
      </div>
    </div>
  );
}
