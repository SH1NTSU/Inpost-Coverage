

import { useQuery } from '@tanstack/react-query';

import { api } from './client';
import type { LockerStatus } from './contracts';

export const queryKeys = {
  cities: ['cities'] as const,
  lockers: (status?: LockerStatus, city?: string) =>
    ['lockers', status ?? 'all', city ?? 'all'] as const,
  locker: (id: number) => ['lockers', id] as const,
  coverageSummary: (city: string, cellM: number) =>
    ['coverage', 'summary', city, cellM] as const,
  coverageGridCells: (city: string, cellM: number) =>
    ['coverage', 'grid-cells', city, cellM] as const,
  coverageRecommendations: (city: string, limit: number) =>
    ['coverage', 'recs', city, limit] as const,
  competitors: (city: string) => ['coverage', 'competitors', city] as const,
};

export function useCities() {
  return useQuery({
    queryKey: queryKeys.cities,
    queryFn: () => api.cities(),
    staleTime: 5 * 60_000,
  });
}

export function useLockers(status?: LockerStatus, city?: string) {
  return useQuery({
    queryKey: queryKeys.lockers(status, city),
    queryFn: () => api.lockers(status, city),
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

export function useCoverageSummary(city: string, cellM = 400) {
  return useQuery({
    queryKey: queryKeys.coverageSummary(city, cellM),
    queryFn: () => api.coverageSummary(city, cellM),
    enabled: !!city,
    staleTime: 5 * 60_000,
  });
}

export function useCoverageGridCells(city: string, cellM = 400, enabled = true) {
  return useQuery({
    queryKey: queryKeys.coverageGridCells(city, cellM),
    queryFn: () => api.coverageGridCells(city, cellM),
    enabled: !!city && enabled,
    staleTime: 5 * 60_000,
  });
}

export function useCoverageRecommendations(city: string, limit = 10) {
  return useQuery({
    queryKey: queryKeys.coverageRecommendations(city, limit),
    queryFn: () => api.coverageRecommendations(city, limit),
    enabled: !!city,
    staleTime: 5 * 60_000,
  });
}

export function useCompetitors(city: string) {
  return useQuery({
    queryKey: queryKeys.competitors(city),
    queryFn: () => api.competitors(city),
    enabled: !!city,
    staleTime: 5 * 60_000,
  });
}
