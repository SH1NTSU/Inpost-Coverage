import type { CoverageTier } from '../api/contracts';

export const TIER_COLOR: Record<CoverageTier, string> = {
  greenfield: '#ef4444',  
  competitive: '#f97316', 
  inpost_only: '#3b82f6', 
  saturated: '#a3a3a3',   
};

export const TIER_LABEL: Record<CoverageTier, string> = {
  greenfield: 'Greenfield',
  competitive: 'Competitive gap',
  inpost_only: 'InPost only',
  saturated: 'Saturated',
};

export const TIER_DESCRIPTION: Record<CoverageTier, string> = {
  greenfield:
    'No locker from any network within 1 km — pure opportunity.',
  competitive:
    'Only a competitor is nearby. InPost is losing customers here.',
  inpost_only:
    'InPost present, no competitor within 1 km — protect this turf.',
  saturated:
    'Both InPost and a competitor present. No urgent action.',
};

export const NETWORK_LABEL: Record<string, string> = {
  AllegroOne: 'Allegro One',
  DHL: 'DHL',
  OrlenPaczka: 'Orlen Paczka',
  PocztaPolska: 'Poczta Polska',
  DPD: 'DPD',
  GLS: 'GLS',
  UPS: 'UPS',
  FedEx: 'FedEx',
};

export function networkLabel(n: string | undefined): string {
  if (!n) return '—';
  return NETWORK_LABEL[n] ?? n;
}

export function formatMeters(m: number): string {
  if (m < 1000) return `${Math.round(m)} m`;
  return `${(m / 1000).toFixed(1)} km`;
}
