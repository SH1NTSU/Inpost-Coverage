const STREETS = [
    'Dobrowolskiego', 'Aleja Pokoju', 'Olszanicka', 'Warszawska', 'Karmelicka',
    'Floriańska', 'Długa', 'Starowiślna', 'Krakowska', 'Lubicz', 'Grzegórzecka',
    'Mogilska', 'Wielicka', 'Kalwaryjska', 'Zwierzyniecka', 'Krowoderska',
];
const CITY = 'Kraków';
function rand(seed) {
    const x = Math.sin(seed) * 10000;
    return x - Math.floor(x);
}
function makeLocker(i) {
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
const MOCK_LOCKERS = Array.from({ length: 80 }, (_, i) => makeLocker(i));
function detailFor(l) {
    return {
        ...l,
        country: 'PL',
        province: 'małopolskie',
        post_code: '30-394',
        location_type: 'outdoor',
        physical_type: 'parcel_locker',
    };
}
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
export const mockApi = {
    async cities() {
        await sleep(50);
        return [
            { name: 'Kraków', province: 'małopolskie', point_count: 764, min_lat: 49.97, min_lng: 19.79, max_lat: 50.13, max_lng: 20.1, center_lat: 50.06, center_lng: 19.94 },
            { name: 'Warszawa', province: 'mazowieckie', point_count: 195, min_lat: 52.1, min_lng: 20.85, max_lat: 52.37, max_lng: 21.27, center_lat: 52.23, center_lng: 21.0 },
        ];
    },
    async lockers(status) {
        await sleep(200);
        if (!status)
            return MOCK_LOCKERS;
        return MOCK_LOCKERS.filter((l) => l.current_status === status);
    },
    async locker(id) {
        await sleep(150);
        const l = MOCK_LOCKERS.find((x) => x.id === id);
        if (!l)
            throw new Error('Locker not found');
        return detailFor(l);
    },
    async coverageGrid(cellM) {
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
    async coverageRecommendations(_cellM, _limit) {
        await sleep(180);
        return { new_points: [], upgrades: [] };
    },
    async competitors() {
        await sleep(150);
        return [];
    },
};
