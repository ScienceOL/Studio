import { Outlet } from 'react-router-dom';

export default function DashboardLayout() {
  return (
    <div className="relative h-screen w-screen overflow-hidden bg-neutral-50 dark:bg-neutral-900">
      {/* Xyzen Side Panel (Global) */}

      {/* Main Content Area (Desktop) */}
      <main className="absolute inset-0 transition-all duration-300 ease-in-out">
        {/* The Desktop component will be rendered here via routing */}
        <Outlet />
      </main>
    </div>
  );
}
