

import type {
  CompetitorPoint,
  CoverageGridResponse,
  CoverageRecommendations,
  LockerDetail,
  LockerSummary,
  ProvinceInfo,
} from './contracts';

const STREETS = [
  'Dobrowolskiego', 'Aleja Pokoju', 'Olszanicka', 'Warszawska', 'Karmelicka',
  'Floriańska', 'Długa', 'Starowiślna', 'Krakowska', 'Lubicz', 'Grzegórzecka',
  'Mogilska', 'Wielicka', 'Kalwaryjska', 'Zwierzyniecka', 'Krowoderska',
];

const CITY = 'Kraków';

function rand(seed: number): number {
  const x = Math.sin(seed) * 10000;
  return x - Math.floor(x);
}

function makeLocker(i: number): LockerSummary {
  const id = i + 1;
  const lat = 50.0614 + (rand(i * 7.1) - 0.5) * 0.18;
  const lng = 19.9366 + (rand(i * 3.7) - 0.5) * 0.18;
  return {
    id,
    inpost_id: `KRA${String(id).padStart(3, '0')}M`,
    city: CITY,
    street: STREETS[i % STREETS.length],
    building_no: String((i % 50) + 1),
    latitude: lat,
    longitude: lng,
    image_url: '',
    location_247: rand(i * 1.5) > 0.4,
    is_next: rand(i * 2.3) > 0.7,
    current_status: 'Operating',
    last_change_at: null,
  };
}

const MOCK_LOCKERS: LockerSummary[] = Array.from({ length: 80 }, (_, i) => makeLocker(i));

function detailFor(l: LockerSummary): LockerDetail {
  return {
    ...l,
    country: 'PL',
    province: 'małopolskie',
    post_code: '30-394',
    location_type: 'outdoor',
    physical_type: 'parcel_locker',
  };
}

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

export const mockApi = {
  async provinces(): Promise<ProvinceInfo[]> {
    await sleep(50);
    return [
      { name: 'małopolskie', point_count: 1240, min_lat: 49.4, min_lng: 19.1, max_lat: 50.5, max_lng: 21.4, center_lat: 49.95, center_lng: 20.25 },
      { name: 'mazowieckie', point_count: 1810, min_lat: 51.0, min_lng: 19.6, max_lat: 53.4, max_lng: 23.1, center_lat: 52.2, center_lng: 21.35 },
    ];
  },
  async lockers(status?: string): Promise<LockerSummary[]> {
    await sleep(200);
    if (!status) return MOCK_LOCKERS;
    return MOCK_LOCKERS.filter((l) => l.current_status === status);
  },
  async locker(id: number): Promise<LockerDetail> {
    await sleep(150);
    const l = MOCK_LOCKERS.find((x) => x.id === id);
    if (!l) throw new Error('Locker not found');
    return detailFor(l);
  },
  async coverageGrid(cellM: number): Promise<CoverageGridResponse> {
    await sleep(200);
    return {
      summary: {
        cell_meters: cellM,
        total_cells: 3200,
        greenfield_cells: 1200,
        competitive_cells: 350,
        inpost_only_cells: 500,
        saturated_cells: 1150,
        underserved_km2: 220,
        inpost_lockers: MOCK_LOCKERS.length,
        competitor_lockers: 900,
      },
      cells: [],
    };
  },
  async coverageRecommendations(_cellM: number, _limit: number): Promise<CoverageRecommendations> {
    await sleep(180);
    return { new_points: [], upgrades: [] };
  },
  async competitors(): Promise<CompetitorPoint[]> {
    await sleep(150);
    return [];
  },
};
