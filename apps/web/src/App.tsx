import { Outlet } from "react-router-dom";
import { Sidebar } from "./components/Sidebar";
import { TopBar } from "./components/TopBar";
import { Toaster } from "./components/Toaster";
import { CommandPalette } from "./components/CommandPalette";
import { useStatusStream } from "./lib/useStatusStream";

export function App() {
  useStatusStream();
  return (
    <div className="flex h-full">
      <Sidebar />
      <div className="relative flex min-w-0 flex-1 flex-col">
        <div className="grid-bg pointer-events-none absolute inset-0 h-64" />
        <TopBar />
        <main className="relative flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
      <Toaster />
      <CommandPalette />
    </div>
  );
}
