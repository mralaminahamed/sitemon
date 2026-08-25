import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export function useStatus() {
  return useQuery({ queryKey: ["status"], queryFn: api.status, refetchInterval: 15000 });
}

export function useMonitors() {
  return useQuery({ queryKey: ["monitors"], queryFn: api.monitors, refetchInterval: 30000 });
}

export function useHistory(url: string | null) {
  return useQuery({
    queryKey: ["history", url],
    queryFn: () => api.history(url!),
    enabled: !!url,
    refetchInterval: 15000,
  });
}

export function useStats(url: string | null) {
  return useQuery({
    queryKey: ["stats", url],
    queryFn: () => api.stats(url!),
    enabled: !!url,
    refetchInterval: 15000,
  });
}

export function useAddMonitor() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (url: string) => api.addMonitor(url),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["monitors"] }),
  });
}

export function useDeleteMonitor() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (url: string) => api.deleteMonitor(url),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["monitors"] });
      qc.invalidateQueries({ queryKey: ["status"] });
    },
  });
}

export function useCheck() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (urls: string[]) => api.check(urls),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["status"] }),
  });
}
