import { useQuery } from '@tanstack/react-query';
import { api } from './client';
export const queryKeys = {
    provinces: ['provinces'],
    lockers: (status, province) => ['lockers', status ?? 'all', province ?? 'all'],
    locker: (id) => ['lockers', id],
    coverageSummary: (province, cellM) => ['coverage', 'summary', province, cellM],
    coverageGridCells: (province, cellM) => ['coverage', 'grid-cells', province, cellM],
    coverageRecommendations: (province, limit) => ['coverage', 'recs', province, limit],
    competitors: (province) => ['coverage', 'competitors', province],
};
export function useProvinces() {
    return useQuery({
        queryKey: queryKeys.provinces,
        queryFn: () => api.provinces(),
        staleTime: 5 * 60_000,
    });
}
export function useLockers(status, province) {
    return useQuery({
        queryKey: queryKeys.lockers(status, province),
        queryFn: () => api.lockers(status, province),
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
export function useCoverageSummary(province, cellM = 800) {
    return useQuery({
        queryKey: queryKeys.coverageSummary(province, cellM),
        queryFn: () => api.coverageSummary(province, cellM),
        enabled: !!province,
        staleTime: 5 * 60_000,
    });
}
export function useCoverageGridCells(province, cellM = 800, enabled = true) {
    return useQuery({
        queryKey: queryKeys.coverageGridCells(province, cellM),
        queryFn: () => api.coverageGridCells(province, cellM),
        enabled: !!province && enabled,
        staleTime: 5 * 60_000,
    });
}
export function useCoverageRecommendations(province, limit = 10) {
    return useQuery({
        queryKey: queryKeys.coverageRecommendations(province, limit),
        queryFn: () => api.coverageRecommendations(province, limit),
        enabled: !!province,
        staleTime: 5 * 60_000,
    });
}
export function useCompetitors(province) {
    return useQuery({
        queryKey: queryKeys.competitors(province),
        queryFn: () => api.competitors(province),
        enabled: !!province,
        staleTime: 5 * 60_000,
    });
}
