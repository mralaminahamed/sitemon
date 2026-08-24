import { NavLink } from "react-router-dom";
import { Activity, Gauge, LayoutDashboard, Settings, ShieldCheck } from "lucide-react";
import { Pulse } from "./Pulse";

const nav = [
  { to: "/", label: "Dashboard", icon: LayoutDashboard, end: true },
  { to: "/loadtest", label: "Load Test", icon: Gauge, end: false },
  { to: "/ssl", label: "SSL", icon: ShieldCheck, end: false },
  { to: "/settings", label: "Settings", icon: Settings, end: false },
];

export function Sidebar() {
  return (
    <aside className="flex w-60 shrink-0 flex-col border-r border-border bg-panel/60 backdrop-blur">
      <div className="flex items-center gap-2.5 px-5 py-5">
        <Activity size={20} className="text-accent" />
        <div>
          <div className="font-display text-lg font-bold leading-none tracking-tight">sitemon</div>
          <div className="-mt-0.5 h-3">
            <Pulse width={78} height={12} />
          </div>
        </div>
      </div>

      <nav className="flex flex-col gap-1 px-3 py-2">
        {nav.map(({ to, label, icon: Icon, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              `flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition ${
                isActive
                  ? "bg-panel2 text-text"
                  : "text-muted hover:bg-panel2 hover:text-text"
              }`
            }
          >
            <Icon size={17} />
            {label}
          </NavLink>
        ))}
      </nav>

      <div className="mt-auto px-5 py-4 font-mono text-[10px] uppercase tracking-wider text-muted">
        control room
      </div>
    </aside>
  );
}
