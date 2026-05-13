import { mockApi } from './mocks';
const BASE = import.meta.env.VITE_API_BASE_URL?.trim() ?? '';
const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true' ||
    import.meta.env.VITE_USE_MOCK === '1';
async function request(path) {
    const res = await fetch(`${BASE}${path}`);
    if (!res.ok) {
        const body = await res.text().catch(() => '');
        throw new Error(`HTTP ${res.status}: ${body || res.statusText}`);
    }
    return res.json();
}
function qs(params) {
    const parts = [];
    for (const [k, v] of Object.entries(params)) {
        if (v === undefined || v === '' || v === null)
            continue;
        parts.push(`${encodeURIComponent(k)}=${encodeURIComponent(String(v))}`);
    }
    return parts.length ? `?${parts.join('&')}` : '';
}
export const api = {
    isMock: USE_MOCK,
    cities() {
        return USE_MOCK ? mockApi.cities() : request('/api/v1/cities?min_points=20');
    },
    lockers(status, city) {
        if (USE_MOCK)
            return mockApi.lockers(status);
        return request(`/api/v1/lockers${qs({ status, city })}`);
    },
    locker(id) {
        return USE_MOCK ? mockApi.locker(id) : request(`/api/v1/lockers/${id}`);
    },
    coverageSummary(city, cellM = 400) {
        return USE_MOCK
            ? mockApi.coverageGrid(cellM).then((r) => r.summary)
            : request(`/api/v1/coverage/summary${qs({ city, cell_m: cellM })}`);
    },
    coverageGridCells(city, cellM = 400) {
        return USE_MOCK
            ? mockApi.coverageGrid(cellM).then((r) => r.cells)
            : request(`/api/v1/coverage/grid-cells${qs({ city, cell_m: cellM })}`);
    },
    coverageRecommendations(city, limit = 10) {
        return USE_MOCK
            ? mockApi.coverageRecommendations(400, limit)
            : request(`/api/v1/coverage/recommendations${qs({ city, limit })}`);
    },
    competitors(city) {
        return USE_MOCK
            ? mockApi.competitors()
            : request(`/api/v1/coverage/competitors${qs({ city })}`);
    },
};
