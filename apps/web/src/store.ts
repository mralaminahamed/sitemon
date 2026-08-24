import { create } from "zustand";

interface UIState {
  selectedUrl: string | null;
  select: (url: string | null) => void;
}

export const useUI = create<UIState>((set) => ({
  selectedUrl: null,
  select: (url) => set({ selectedUrl: url }),
}));
