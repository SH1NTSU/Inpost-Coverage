

import type {
  CityInfo,
  CompetitorPoint,
  CoverageRecommendations,
  CoverageSummary,
  GridCell,
  LockerDetail,
  LockerSummary,
} from './contracts';
import { mockApi } from './mocks';

const BASE = import.meta.env.VITE_API_BASE_URL?.trim() ?? '';
const USE_MOCK =
  import.meta.env.VITE_USE_MOCK === 'true' ||
  import.meta.env.VITE_USE_MOCK === '1';

async function request<T>(path: string): Promise<T> {
  const res = await fetch(`${BASE}${path}`);
  if (!res.ok) {
    const body = await res.text().catch(() => '');
    throw new Error(`HTTP ${res.status}: ${body || res.statusText}`);
  }
  return res.json() as Promise<T>;
}

function qs(params: Record<string, string | number | undefined>): string {
  const parts: string[] = [];
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === '' || v === null) continue;
    parts.push(`${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`);
  }
  return parts.length ? `?${parts.join('&')}` : '';
}

export const api = {
  isMock: USE_MOCK,

  cities(): Promise<CityInfo[]> {
    return USE_MOCK ? mockApi.cities() : request('/api/v1/cities?min_points=20');
  },

  lockers(status?: 'Operating' | 'Disabled', city?: string): Promise<LockerSummary[]> {
    if (USE_MOCK) return mockApi.lockers(status);
    return request(`/api/v1/lockers${qs({ status, city })}`);
  },

  locker(id: number): Promise<LockerDetail> {
    return USE_MOCK ? mockApi.locker(id) : request(`/api/v1/lockers/${id}`);
  },

  coverageSummary(city: string, cellM = 400): Promise<CoverageSummary> {
    return USE_MOCK
      ? mockApi.coverageGrid(cellM).then((r) => r.summary)
      : request(`/api/v1/coverage/summary${qs({ city, cell_m: cellM })}`);
  },

  coverageGridCells(city: string, cellM = 400): Promise<GridCell[]> {
    return USE_MOCK
      ? mockApi.coverageGrid(cellM).then((r) => r.cells)
      : request(`/api/v1/coverage/grid-cells${qs({ city, cell_m: cellM })}`);
  },

  coverageRecommendations(city: string, limit = 10): Promise<CoverageRecommendations> {
    return USE_MOCK
      ? mockApi.coverageRecommendations(400, limit)
      : request(`/api/v1/coverage/recommendations${qs({ city, limit })}`);
  },

  competitors(city: string): Promise<CompetitorPoint[]> {
    return USE_MOCK
      ? mockApi.competitors()
      : request(`/api/v1/coverage/competitors${qs({ city })}`);
  },
};
