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
    provinces() {
        return USE_MOCK ? mockApi.provinces() : request('/api/v1/provinces');
    },
    lockers(status, province) {
        if (USE_MOCK)
            return mockApi.lockers(status);
        return request(`/api/v1/lockers${qs({ status, province })}`);
    },
    locker(id) {
        return USE_MOCK ? mockApi.locker(id) : request(`/api/v1/lockers/${id}`);
    },
    coverageSummary(province, cellM = 800) {
        return USE_MOCK
            ? mockApi.coverageGrid(cellM).then((r) => r.summary)
            : request(`/api/v1/coverage/summary${qs({ province, cell_m: cellM })}`);
    },
    coverageGridCells(province, cellM = 800) {
        return USE_MOCK
            ? mockApi.coverageGrid(cellM).then((r) => r.cells)
            : request(`/api/v1/coverage/grid-cells${qs({ province, cell_m: cellM })}`);
    },
    coverageRecommendations(province, limit = 10) {
        return USE_MOCK
            ? mockApi.coverageRecommendations(800, limit)
            : request(`/api/v1/coverage/recommendations${qs({ province, limit })}`);
    },
    competitors(province) {
        return USE_MOCK
            ? mockApi.competitors()
            : request(`/api/v1/coverage/competitors${qs({ province })}`);
    },
};
