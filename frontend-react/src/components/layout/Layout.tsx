import { Outlet, useLocation } from 'react-router-dom';
import { GlobalNav } from './GlobalNav';
import { Sidebar } from './Sidebar';
import { cn } from '@/lib/utils/cn';

// Pages that have a sidebar
const pagesWithSidebar = ['/tasks', '/meetings', '/calendar', '/service-desk', '/sprints', '/releases'];

export function Layout() {
  const location = useLocation();
  const hasSidebar = pagesWithSidebar.some((path) => location.pathname.startsWith(path));

  return (
    <div className="min-h-screen bg-gray-50">
      <GlobalNav />
      {hasSidebar && <Sidebar />}
      <main
        className={cn(
          'pt-12 min-h-screen',
          hasSidebar ? 'pl-60' : ''
        )}
      >
        <div className="p-6">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
