import { type HTMLAttributes, type TdHTMLAttributes, type ThHTMLAttributes } from 'react';
import { cn } from '@/lib/utils/cn';
import { ChevronUp, ChevronDown, ChevronsUpDown } from 'lucide-react';

// Table Root
export interface TableProps extends HTMLAttributes<HTMLTableElement> {}

export function Table({ className, children, ...props }: TableProps) {
  return (
    <div className="w-full overflow-x-auto">
      <table
        className={cn('w-full text-sm text-left', className)}
        {...props}
      >
        {children}
      </table>
    </div>
  );
}

// Table Header
export interface TableHeaderProps extends HTMLAttributes<HTMLTableSectionElement> {}

export function TableHeader({ className, children, ...props }: TableHeaderProps) {
  return (
    <thead
      className={cn('bg-gray-50 border-b border-gray-200', className)}
      {...props}
    >
      {children}
    </thead>
  );
}

// Table Body
export interface TableBodyProps extends HTMLAttributes<HTMLTableSectionElement> {}

export function TableBody({ className, children, ...props }: TableBodyProps) {
  return (
    <tbody className={cn('divide-y divide-gray-200', className)} {...props}>
      {children}
    </tbody>
  );
}

// Table Row
export interface TableRowProps extends HTMLAttributes<HTMLTableRowElement> {
  isClickable?: boolean;
  isSelected?: boolean;
}

export function TableRow({ className, isClickable, isSelected, children, ...props }: TableRowProps) {
  return (
    <tr
      className={cn(
        'bg-white',
        isClickable && 'cursor-pointer hover:bg-gray-50',
        isSelected && 'bg-ekf-red/5',
        className
      )}
      {...props}
    >
      {children}
    </tr>
  );
}

// Table Head Cell
export type SortDirection = 'asc' | 'desc' | null;

export interface TableHeadProps extends ThHTMLAttributes<HTMLTableCellElement> {
  sortable?: boolean;
  sortDirection?: SortDirection;
  onSort?: () => void;
}

export function TableHead({
  className,
  sortable,
  sortDirection,
  onSort,
  children,
  ...props
}: TableHeadProps) {
  const SortIcon = () => {
    if (!sortable) return null;
    if (sortDirection === 'asc') return <ChevronUp className="h-4 w-4" />;
    if (sortDirection === 'desc') return <ChevronDown className="h-4 w-4" />;
    return <ChevronsUpDown className="h-4 w-4 text-gray-400" />;
  };

  return (
    <th
      scope="col"
      className={cn(
        'px-4 py-3 text-xs font-semibold text-gray-600 uppercase tracking-wider',
        sortable && 'cursor-pointer select-none hover:bg-gray-100',
        className
      )}
      onClick={sortable ? onSort : undefined}
      {...props}
    >
      <div className="flex items-center gap-1">
        {children}
        <SortIcon />
      </div>
    </th>
  );
}

// Table Cell
export interface TableCellProps extends TdHTMLAttributes<HTMLTableCellElement> {}

export function TableCell({ className, children, ...props }: TableCellProps) {
  return (
    <td
      className={cn('px-4 py-3 text-gray-700', className)}
      {...props}
    >
      {children}
    </td>
  );
}

// Table Footer
export interface TableFooterProps extends HTMLAttributes<HTMLTableSectionElement> {}

export function TableFooter({ className, children, ...props }: TableFooterProps) {
  return (
    <tfoot
      className={cn('bg-gray-50 border-t border-gray-200', className)}
      {...props}
    >
      {children}
    </tfoot>
  );
}

// Empty State
export interface TableEmptyProps extends HTMLAttributes<HTMLTableRowElement> {
  colSpan: number;
  message?: string;
}

export function TableEmpty({
  className,
  colSpan,
  message = 'Нет данных для отображения',
  ...props
}: TableEmptyProps) {
  return (
    <tr className={className} {...props}>
      <td colSpan={colSpan} className="px-4 py-12 text-center text-gray-500">
        {message}
      </td>
    </tr>
  );
}

// Loading State
export interface TableLoadingProps extends HTMLAttributes<HTMLTableRowElement> {
  colSpan: number;
  rows?: number;
}

export function TableLoading({
  className,
  colSpan,
  rows = 5,
  ...props
}: TableLoadingProps) {
  return (
    <>
      {Array.from({ length: rows }).map((_, i) => (
        <tr key={i} className={cn('animate-pulse', className)} {...props}>
          {Array.from({ length: colSpan }).map((_, j) => (
            <td key={j} className="px-4 py-3">
              <div className="h-4 bg-gray-200 rounded w-3/4" />
            </td>
          ))}
        </tr>
      ))}
    </>
  );
}
