import { useCallback, useEffect, useState } from 'react';

import {
  useCompetitors,
  useCoverageGridCells,
  useCoverageRecommendations,
  useCoverageSummary,
  useLockers,
  useProvinces,
} from '../../api/queries';
import type { CoverageTier, GapSuggestion, ProvinceInfo } from '../../api/contracts';
import { CoverageMap } from './CoverageMap';
import { LayerControls } from './LayerControls';
import { ProvinceSelector } from './ProvinceSelector';
import { RecommendationsPanel } from './RecommendationsPanel';
import { SummaryRibbon } from './SummaryRibbon';

const DEFAULT_TIERS: CoverageTier[] = ['greenfield', 'competitive'];

export type MapFocus =
  | { kind: 'point'; lat: number; lng: number; zoom: number; tick: number }
  | {
      kind: 'bounds';
      minLat: number;
      minLng: number;
      maxLat: number;
      maxLng: number;
      tick: number;
    };

export default function CoverageView() {
  const [cellM, setCellM] = useState(1500);
  const [visibleTiers, setVisibleTiers] = useState<Set<CoverageTier>>(
    new Set(DEFAULT_TIERS),
  );
  const [showInpost, setShowInpost] = useState(true);
  const [showCompetitors, setShowCompetitors] = useState(true);
  const [showSuggestions, setShowSuggestions] = useState(true);
  const [focus, setFocus] = useState<MapFocus | null>(null);

  const [province, setProvince] = useState<ProvinceInfo | null>(null);
  const { data: provinces } = useProvinces();

  useEffect(() => {
    if (!province && provinces && provinces.length > 0) {
      const next = provinces[0];
      setProvince(next);
      setFocus({
        kind: 'bounds',
        minLat: next.min_lat,
        minLng: next.min_lng,
        maxLat: next.max_lat,
        maxLng: next.max_lng,
        tick: Date.now(),
      });
    }
  }, [province, provinces]);

  const provinceName = province?.name ?? '';
  const summary = useCoverageSummary(provinceName, cellM);
  const gridCells = useCoverageGridCells(provinceName, cellM, !!summary.data);
  const recs = useCoverageRecommendations(provinceName, 50);
  const competitors = useCompetitors(provinceName);
  const lockers = useLockers(undefined, provinceName);

  const focusOn = useCallback((s: GapSuggestion) => {
    setFocus({ kind: 'point', lat: s.lat, lng: s.lng, zoom: 15, tick: Date.now() });
  }, []);

  const handleProvinceChange = useCallback((p: ProvinceInfo) => {
    setProvince(p);
    setFocus({
      kind: 'bounds',
      minLat: p.min_lat,
      minLng: p.min_lng,
      maxLat: p.max_lat,
      maxLng: p.max_lng,
      tick: Date.now(),
    });
  }, []);

  return (
    <section className="coverage">
      <header className="coverage__head">
        <div className="coverage__head-row">
          <div>
            <h1>Coverage gaps & upgrade map</h1>
            <p className="muted">
              Where InPost's network is thin, who else is there, and which existing
              lockers are due for replacement.
            </p>
          </div>
          <ProvinceSelector value={provinceName} onChange={handleProvinceChange} />
        </div>
      </header>

      <SummaryRibbon summary={summary.data} />

      <LayerControls
        visibleTiers={visibleTiers}
        setVisibleTiers={setVisibleTiers}
        showInpost={showInpost}
        setShowInpost={setShowInpost}
        showCompetitors={showCompetitors}
        setShowCompetitors={setShowCompetitors}
        showSuggestions={showSuggestions}
        setShowSuggestions={setShowSuggestions}
        cellM={cellM}
        setCellM={setCellM}
      />

      <div className="coverage__body">
        <div className="coverage__map">
          {(summary.isLoading ||
            gridCells.isLoading ||
            competitors.isLoading ||
            lockers.isLoading) && (
            <div
              className="map-loader"
              role="status"
              aria-live="polite"
            >
              <div className="map-loader__spinner" aria-hidden="true" />
              <div className="map-loader__label">
                Computing coverage for{' '}
                <strong>{provinceName || 'Poland'}</strong>…
              </div>
              <div className="map-loader__sub">
                First load per province can take a few seconds — results are
                cached afterwards.
              </div>
            </div>
          )}
          <CoverageMap
            cells={gridCells.data ?? []}
            cellMeters={summary.data?.cell_meters ?? cellM}
            inpostLockers={lockers.data ?? []}
            competitors={competitors.data ?? []}
            suggestions={recs.data?.new_points ?? []}
            showInpost={showInpost}
            showCompetitors={showCompetitors}
            showSuggestions={showSuggestions}
            visibleTiers={visibleTiers}
            focus={focus}
          />
        </div>
        <RecommendationsPanel
          data={recs.data}
          isLoading={recs.isLoading}
          onPickSuggestion={focusOn}
        />
      </div>

      <footer className="coverage__footer">
        Coverage analysis based on InPost Points API + community-maintained{' '}
        <a
          href="https://www.openstreetmap.org/copyright"
          target="_blank"
          rel="noreferrer"
        >
          OpenStreetMap
        </a>{' '}
        competitor data. Counts are a lower bound — some competitor lockers may
        not yet be tagged in OSM.
      </footer>
    </section>
  );
}
