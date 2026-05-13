export const TIER_COLOR = {
    greenfield: '#ef4444',
    competitive: '#f97316',
    inpost_only: '#3b82f6',
    saturated: '#a3a3a3',
};
export const TIER_LABEL = {
    greenfield: 'Greenfield',
    competitive: 'Competitive gap',
    inpost_only: 'InPost only',
    saturated: 'Saturated',
};
export const TIER_DESCRIPTION = {
    greenfield: 'No locker from any network within 1 km — pure opportunity.',
    competitive: 'Only a competitor is nearby. InPost is losing customers here.',
    inpost_only: 'InPost present, no competitor within 1 km — protect this turf.',
    saturated: 'Both InPost and a competitor present. No urgent action.',
};
export const NETWORK_LABEL = {
    AllegroOne: 'Allegro One',
    DHL: 'DHL',
    OrlenPaczka: 'Orlen Paczka',
    PocztaPolska: 'Poczta Polska',
    DPD: 'DPD',
    GLS: 'GLS',
    UPS: 'UPS',
    FedEx: 'FedEx',
};
export function networkLabel(n) {
    if (!n)
        return '—';
    return NETWORK_LABEL[n] ?? n;
}
export function formatMeters(m) {
    if (m < 1000)
        return `${Math.round(m)} m`;
    return `${(m / 1000).toFixed(1)} km`;
}
