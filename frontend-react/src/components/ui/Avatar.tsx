import { type HTMLAttributes } from 'react';
import { cn } from '@/lib/utils/cn';
import { User } from 'lucide-react';

export type AvatarSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl';

export interface AvatarProps extends HTMLAttributes<HTMLDivElement> {
  src?: string | null;
  alt?: string;
  name?: string;
  size?: AvatarSize;
  showStatus?: boolean;
  status?: 'online' | 'offline' | 'busy' | 'away';
}

function getInitials(name: string): string {
  return name
    .split(' ')
    .map((part) => part[0])
    .join('')
    .toUpperCase()
    .slice(0, 2);
}

export function Avatar({
  className,
  src,
  alt,
  name,
  size = 'md',
  showStatus = false,
  status = 'offline',
  ...props
}: AvatarProps) {
  const sizes: Record<AvatarSize, { container: string; text: string; icon: string; status: string }> = {
    xs: { container: 'h-6 w-6', text: 'text-xs', icon: 'h-3 w-3', status: 'h-1.5 w-1.5 border' },
    sm: { container: 'h-8 w-8', text: 'text-sm', icon: 'h-4 w-4', status: 'h-2 w-2 border' },
    md: { container: 'h-10 w-10', text: 'text-base', icon: 'h-5 w-5', status: 'h-2.5 w-2.5 border-2' },
    lg: { container: 'h-12 w-12', text: 'text-lg', icon: 'h-6 w-6', status: 'h-3 w-3 border-2' },
    xl: { container: 'h-16 w-16', text: 'text-xl', icon: 'h-8 w-8', status: 'h-4 w-4 border-2' },
  };

  const statusColors: Record<string, string> = {
    online: 'bg-green-500',
    offline: 'bg-gray-400',
    busy: 'bg-red-500',
    away: 'bg-yellow-500',
  };

  const sizeConfig = sizes[size];

  return (
    <div className={cn('relative inline-block', className)} {...props}>
      <div
        className={cn(
          'rounded-full overflow-hidden bg-gray-200 flex items-center justify-center',
          sizeConfig.container
        )}
      >
        {src ? (
          <img
            src={src}
            alt={alt || name || 'Avatar'}
            className="h-full w-full object-cover"
          />
        ) : name ? (
          <span className={cn('font-medium text-gray-600', sizeConfig.text)}>
            {getInitials(name)}
          </span>
        ) : (
          <User className={cn('text-gray-400', sizeConfig.icon)} />
        )}
      </div>
      {showStatus && (
        <span
          className={cn(
            'absolute bottom-0 right-0 rounded-full border-white',
            statusColors[status],
            sizeConfig.status
          )}
        />
      )}
    </div>
  );
}

// Avatar Group
export interface AvatarGroupProps extends HTMLAttributes<HTMLDivElement> {
  max?: number;
  size?: AvatarSize;
  children: React.ReactNode;
}

export function AvatarGroup({
  className,
  max = 4,
  size = 'md',
  children,
  ...props
}: AvatarGroupProps) {
  const childArray = Array.isArray(children) ? children : [children];
  const visible = childArray.slice(0, max);
  const remaining = childArray.length - max;

  const sizes: Record<AvatarSize, string> = {
    xs: 'h-6 w-6 text-xs',
    sm: 'h-8 w-8 text-sm',
    md: 'h-10 w-10 text-base',
    lg: 'h-12 w-12 text-lg',
    xl: 'h-16 w-16 text-xl',
  };

  return (
    <div className={cn('flex -space-x-2', className)} {...props}>
      {visible.map((child, index) => (
        <div key={index} className="ring-2 ring-white rounded-full">
          {child}
        </div>
      ))}
      {remaining > 0 && (
        <div
          className={cn(
            'flex items-center justify-center rounded-full bg-gray-100 text-gray-600 font-medium ring-2 ring-white',
            sizes[size]
          )}
        >
          +{remaining}
        </div>
      )}
    </div>
  );
}
