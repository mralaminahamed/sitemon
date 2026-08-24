import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";

export function useStatus() {
  return useQuery({
    queryKey: ["status"],
    queryFn: api.status,
    refetchInterval: 5000,
  });
}

export function useHistory(url: string | null) {
  return useQuery({
    queryKey: ["history", url],
    queryFn: () => api.history(url!),
    enabled: !!url,
    refetchInterval: 10000,
  });
}

export function useStats(url: string | null) {
  return useQuery({
    queryKey: ["stats", url],
    queryFn: () => api.stats(url!),
    enabled: !!url,
    refetchInterval: 10000,
  });
}

export function useCheck() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (urls: string[]) => api.check(urls),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["status"] }),
  });
}
