

import { useQuery } from '@tanstack/react-query';

import { api } from './client';
import type { LockerStatus } from './contracts';

export const queryKeys = {
  provinces: ['provinces'] as const,
  lockers: (status?: LockerStatus, province?: string) =>
    ['lockers', status ?? 'all', province ?? 'all'] as const,
  locker: (id: number) => ['lockers', id] as const,
  coverageSummary: (province: string, cellM: number) =>
    ['coverage', 'summary', province, cellM] as const,
  coverageGridCells: (province: string, cellM: number) =>
    ['coverage', 'grid-cells', province, cellM] as const,
  coverageRecommendations: (province: string, limit: number) =>
    ['coverage', 'recs', province, limit] as const,
  competitors: (province: string) => ['coverage', 'competitors', province] as const,
};

export function useProvinces() {
  return useQuery({
    queryKey: queryKeys.provinces,
    queryFn: () => api.provinces(),
    staleTime: 5 * 60_000,
  });
}

export function useLockers(status?: LockerStatus, province?: string) {
  return useQuery({
    queryKey: queryKeys.lockers(status, province),
    queryFn: () => api.lockers(status, province),
    staleTime: 5 * 60_000,
  });
}

export function useLocker(id: number | undefined) {
  return useQuery({
    queryKey: queryKeys.locker(id ?? 0),
    queryFn: () => api.locker(id!),
    enabled: typeof id === 'number' && id > 0,
  });
}

export function useCoverageSummary(province: string, cellM = 800) {
  return useQuery({
    queryKey: queryKeys.coverageSummary(province, cellM),
    queryFn: () => api.coverageSummary(province, cellM),
    enabled: !!province,
    staleTime: 5 * 60_000,
  });
}

export function useCoverageGridCells(province: string, cellM = 800, enabled = true) {
  return useQuery({
    queryKey: queryKeys.coverageGridCells(province, cellM),
    queryFn: () => api.coverageGridCells(province, cellM),
    enabled: !!province && enabled,
    staleTime: 5 * 60_000,
  });
}

export function useCoverageRecommendations(province: string, limit = 10) {
  return useQuery({
    queryKey: queryKeys.coverageRecommendations(province, limit),
    queryFn: () => api.coverageRecommendations(province, limit),
    enabled: !!province,
    staleTime: 5 * 60_000,
  });
}

export function useCompetitors(province: string) {
  return useQuery({
    queryKey: queryKeys.competitors(province),
    queryFn: () => api.competitors(province),
    enabled: !!province,
    staleTime: 5 * 60_000,
  });
}
