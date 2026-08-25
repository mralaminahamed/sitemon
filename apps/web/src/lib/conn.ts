import { create } from "zustand";

// Live-stream connection state, surfaced in the top bar.
interface ConnState {
  connected: boolean;
  setConnected: (c: boolean) => void;
}

export const useConn = create<ConnState>((set) => ({
  connected: false,
  setConnected: (connected) => set({ connected }),
}));
