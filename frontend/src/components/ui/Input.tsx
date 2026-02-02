'use client';

import { forwardRef, type InputHTMLAttributes } from 'react';
import { cn } from '@/lib/utils';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  icon?: React.ReactNode;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, icon, ...props }, ref) => {
    return (
      <div className="relative">
        {icon && (
          <div className="absolute left-3 top-1/2 -translate-y-1/2 text-[#606068]">
            {icon}
          </div>
        )}
        <input
          ref={ref}
          className={cn(
            'w-full h-9 bg-[#111114] border border-[#2a2a32] rounded-lg',
            'text-sm text-[#f0f0f2] placeholder:text-[#606068]',
            'focus:outline-none focus:ring-2 focus:ring-[#3b82f6] focus:border-transparent',
            'transition-all duration-150',
            icon ? 'pl-10 pr-4' : 'px-4',
            className
          )}
          {...props}
        />
      </div>
    );
  }
);

Input.displayName = 'Input';
