import { useQuery } from '@tanstack/react-query';
import { api } from './client';
export const queryKeys = {
    cities: ['cities'],
    lockers: (status, city) => ['lockers', status ?? 'all', city ?? 'all'],
    locker: (id) => ['lockers', id],
    coverageSummary: (city, cellM) => ['coverage', 'summary', city, cellM],
    coverageGridCells: (city, cellM) => ['coverage', 'grid-cells', city, cellM],
    coverageRecommendations: (city, limit) => ['coverage', 'recs', city, limit],
    competitors: (city) => ['coverage', 'competitors', city],
};
export function useCities() {
    return useQuery({
        queryKey: queryKeys.cities,
        queryFn: () => api.cities(),
        staleTime: 5 * 60_000,
    });
}
export function useLockers(status, city) {
    return useQuery({
        queryKey: queryKeys.lockers(status, city),
        queryFn: () => api.lockers(status, city),
        staleTime: 5 * 60_000,
    });
}
export function useLocker(id) {
    return useQuery({
        queryKey: queryKeys.locker(id ?? 0),
        queryFn: () => api.locker(id),
        enabled: typeof id === 'number' && id > 0,
    });
}
export function useCoverageSummary(city, cellM = 400) {
    return useQuery({
        queryKey: queryKeys.coverageSummary(city, cellM),
        queryFn: () => api.coverageSummary(city, cellM),
        enabled: !!city,
        staleTime: 5 * 60_000,
    });
}
export function useCoverageGridCells(city, cellM = 400, enabled = true) {
    return useQuery({
        queryKey: queryKeys.coverageGridCells(city, cellM),
        queryFn: () => api.coverageGridCells(city, cellM),
        enabled: !!city && enabled,
        staleTime: 5 * 60_000,
    });
}
export function useCoverageRecommendations(city, limit = 10) {
    return useQuery({
        queryKey: queryKeys.coverageRecommendations(city, limit),
        queryFn: () => api.coverageRecommendations(city, limit),
        enabled: !!city,
        staleTime: 5 * 60_000,
    });
}
export function useCompetitors(city) {
    return useQuery({
        queryKey: queryKeys.competitors(city),
        queryFn: () => api.competitors(city),
        enabled: !!city,
        staleTime: 5 * 60_000,
    });
}
