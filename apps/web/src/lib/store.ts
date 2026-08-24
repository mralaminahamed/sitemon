import { create } from "zustand";

type Theme = "dark" | "light";

function initialTheme(): Theme {
  try {
    const t = localStorage.getItem("sitemon.theme");
    if (t === "light" || t === "dark") return t;
  } catch {
    /* ignore */
  }
  return "dark";
}

export function applyTheme(t: Theme) {
  document.documentElement.classList.toggle("light", t === "light");
}

interface UIState {
  theme: Theme;
  toggleTheme: () => void;
  cmdkOpen: boolean;
  setCmdk: (open: boolean) => void;
}

export const useUI = create<UIState>((set, get) => ({
  theme: initialTheme(),
  toggleTheme: () => {
    const next: Theme = get().theme === "dark" ? "light" : "dark";
    applyTheme(next);
    try {
      localStorage.setItem("sitemon.theme", next);
    } catch {
      /* ignore */
    }
    set({ theme: next });
  },
  cmdkOpen: false,
  setCmdk: (cmdkOpen) => set({ cmdkOpen }),
}));
