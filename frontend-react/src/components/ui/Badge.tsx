import { type HTMLAttributes } from 'react';
import { cn } from '@/lib/utils/cn';

export type BadgeVariant =
  | 'default'
  | 'primary'
  | 'success'
  | 'warning'
  | 'error'
  | 'info'
  | 'outline';

export type BadgeSize = 'sm' | 'md' | 'lg';

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant;
  size?: BadgeSize;
}

export function Badge({
  className,
  variant = 'default',
  size = 'md',
  children,
  ...props
}: BadgeProps) {
  const variants: Record<BadgeVariant, string> = {
    default: 'bg-gray-100 text-gray-700',
    primary: 'bg-ekf-red text-white',
    success: 'bg-green-100 text-green-800',
    warning: 'bg-yellow-100 text-yellow-800',
    error: 'bg-red-100 text-red-800',
    info: 'bg-blue-100 text-blue-800',
    outline: 'bg-transparent border border-gray-300 text-gray-700',
  };

  const sizes: Record<BadgeSize, string> = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-2.5 py-0.5 text-sm',
    lg: 'px-3 py-1 text-sm',
  };

  return (
    <span
      className={cn(
        'inline-flex items-center font-medium rounded-full',
        variants[variant],
        sizes[size],
        className
      )}
      {...props}
    >
      {children}
    </span>
  );
}

// Status Badge for tasks/tickets
export type StatusType = 'open' | 'in_progress' | 'done' | 'blocked' | 'review' | 'closed';

export interface StatusBadgeProps extends Omit<BadgeProps, 'variant'> {
  status: StatusType;
}

export function StatusBadge({ status, className, ...props }: StatusBadgeProps) {
  const statusConfig: Record<StatusType, { label: string; className: string }> = {
    open: { label: 'Открыта', className: 'bg-blue-100 text-blue-800' },
    in_progress: { label: 'В работе', className: 'bg-yellow-100 text-yellow-800' },
    done: { label: 'Готово', className: 'bg-green-100 text-green-800' },
    blocked: { label: 'Заблокирована', className: 'bg-red-100 text-red-800' },
    review: { label: 'На ревью', className: 'bg-purple-100 text-purple-800' },
    closed: { label: 'Закрыта', className: 'bg-gray-100 text-gray-600' },
  };

  const config = statusConfig[status];

  return (
    <Badge className={cn(config.className, className)} {...props}>
      {config.label}
    </Badge>
  );
}

// Priority Badge
export type PriorityType = 'critical' | 'high' | 'medium' | 'low';

export interface PriorityBadgeProps extends Omit<BadgeProps, 'variant'> {
  priority: PriorityType;
}

export function PriorityBadge({ priority, className, ...props }: PriorityBadgeProps) {
  const priorityConfig: Record<PriorityType, { label: string; className: string }> = {
    critical: { label: 'Критический', className: 'bg-red-100 text-red-800' },
    high: { label: 'Высокий', className: 'bg-orange-100 text-orange-800' },
    medium: { label: 'Средний', className: 'bg-yellow-100 text-yellow-800' },
    low: { label: 'Низкий', className: 'bg-green-100 text-green-800' },
  };

  const config = priorityConfig[priority];

  return (
    <Badge className={cn(config.className, className)} {...props}>
      {config.label}
    </Badge>
  );
}
