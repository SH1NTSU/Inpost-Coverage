import { useCallback, useEffect, useState } from 'react';

import {
  useCities,
  useCompetitors,
  useCoverageGridCells,
  useCoverageRecommendations,
  useCoverageSummary,
  useLockers,
} from '../../api/queries';
import type { CityInfo, CoverageTier, GapSuggestion } from '../../api/contracts';
import { CitySelector } from './CitySelector';
import { CoverageMap } from './CoverageMap';
import { LayerControls } from './LayerControls';
import { RecommendationsPanel } from './RecommendationsPanel';
import { SummaryRibbon } from './SummaryRibbon';

const DEFAULT_TIERS: CoverageTier[] = ['greenfield', 'competitive'];

interface MapFocus {
  lat: number;
  lng: number;
  tick: number;
}

export default function CoverageView() {
  const [cellM] = useState(400);
  const [visibleTiers, setVisibleTiers] = useState<Set<CoverageTier>>(
    new Set(DEFAULT_TIERS),
  );
  const [showInpost, setShowInpost] = useState(true);
  const [showCompetitors, setShowCompetitors] = useState(true);
  const [showSuggestions, setShowSuggestions] = useState(true);
  const [focus, setFocus] = useState<MapFocus | null>(null);

  const [city, setCity] = useState<CityInfo | null>(null);
  const { data: cities } = useCities();

  useEffect(() => {
    if (!city && cities && cities.length > 0) {
      const next = cities[0];
      setCity(next);
      setFocus({
        lat: next.center_lat,
        lng: next.center_lng,
        tick: Date.now(),
      });
    }
  }, [city, cities]);

  const cityName = city?.name ?? '';
  const summary = useCoverageSummary(cityName, cellM);
  const gridCells = useCoverageGridCells(cityName, cellM, !!summary.data);
  const recs = useCoverageRecommendations(cityName, 5);
  const competitors = useCompetitors(cityName);
  const lockers = useLockers(undefined, cityName);

  const focusOn = useCallback((s: GapSuggestion) => {
    setFocus({ lat: s.lat, lng: s.lng, tick: Date.now() });
  }, []);

  const handleCityChange = useCallback((c: CityInfo) => {
    setCity(c);
    setFocus({ lat: c.center_lat, lng: c.center_lng, tick: Date.now() });
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
          <CitySelector value={cityName} onChange={handleCityChange} />
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
      />

      <div className="coverage__body">
        <div className="coverage__map">
          {(summary.isLoading || gridCells.isLoading) && (
            <div className="map-loader">loading coverage…</div>
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
