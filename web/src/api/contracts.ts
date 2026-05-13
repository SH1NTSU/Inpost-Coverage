

export type LockerStatus = 'Operating' | 'Disabled';

export type LocationClass =
  | 'generic'
  | 'shopping_mall'
  | 'residential'
  | 'transport_hub'
  | 'university'
  | 'office';

export interface LockerSummary {
  id: number;
  inpost_id: string;
  city: string;
  street: string;
  building_no: string;
  latitude: number;
  longitude: number;
  image_url: string;
  location_247: boolean;
  is_next: boolean;
  current_status: LockerStatus;
  last_change_at: string | null;
}

export interface LockerDetail extends LockerSummary {
  country: string;
  province: string;
  post_code: string;
  location_type: string;
  physical_type: string;
}

export type CoverageTier =
  | 'greenfield'
  | 'competitive'
  | 'inpost_only'
  | 'saturated';

export interface GridCell {
  lat: number;
  lng: number;
  nearest_inpost_m: number;
  nearest_inpost_id?: number;
  nearest_competitor_m: number;
  nearest_competitor_network?: string;
  tier: CoverageTier;
}

export interface CoverageSummary {
  cell_meters: number;
  total_cells: number;
  greenfield_cells: number;
  competitive_cells: number;
  inpost_only_cells: number;
  saturated_cells: number;
  underserved_km2: number;
  inpost_lockers: number;
  competitor_lockers: number;
}

export interface CoverageGridResponse {
  summary: CoverageSummary;
  cells: GridCell[];
}

export interface NearbyAnchor {
  poi_type: string;          
  brand: string;             
  name: string;
  latitude: number;
  longitude: number;
  distance_m: number;
}

export interface GapSuggestion {
  lat: number;
  lng: number;
  tier: CoverageTier;
  nearest_inpost_m: number;
  nearest_competitor_m: number;
  nearest_competitor_network?: string;
  reason: string;
  
  anchor: NearbyAnchor;
  
  nearby_anchors: NearbyAnchor[];
}

export interface UpgradeCandidate {
  locker: LockerSummary;
  is_next: boolean;
  competitor_pressure: number;   
  score: number;
  reasons: string[];
}

export interface CoverageRecommendations {
  new_points: GapSuggestion[];
  upgrades: UpgradeCandidate[];
}

export interface CompetitorPoint {
  id: number;
  network: string;
  name: string;
  latitude: number;
  longitude: number;
  address: string;
  osm_id: number;
  fetched_at: string;
}

export interface CityInfo {
  name: string;
  province: string;
  point_count: number;
  min_lat: number;
  min_lng: number;
  max_lat: number;
  max_lng: number;
  center_lat: number;
  center_lng: number;
}
