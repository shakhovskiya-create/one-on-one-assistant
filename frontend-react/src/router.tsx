import { createBrowserRouter, Navigate } from 'react-router-dom';
import { Layout } from '@/components/layout/Layout';

// Lazy load pages
import { lazy, Suspense } from 'react';

// Loading component
function PageLoader() {
  return (
    <div className="flex items-center justify-center min-h-[400px]">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-ekf-red"></div>
    </div>
  );
}

// Wrap lazy components with Suspense
function lazyLoad(Component: React.LazyExoticComponent<React.ComponentType>) {
  return (
    <Suspense fallback={<PageLoader />}>
      <Component />
    </Suspense>
  );
}

// Pages - will be created later
const DashboardPage = lazy(() => import('@/pages/Dashboard'));
const LoginPage = lazy(() => import('@/pages/Login'));
const TasksPage = lazy(() => import('@/pages/Tasks'));
const EmployeesPage = lazy(() => import('@/pages/Employees'));
const MeetingsPage = lazy(() => import('@/pages/Meetings'));
const CalendarPage = lazy(() => import('@/pages/Calendar'));
const MailPage = lazy(() => import('@/pages/Mail'));
const MessengerPage = lazy(() => import('@/pages/Messenger'));
const ServiceDeskPage = lazy(() => import('@/pages/ServiceDesk'));
const AnalyticsPage = lazy(() => import('@/pages/Analytics'));
const ProfilePage = lazy(() => import('@/pages/Profile'));
const SprintsPage = lazy(() => import('@/pages/Sprints'));
const ReleasesPage = lazy(() => import('@/pages/Releases'));
const ConfluencePage = lazy(() => import('@/pages/Confluence'));

export const router = createBrowserRouter([
  {
    path: '/login',
    element: lazyLoad(LoginPage),
  },
  {
    path: '/',
    element: <Layout />,
    children: [
      {
        index: true,
        element: lazyLoad(DashboardPage),
      },
      {
        path: 'employees',
        element: lazyLoad(EmployeesPage),
      },
      {
        path: 'employees/:id',
        element: lazyLoad(EmployeesPage),
      },
      {
        path: 'tasks',
        element: lazyLoad(TasksPage),
      },
      {
        path: 'tasks/:id',
        element: lazyLoad(TasksPage),
      },
      {
        path: 'meetings',
        element: lazyLoad(MeetingsPage),
      },
      {
        path: 'meetings/:id',
        element: lazyLoad(MeetingsPage),
      },
      {
        path: 'calendar',
        element: lazyLoad(CalendarPage),
      },
      {
        path: 'mail',
        element: lazyLoad(MailPage),
      },
      {
        path: 'mail/:id',
        element: lazyLoad(MailPage),
      },
      {
        path: 'messenger',
        element: lazyLoad(MessengerPage),
      },
      {
        path: 'service-desk',
        element: lazyLoad(ServiceDeskPage),
      },
      {
        path: 'service-desk/create',
        element: lazyLoad(ServiceDeskPage),
      },
      {
        path: 'service-desk/tickets/:id',
        element: lazyLoad(ServiceDeskPage),
      },
      {
        path: 'service-desk/agent',
        element: lazyLoad(ServiceDeskPage),
      },
      {
        path: 'analytics',
        element: lazyLoad(AnalyticsPage),
      },
      {
        path: 'profile',
        element: lazyLoad(ProfilePage),
      },
      {
        path: 'sprints',
        element: lazyLoad(SprintsPage),
      },
      {
        path: 'releases',
        element: lazyLoad(ReleasesPage),
      },
      {
        path: 'confluence',
        element: lazyLoad(ConfluencePage),
      },
      {
        path: '*',
        element: <Navigate to="/" replace />,
      },
    ],
  },
]);
