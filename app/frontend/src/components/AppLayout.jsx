import { useState } from "react";
import Sidebar from "./Sidebar";
import NotificationBell from "./NotificationBell";

const AppLayout = ({ children }) => {
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  return (
    <div className="flex min-h-screen bg-slate-50">
      <Sidebar
        isCollapsed={sidebarCollapsed}
        setIsCollapsed={setSidebarCollapsed}
      />

      <div
        className={`relative flex-1 transition-all duration-300 ${sidebarCollapsed ? "ml-20" : "ml-64"}`}
      >
        <div className="absolute right-6 top-4 z-20">
          <NotificationBell />
        </div>
        {children}
      </div>
    </div>
  );
};

export default AppLayout;
