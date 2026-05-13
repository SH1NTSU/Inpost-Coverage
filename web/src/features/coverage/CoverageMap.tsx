import { useDeferredValue, useEffect, useMemo } from 'react';
import {
  CircleMarker,
  MapContainer,
  Popup,
  Rectangle,
  TileLayer,
  useMap,
} from 'react-leaflet';
import { Link } from 'react-router-dom';

import type {
  CompetitorPoint,
  CoverageTier,
  GapSuggestion,
  GridCell,
  LockerSummary,
} from '../../api/contracts';
import { TIER_COLOR, networkLabel, formatMeters } from '../../lib/tier';

interface Props {
  cells: GridCell[];
  cellMeters: number;
  inpostLockers: LockerSummary[];
  competitors: CompetitorPoint[];
  suggestions: GapSuggestion[];
  showInpost: boolean;
  showCompetitors: boolean;
  showSuggestions: boolean;
  visibleTiers: Set<CoverageTier>;
  focus: { lat: number; lng: number; tick: number } | null;
}

function MapController({
  focus,
}: {
  focus: { lat: number; lng: number; tick: number } | null;
}) {
  const map = useMap();
  useEffect(() => {
    if (!focus) return;
    map.flyTo([focus.lat, focus.lng], 16, { animate: true, duration: 0.7 });
  }, [focus, map]);
  return null;
}

const DEFAULT_CENTER: [number, number] = [52.23, 21.0];

const M_PER_DEG_LAT = 111_320;
const M_PER_DEG_LNG_50 = 71_500;

export function CoverageMap({
  cells,
  cellMeters,
  inpostLockers,
  competitors,
  suggestions,
  showInpost,
  showCompetitors,
  showSuggestions,
  visibleTiers,
  focus,
}: Props) {
  const deferredCells = useDeferredValue(cells);
  const deferredInpostLockers = useDeferredValue(inpostLockers);
  const deferredCompetitors = useDeferredValue(competitors);
  const deferredSuggestions = useDeferredValue(suggestions);

  const halfLat = cellMeters / 2 / M_PER_DEG_LAT;
  const halfLng = cellMeters / 2 / M_PER_DEG_LNG_50;

  const rectangles = useMemo(
    () =>
      deferredCells
        .filter((c) => visibleTiers.has(c.tier))
        .map((c) => ({
          key: `${c.lat.toFixed(5)},${c.lng.toFixed(5)}`,
          bounds: [
            [c.lat - halfLat, c.lng - halfLng],
            [c.lat + halfLat, c.lng + halfLng],
          ] as [[number, number], [number, number]],
          tier: c.tier,
          inpostM: c.nearest_inpost_m,
          competitorM: c.nearest_competitor_m,
          net: c.nearest_competitor_network,
        })),
    [deferredCells, halfLat, halfLng, visibleTiers],
  );

  return (
    <MapContainer center={DEFAULT_CENTER} zoom={12} className="map" preferCanvas>
      <MapController focus={focus} />
      <TileLayer
        attribution='© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors © <a href="https://carto.com/attributions">CARTO</a>'
        url="https://{s}.basemaps.cartocdn.com/rastertiles/voyager/{z}/{x}/{y}{r}.png"
      />

      {rectangles.map((r) => (
        <Rectangle
          key={r.key}
          bounds={r.bounds}
          pathOptions={{
            color: TIER_COLOR[r.tier],
            fillColor: TIER_COLOR[r.tier],
            fillOpacity: 0.35,
            weight: 0,
          }}
        >
          <Popup>
            <div className="popup popup--coverage">
              <strong>{r.tier.replace('_', ' ')}</strong>
              <div>InPost: {formatMeters(r.inpostM)}</div>
              <div>
                Competitor: {formatMeters(r.competitorM)} ({networkLabel(r.net)})
              </div>
            </div>
          </Popup>
        </Rectangle>
      ))}

      {showInpost &&
        deferredInpostLockers.map((l) => (
          <CircleMarker
            key={`i-${l.id}`}
            center={[l.latitude, l.longitude]}
            radius={3}
            pathOptions={{
              color: '#3b82f6',
              fillColor: '#3b82f6',
              fillOpacity: 0.9,
              weight: 0,
            }}
          >
            <Popup>
              <div className="popup">
                <strong>{l.inpost_id}</strong>
                <div>{l.street} {l.building_no}</div>
                <div className="popup__city">{l.city}</div>
                <Link to={`/lockers/${l.id}`} className="popup__link">
                  Details →
                </Link>
              </div>
            </Popup>
          </CircleMarker>
        ))}

      {showCompetitors &&
        deferredCompetitors.map((c) => (
          <CircleMarker
            key={`c-${c.id}`}
            center={[c.latitude, c.longitude]}
            radius={3}
            pathOptions={{
              color: '#9ca3af',
              fillColor: '#9ca3af',
              fillOpacity: 0.8,
              weight: 0,
            }}
          >
            <Popup>
              <strong>{networkLabel(c.network)}</strong>
              {c.name && <div>{c.name}</div>}
            </Popup>
          </CircleMarker>
        ))}

      {showSuggestions &&
        deferredSuggestions.map((s, i) => (
          <CircleMarker
            key={`s-${i}`}
            center={[s.lat, s.lng]}
            radius={10}
            pathOptions={{
              color: '#fbbf24',
              fillColor: '#fbbf24',
              fillOpacity: 0.9,
              weight: 2,
            }}
          >
            <Popup>
              <div className="popup popup--coverage">
                <strong>
                  #{i + 1} {s.anchor.brand || s.anchor.name || s.anchor.poi_type}
                </strong>
                <div className="muted">{s.anchor.poi_type}</div>
                <div className="popup__city">{s.reason}</div>
              </div>
            </Popup>
          </CircleMarker>
        ))}
    </MapContainer>
  );
}
